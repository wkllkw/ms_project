package task

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test.com/project-api/testutil"
	"test.com/project-api/pkg/codecs"
)

// ============================================================
// 测试套件：覆盖 task 模块的核心数据库写入操作
// 运行方式：
//   cd /data/workspace/ms_project/backend/api
//   TEST_DB_USER=root TEST_DB_PASS=root TEST_DB_HOST=127.0.0.1 TEST_DB_PORT=13306 TEST_DB_NAME=ms_project_test \
//     go test ./api/task/ -v -count=1
//   go test ./api/task/ -v -run TestSave    # 只测创建
// ============================================================

func encryptID(id int64) string { return codecs.EncryptInt64(id) }

// insertTask 在测试库中插入一条任务并返回自增ID（MySQL 不支持 RETURNING）
func insertTask(db *gorm.DB, projectID, stageID int64, name string, deleted, done, like int) int64 {
	now := time.Now().UnixMilli()
	db.Exec(
		"INSERT INTO ms_task(project_code,name,stage_code,member_code,owner_code,create_time,sort,deleted,private,done,`like`) "+
			"VALUES(?,?,?,?,?,?,?,?,?,?,?)",
		projectID, name, stageID, 1, 1, now, 1, deleted, 0, done, like,
	)
	var id int64
	db.Raw("SELECT LAST_INSERT_ID()").Scan(&id)
	return id
}

// ============================================================
// 测试环境：Gin Router + 测试DB注入
// ============================================================

func setupRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 注入 memberId + 测试DB 到 context（让 authz 包也能拿到测试DB）
	r.Use(func(c *gin.Context) {
		c.Set("memberId", int64(1))
		c.Set("memberName", "测试用户")
		c.Set("test_db", db) // authz.getDB(c) 会取这个值
		c.Next()
	})

	h := &HandlerTask{dbOverride: db}
	taskGroup := r.Group("/project")
	{
		taskGroup.POST("/task/save", h.save)
		taskGroup.POST("/task/edit", h.edit)
		taskGroup.POST("/task/taskDone", h.taskDone)
		taskGroup.POST("/task/recycle", h.recycle)
		taskGroup.POST("/task/recovery", h.recovery)
		taskGroup.POST("/task/delete", h.del)
		taskGroup.POST("/task/like", h.like)
		taskGroup.POST("/task/star", h.star)
		taskGroup.POST("/task/setPrivate", h.setPrivate)
		taskGroup.POST("/task/createComment", h.createComment)
		taskGroup.POST("/task/saveTaskWorkTime", h.saveTaskWorkTime)
		taskGroup.POST("/task/assignTask", h.assignTask)
		taskGroup.POST("/task/sort", h.sort)
	}
	return r
}

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testutil.NewTestDB(t)
}

func seedProject(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	now := time.Now().UnixMilli()
	pID := time.Now().UnixNano()
	testutil.MustExecSQL(db, "INSERT INTO ms_project(id,name,description,create_time) VALUES(?,?,?,?)", pID, "测试项目", "用于自动化测试", now)
	testutil.MustExecSQL(db, "INSERT INTO ms_project_member(project_code,member_code,join_time,is_owner) VALUES(?,?,?,?)", pID, 1, now, 1)
	for i, name := range []string{"待处理", "进行中", "已完成"} {
		testutil.MustExecSQL(db, "INSERT INTO ms_task_stages(project_code,name,sort,create_time) VALUES(?,?,?,?)", pID, name, i+1, now)
	}
	return pID
}

func getStageID(t *testing.T, db *gorm.DB, projectID int64, stageName string) int64 {
	t.Helper()
	var id int64
	if err := db.Raw("SELECT id FROM ms_task_stages WHERE project_code=? AND name=? LIMIT 1", projectID, stageName).Scan(&id).Error; err != nil || id == 0 {
		t.Fatalf("找不到阶段 %s: %v", stageName, err)
	}
	return id
}

// ============================================================
// Test 1: 创建任务 save
// ============================================================

func TestSaveTask(t *testing.T) {
	db := openDB(t)
	if db == nil { return }
	r := setupRouter(db)
	projectID := seedProject(t, db)
	stageID := getStageID(t, db, projectID, "待处理")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("name", "测试任务1")
	_ = writer.WriteField("description", "这是描述")
	_ = writer.WriteField("project_code", encryptID(projectID))
	_ = writer.WriteField("stage_code", encryptID(stageID))
	_ = writer.WriteField("begin_time", time.Now().Format("2006-01-02 15:04"))
	_ = writer.WriteField("end_time", time.Now().Add(24*time.Hour).Format("2006-01-02 15:04"))
	writer.Close()

	req := httptest.NewRequest("POST", "/project/task/save", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 200 {
		t.Fatalf("创建失败: %s", w.Body.String())
	}
	if resp.Data.Code == "" {
		t.Fatal("创建成功但没有返回 code")
	}
	t.Logf("✅ 创建任务: code=%s, name=%s", resp.Data.Code, resp.Data.Name)
}

// ============================================================
// Test 2: 任务完成 → done + done_time + 事件日志
// ============================================================

func TestTaskDone(t *testing.T) {
	db := openDB(t)
	if db == nil { return }
	r := setupRouter(db)
	projectID := seedProject(t, db)
	stageID := getStageID(t, db, projectID, "待处理")

	taskID := insertTask(db, projectID, stageID, "待完成任务", 0, 0, 0)
	encID := encryptID(taskID)

	// 走 handler 完成任务
	form := url.Values{}
	form.Set("taskCode", encID)
	form.Set("done", "1")

	req := httptest.NewRequest("POST", "/project/task/taskDone", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	t.Logf("Handler response: %s", w.Body.String())

	var resp struct{ Code int }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 200 {
		t.Fatalf("完成失败: %s", w.Body.String())
	}

	// 分两次查询，避免链式 Scan 数据丢失
	var tDone int8
	var tDoneTime int64
	db.Raw("SELECT done FROM ms_task WHERE id=?", taskID).Scan(&tDone)
	db.Raw("SELECT done_time FROM ms_task WHERE id=?", taskID).Scan(&tDoneTime)
	if tDone != 1 {
		t.Fatalf("done 应为 1, 实际 %d", tDone)
	}
	if tDoneTime == 0 {
		t.Fatal("done_time 未被设置")
	}

	var logCount int64
	db.Raw("SELECT count(1) FROM ms_project_event WHERE task_id=? AND event_type='task:done'", taskID).Scan(&logCount)
	t.Logf("✅ 任务完成: done=%d, done_time=%d, 事件日志=%d", tDone, tDoneTime, logCount)
}

// ============================================================
// Test 3: 软删除 + 恢复
// ============================================================

func TestRecycleAndRecovery(t *testing.T) {
	db := openDB(t)
	if db == nil { return }
	r := setupRouter(db)
	projectID := seedProject(t, db)
	stageID := getStageID(t, db, projectID, "待处理")

	taskID := insertTask(db, projectID, stageID, "待回收任务", 0, 0, 0)
	encID := encryptID(taskID)
	form := url.Values{"taskCode": {encID}}

	// 软删除
	req := httptest.NewRequest("POST", "/project/task/recycle", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct{ Code int }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 200 {
		t.Fatalf("软删除失败: %s", w.Body.String())
	}
	var deleted int8
	db.Raw("SELECT deleted FROM ms_task WHERE id=?", taskID).Scan(&deleted)
	if deleted != 1 {
		t.Fatalf("deleted 应为 1, 实际 %d", deleted)
	}
	t.Log("✅ 软删除成功")

	// 恢复
	req2 := httptest.NewRequest("POST", "/project/task/recovery", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	json.Unmarshal(w2.Body.Bytes(), &resp)
	if resp.Code != 200 {
		t.Fatalf("恢复失败: %s", w2.Body.String())
	}
	db.Raw("SELECT deleted FROM ms_task WHERE id=?", taskID).Scan(&deleted)
	if deleted != 0 {
		t.Fatalf("deleted 应恢复为 0, 实际 %d", deleted)
	}
	t.Log("✅ 恢复成功")
}

// ============================================================
// Test 4: 物理删除
// ============================================================

func TestDeleteTask(t *testing.T) {
	db := openDB(t)
	if db == nil { return }
	r := setupRouter(db)
	projectID := seedProject(t, db)
	stageID := getStageID(t, db, projectID, "待处理")

	taskID := insertTask(db, projectID, stageID, "待删除任务", 1, 0, 0) // deleted=1

	form := url.Values{"taskCode": {encryptID(taskID)}}
	req := httptest.NewRequest("POST", "/project/task/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct{ Code int }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 200 {
		t.Fatalf("物理删除失败: %s", w.Body.String())
	}
	var count int64
	db.Raw("SELECT count(1) FROM ms_task WHERE id=?", taskID).Scan(&count)
	if count != 0 {
		t.Fatal("任务应被物理删除")
	}
	t.Logf("✅ 物理删除成功, 行数=%d", count)
}

// ============================================================
// Test 5: 点赞原子自增/自减
// ============================================================

func TestLikeTask(t *testing.T) {
	db := openDB(t)
	if db == nil { return }
	r := setupRouter(db)
	projectID := seedProject(t, db)
	stageID := getStageID(t, db, projectID, "待处理")

	taskID := insertTask(db, projectID, stageID, "点赞测试", 0, 0, 5)

	// 点赞
	form := url.Values{"taskCode": {encryptID(taskID)}, "like": {"1"}}
	req := httptest.NewRequest("POST", "/project/task/like", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct{ Code int }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 200 {
		t.Fatalf("点赞失败: %s", w.Body.String())
	}
	var likeCount int
	db.Raw("SELECT `like` FROM ms_task WHERE id=?", taskID).Scan(&likeCount)
	if likeCount != 6 {
		t.Fatalf("点赞后应为 6, 实际 %d", likeCount)
	}

	// 取消
	form2 := url.Values{"taskCode": {encryptID(taskID)}, "like": {"0"}}
	req2 := httptest.NewRequest("POST", "/project/task/like", strings.NewReader(form2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	json.Unmarshal(w2.Body.Bytes(), &resp)
	if resp.Code != 200 {
		t.Fatalf("取消点赞失败: %s", w2.Body.String())
	}
	db.Raw("SELECT `like` FROM ms_task WHERE id=?", taskID).Scan(&likeCount)
	if likeCount != 5 {
		t.Fatalf("取消后应为 5, 实际 %d", likeCount)
	}
	t.Logf("✅ 点赞原子性: 5→6→5")
}

// ============================================================
// Test 6: 创建评论 → 评论表 + 事件日志
// ============================================================

func TestCreateComment(t *testing.T) {
	db := openDB(t)
	if db == nil { return }
	r := setupRouter(db)
	projectID := seedProject(t, db)
	stageID := getStageID(t, db, projectID, "待处理")

	taskID := insertTask(db, projectID, stageID, "评论测试", 0, 0, 0)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("taskCode", encryptID(taskID)) // ← camelCase
	_ = writer.WriteField("comment", "这是一条测试评论")
	writer.Close()

	req := httptest.NewRequest("POST", "/project/task/createComment", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct{ Code int }
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 200 {
		t.Fatalf("评论失败: %s", w.Body.String())
	}
	var comCount, eventCount int64
	db.Raw("SELECT count(1) FROM ms_task_comment WHERE task_id=?", taskID).Scan(&comCount)
	db.Raw("SELECT count(1) FROM ms_project_event WHERE task_id=? AND event_type='task:comment'", taskID).Scan(&eventCount)
	if comCount == 0 {
		t.Fatal("评论未写入 ms_task_comment")
	}
	if eventCount == 0 {
		t.Fatal("事件日志未写入 ms_project_event")
	}
	t.Logf("✅ 评论写入: 评论表=%d, 事件日志=%d", comCount, eventCount)
}

// ============================================================
// Test 7: 编辑任务
// ============================================================

func TestEditTask(t *testing.T) {
	db := openDB(t)
	if db == nil { return }
	r := setupRouter(db)
	projectID := seedProject(t, db)
	stageID := getStageID(t, db, projectID, "待处理")

	taskID := insertTask(db, projectID, stageID, "原始名称", 0, 0, 0)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("taskCode", encryptID(taskID)) // ← camelCase
	_ = writer.WriteField("name", "修改后的名称")
	_ = writer.WriteField("description", "修改后的描述")
	writer.Close()

	req := httptest.NewRequest("POST", "/project/task/edit", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Name string `json:"name"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 200 {
		t.Fatalf("编辑失败: %s", w.Body.String())
	}
	var dbName string
	db.Raw("SELECT name FROM ms_task WHERE id=?", taskID).Scan(&dbName)
	if dbName != "修改后的名称" {
		t.Fatalf("DB 名称为: %s", dbName)
	}
	t.Logf("✅ 编辑任务: %s", dbName)
}

// ============================================================
// Test 8: 缺少必填字段
// ============================================================

func TestSaveTaskMissingFields(t *testing.T) {
	db := openDB(t)
	if db == nil { return }
	r := setupRouter(db)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("project_code", encryptID(1))
	_ = writer.WriteField("stage_code", encryptID(1))
	// 不传 name
	writer.Close()

	req := httptest.NewRequest("POST", "/project/task/save", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code == 200 {
		t.Fatal("空名称应返回错误")
	}
	t.Logf("✅ 参数校验: code=%d, msg=%s", resp.Code, resp.Msg)
}

// ============================================================
// Test 9: 并发点赞 → 原子性验证
// ============================================================

func TestLikeConcurrency(t *testing.T) {
	db := openDB(t)
	if db == nil { return }
	r := setupRouter(db)
	projectID := seedProject(t, db)
	stageID := getStageID(t, db, projectID, "待处理")

	taskID := insertTask(db, projectID, stageID, "并发点赞", 0, 0, 0)
	encID := encryptID(taskID)
	const n = 10
	done := make(chan bool, n)

	for i := 0; i < n; i++ {
		go func() {
			form := url.Values{"taskCode": {encID}, "like": {"1"}}
			req := httptest.NewRequest("POST", "/project/task/like", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			done <- true
		}()
	}
	for i := 0; i < n; i++ {
		<-done
	}

	var likeCount int
	db.Raw("SELECT `like` FROM ms_task WHERE id=?", taskID).Scan(&likeCount)
	if likeCount != n {
		t.Fatalf("并发点赞: 期望=%d, 实际=%d", n, likeCount)
	}
	t.Logf("✅ 并发 %d 次点赞原子性正常: like=%d", n, likeCount)
}

// ============================================================
// 清理
// ============================================================

func TestCleanup(t *testing.T) {
	db := openDB(t)
	if db == nil { return }
	db.Exec("DELETE FROM ms_task WHERE name LIKE '%测试%' OR name LIKE '%待%' OR name LIKE '%原始%' OR name LIKE '%点赞%' OR name LIKE '%评论%'")
	db.Exec("DELETE FROM ms_task_comment WHERE comment LIKE '%测试%'")
	db.Exec("DELETE FROM ms_project_event WHERE event_type IN ('task:comment','task:done')")
	db.Exec("DELETE FROM ms_project WHERE name='测试项目'")
	db.Exec("DELETE FROM ms_task_stages WHERE project_code NOT IN (SELECT id FROM ms_project)")
	db.Exec("DELETE FROM ms_project_member WHERE project_code NOT IN (SELECT id FROM ms_project)")
	t.Log("✅ 测试数据清理完成")
}
