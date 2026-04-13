package task

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test.com/project-api/internal/authz"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerTask struct {
}

func New() *HandlerTask {
	return &HandlerTask{}
}

type taskRow struct {
	Id           int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode  int64
	Name         string
	Description  string
	Status       int8
	Priority     int8
	BeginTime    int64
	EndTime      int64
	CreateTime   int64
	MemberCode   int64
	OwnerCode    int64
	AssignTo     int64
	StageCode    int64
	ParentTaskId int64
	Sort         int
	Deleted      int8
	Private      int8
	Done         int8
	LikeCount    int `gorm:"column:like"`
	Star         int
}

func (*taskRow) TableName() string { return "ms_task" }

type memberRow struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	Name   string
	Avatar string
}

func (*memberRow) TableName() string { return "ms_member" }

type taskCommentRow struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	TaskId     int64
	MemberId   int64
	Comment    string
	CreateTime int64
}

func (*taskCommentRow) TableName() string { return "ms_task_comment" }

type taskWorkTimeRow struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	TaskId     int64
	MemberId   int64
	WorkTime   int64
	Remark     string
	CreateTime int64
}

func (*taskWorkTimeRow) TableName() string { return "ms_task_work_time" }

func taskToResponse(db *gorms.GormConn, c *gin.Context, t taskRow) gin.H {
	var projectInfo any
	if t.ProjectCode != 0 {
		projectInfo = gin.H{
			"code": codecs.EncryptInt64(t.ProjectCode),
			"name": "",
		}
		var pr struct {
			Id   int64
			Name string
		}
		_ = db.Session(c.Request.Context()).Table("ms_project").Select("id,name").Where("id=?", t.ProjectCode).First(&pr).Error
		if pr.Id != 0 {
			projectInfo = gin.H{
				"code": codecs.EncryptInt64(pr.Id),
				"name": pr.Name,
			}
		}
	}
	execId := t.AssignTo
	if execId == 0 {
		execId = t.OwnerCode
	}
	var executor any
	if execId != 0 {
		var m memberRow
		_ = db.Session(c.Request.Context()).Where("id=?", execId).First(&m).Error
		if m.Id != 0 {
			executor = gin.H{
				"code":   codecs.EncryptInt64(m.Id),
				"name":   m.Name,
				"avatar": m.Avatar,
			}
		}
	}
	return gin.H{
		"id":           t.Id,
		"id_num":       t.Id,
		"code":         codecs.EncryptInt64(t.Id),
		"name":         t.Name,
		"description":  t.Description,
		"done":         t.Done == 1,
		"canRead":      true,
		"status":       t.Status,
		"statusText":   "",
		"pri":          t.Priority,
		"begin_time":   t.BeginTime,
		"end_time":     t.EndTime,
		"executor":     executor,
		"childCount":   []int{0, 0},
		"tags":         []any{},
		"hasUnDone":    false,
		"hasSource":    false,
		"hasComment":   false,
		"like":         t.LikeCount,
		"star":         t.Star,
		"project_code": codecs.EncryptInt64(t.ProjectCode),
		"stage_code":   codecs.EncryptInt64(t.StageCode),
		"projectInfo":  projectInfo,
	}
}

func (h *HandlerTask) list(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	projectCode := c.PostForm("projectCode")
	deleted := c.PostForm("deleted")
	db := gorms.New()
	session := db.Session(c.Request.Context())
	query := session.Model(&taskRow{})
	if projectCode != "" {
		pid, err := codecs.DecryptInt64(projectCode)
		if err == nil {
			query = query.Where("project_code=?", pid)
		}
	}
	if deleted == "1" {
		query = query.Where("deleted=1")
	} else {
		query = query.Where("deleted=0")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "查询失败"))
		return
	}
	var rows []taskRow
	if err := query.Order("id desc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&rows).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "查询失败"))
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, t := range rows {
		out = append(out, taskToResponse(db, c, t))
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (h *HandlerTask) selfList(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	memberId := c.GetInt64("memberId")
	if memberCode := c.PostForm("memberCode"); memberCode != "" {
		if id, err := codecs.DecryptInt64(memberCode); err == nil && id != 0 {
			memberId = id
		}
	}
	db := gorms.New()
	session := db.Session(c.Request.Context())
	query := session.Model(&taskRow{}).Where("deleted=0").Where("assign_to=? or owner_code=? or member_code=?", memberId, memberId, memberId)
	if typeStr := c.PostForm("type"); typeStr != "" {
		if t, err := strconv.ParseInt(typeStr, 10, 64); err == nil && (t == 0 || t == 1) {
			query = query.Where("done=?", t)
		}
	}
	var total int64
	_ = query.Count(&total).Error
	var rows []taskRow
	_ = query.Order("id desc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, t := range rows {
		out = append(out, taskToResponse(db, c, t))
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (h *HandlerTask) save(c *gin.Context) {
	result := &common.Result{}
	name := c.PostForm("name")
	projectCode := c.PostForm("project_code")
	if projectCode == "" {
		projectCode = c.PostForm("projectCode")
	}
	stageCode := c.PostForm("stage_code")
	if stageCode == "" {
		stageCode = c.PostForm("stageCode")
	}
	assignToCode := c.PostForm("assign_to")
	if name == "" || projectCode == "" || stageCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "参数不完整"))
		return
	}
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	sid, err := codecs.DecryptInt64(stageCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "stageCode无效"))
		return
	}
	var assignTo int64
	if assignToCode != "" {
		assignTo, _ = codecs.DecryptInt64(assignToCode)
	}
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB()
	var maxSort int
	_ = db.Model(&taskRow{}).Where("stage_code=? and deleted=0", sid).Select("coalesce(max(sort),0)").Scan(&maxSort).Error
	row := &taskRow{
		ProjectCode: pid,
		Name:        name,
		Description: c.PostForm("description"),
		Priority:    0,
		CreateTime:  time.Now().UnixMilli(),
		MemberCode:  memberId,
		OwnerCode:   memberId,
		AssignTo:    assignTo,
		StageCode:   sid,
		Sort:        maxSort + 1,
		Deleted:     0,
		Private:     0,
		Done:        0,
	}
	if err := db.Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(taskToResponse(gorms.New(), c, *row)))
}

func (h *HandlerTask) edit(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	if taskCode == "" {
		taskCode = c.PostForm("code")
	}
	if taskCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode必填"))
		return
	}

	// 权限校验：只有项目成员才能编辑任务
	id, projectId, ok := authz.CanOperateTask(c, taskCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}

	updates := map[string]any{}
	if v := c.PostForm("name"); v != "" {
		updates["name"] = v
	}
	if v := c.PostForm("description"); v != "" {
		updates["description"] = v
	}
	db := gorms.GetDB()
	if err := db.Model(&taskRow{}).Where("id=?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"code": codecs.EncryptInt64(id), "projectCode": codecs.EncryptInt64(projectId)}))
}

func (h *HandlerTask) read(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	if taskCode == "" {
		taskCode = c.PostForm("code")
	}
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	db := gorms.New()
	var row taskRow
	if err := db.Session(c.Request.Context()).Where("id=?", id).First(&row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	var comments []taskCommentRow
	_ = db.Session(c.Request.Context()).Where("task_id=?", row.Id).Order("id desc").Find(&comments).Error
	commentOut := make([]gin.H, 0, len(comments))
	for _, cm := range comments {
		commentOut = append(commentOut, gin.H{
			"id":          cm.Id,
			"task_code":   taskCode,
			"comment":     cm.Comment,
			"create_time": cm.CreateTime,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"task":     taskToResponse(db, c, row),
		"comments": commentOut,
	}))
}

func (h *HandlerTask) taskDone(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	doneStr := c.PostForm("done")
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	done := int8(0)
	if doneStr == "1" || doneStr == "true" {
		done = 1
	}
	db := gorms.GetDB()
	if err := db.Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"done": done}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) recycle(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")

	// 权限校验：只有项目成员才能回收任务
	id, _, ok := authz.CanOperateTask(c, taskCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}

	if err := gorms.GetDB().Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"deleted": 1}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "回收失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) recovery(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")

	// 权限校验：只有项目成员才能恢复任务
	id, _, ok := authz.CanOperateTask(c, taskCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}

	if err := gorms.GetDB().Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"deleted": 0}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "恢复失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) del(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")

	// 权限校验：只有项目成员才能删除任务
	id, _, ok := authz.CanOperateTask(c, taskCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}

	if err := gorms.GetDB().Where("id=?", id).Delete(&taskRow{}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "删除失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) recycleBatch(c *gin.Context) {
	result := &common.Result{}
	db := gorms.GetDB().WithContext(c.Request.Context())
	raw := c.PostForm("taskCodes")
	var codes []string
	_ = json.Unmarshal([]byte(raw), &codes)
	ids := make([]int64, 0, len(codes))
	for _, code := range codes {
		id, _, ok := authz.CanOperateTask(c, code)
		if ok {
			ids = append(ids, id)
		}
	}
	if len(ids) > 0 {
		_ = db.Model(&taskRow{}).Where("id in ?", ids).Updates(map[string]any{"deleted": 1}).Error
		c.JSON(http.StatusOK, result.Success([]int{}))
		return
	}
	stageCode := c.PostForm("stageCode")
	if stageCode != "" {
		stageId, err := codecs.DecryptInt64(stageCode)
		if err == nil && stageId != 0 {
			var stage struct {
				ProjectCode int64 `gorm:"column:project_code"`
			}
			_ = db.Table("ms_task_stages").Select("project_code").Where("id=?", stageId).Scan(&stage).Error
			if stage.ProjectCode != 0 && authz.IsProjectMember(db, c.GetInt64("memberId"), stage.ProjectCode) {
				_ = db.Model(&taskRow{}).Where("stage_code=? and deleted=0", stageId).Updates(map[string]any{"deleted": 1}).Error
			}
		}
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) batchAssignTask(c *gin.Context) {
	result := &common.Result{}
	raw := c.PostForm("taskCodes")
	executorCode := c.PostForm("executorCode")
	var codes []string
	_ = json.Unmarshal([]byte(raw), &codes)
	ids := make([]int64, 0, len(codes))
	for _, code := range codes {
		id, err := codecs.DecryptInt64(code)
		if err == nil {
			ids = append(ids, id)
		}
	}
	execId, err := codecs.DecryptInt64(executorCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "executorCode无效"))
		return
	}
	if len(ids) > 0 {
		_ = gorms.GetDB().Model(&taskRow{}).Where("id in ?", ids).Updates(map[string]any{"assign_to": execId}).Error
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) setPrivate(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	privateStr := c.PostForm("private")
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	private := int8(0)
	if privateStr == "1" || privateStr == "true" {
		private = 1
	}
	_ = gorms.GetDB().Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"private": private}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) like(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	likeStr := c.PostForm("like")
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	like := 0
	if likeStr == "1" || likeStr == "true" {
		like = 1
	}
	db := gorms.GetDB()
	if like == 1 {
		_ = db.Model(&taskRow{}).Where("id=?", id).UpdateColumn("like", gorm.Expr("coalesce(`like`,0)+1")).Error
	} else {
		_ = db.Model(&taskRow{}).Where("id=?", id).UpdateColumn("like", gorm.Expr("greatest(coalesce(`like`,0)-1,0)")).Error
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) star(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	starStr := c.PostForm("star")
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	star := 0
	if starStr == "1" || starStr == "true" {
		star = 1
	}
	_ = gorms.GetDB().Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"star": star}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) createComment(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	comment := c.PostForm("comment")
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	memberId := c.GetInt64("memberId")
	row := &taskCommentRow{TaskId: id, MemberId: memberId, Comment: comment, CreateTime: time.Now().UnixMilli()}
	_ = gorms.GetDB().Create(row).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) sort(c *gin.Context) {
	result := &common.Result{}
	preTaskCode := c.PostForm("preTaskCode")
	nextTaskCode := c.PostForm("nextTaskCode")
	toStageCode := c.PostForm("toStageCode")
	if preTaskCode == "" || toStageCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "参数不完整"))
		return
	}
	movedId, err := codecs.DecryptInt64(preTaskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "preTaskCode无效"))
		return
	}
	stageId, err := codecs.DecryptInt64(toStageCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "toStageCode无效"))
		return
	}
	var nextId int64
	if nextTaskCode != "" {
		nextId, _ = codecs.DecryptInt64(nextTaskCode)
	}
	db := gorms.GetDB()
	var list []taskRow
	if err := db.Where("stage_code=? and deleted=0", stageId).Order("sort asc, id asc").Find(&list).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "排序失败"))
		return
	}
	var moved taskRow
	if err := db.Where("id=?", movedId).First(&moved).Error; err == nil {
		moved.StageCode = stageId
	}
	newList := make([]taskRow, 0, len(list)+1)
	for _, t := range list {
		if t.Id == movedId {
			continue
		}
		newList = append(newList, t)
	}
	inserted := false
	if nextId != 0 {
		for i, t := range newList {
			if t.Id == nextId {
				left := append([]taskRow{}, newList[:i]...)
				right := append([]taskRow{}, newList[i:]...)
				newList = append(left, moved)
				newList = append(newList, right...)
				inserted = true
				break
			}
		}
	}
	if !inserted {
		newList = append(newList, moved)
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&taskRow{}).Where("id=?", movedId).Updates(map[string]any{"stage_code": stageId}).Error; err != nil {
			return err
		}
		for i, t := range newList {
			if err := tx.Model(&taskRow{}).Where("id=?", t.Id).Update("sort", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "排序失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// taskTagRow 任务标签表
type taskTagRow struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64
	Name        string
	Color       string
	CreateTime  int64
	Deleted     int8
}

func (*taskTagRow) TableName() string { return "ms_task_tag" }

// taskTagRelRow 任务标签关联表
type taskTagRelRow struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	TaskId int64
	TagId  int64
}

func (*taskTagRelRow) TableName() string { return "ms_task_tag_rel" }

// taskLogRow 任务操作日志（复用 ms_project_event 表）
type taskLogRow struct {
	Id           int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode  int64
	MemberCode   int64
	EventType    string
	EventContent string
	CreateTime   int64
}

func (*taskLogRow) TableName() string { return "ms_project_event" }

// assignTask 单个任务指派
func (h *HandlerTask) assignTask(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	executorCode := c.PostForm("executorCode")
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	execId, err := codecs.DecryptInt64(executorCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "executorCode无效"))
		return
	}
	if err := gorms.GetDB().Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"assign_to": execId}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "指派失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// taskSources 任务来源列表（按项目统计各阶段任务数）
func (h *HandlerTask) taskSources(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	projectCode := c.PostForm("projectCode")
	var pid int64
	var err error

	// 优先使用 taskCode 查询项目
	if taskCode != "" {
		taskId, decErr := codecs.DecryptInt64(taskCode)
		if decErr == nil && taskId != 0 {
			var task taskRow
			_ = gorms.GetDB().Select("project_code").Where("id=?", taskId).First(&task).Error
			pid = task.ProjectCode
		}
	} else if projectCode != "" {
		pid, err = codecs.DecryptInt64(projectCode)
		if err != nil || pid == 0 {
			c.JSON(http.StatusOK, result.Success([]any{}))
			return
		}
	}

	if pid == 0 {
		c.JSON(http.StatusOK, result.Success([]any{}))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var rows []struct {
		StageCode int64  `gorm:"column:stage_code"`
		StageName string `gorm:"column:stage_name"`
		Count     int64  `gorm:"column:count"`
	}
	_ = db.Table("ms_task t").
		Joins("left join ms_task_stages s on s.id=t.stage_code").
		Select("t.stage_code, coalesce(s.name,'') as stage_name, count(*) as count").
		Where("t.project_code=? and t.deleted=0", pid).
		Group("t.stage_code").
		Scan(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"stage_code": codecs.EncryptInt64(r.StageCode),
			"stage_name": r.StageName,
			"count":      r.Count,
		})
	}
	c.JSON(http.StatusOK, result.Success(out))
}

// getListByTaskTag 按标签查询任务列表
func (h *HandlerTask) getListByTaskTag(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	projectCode := c.PostForm("projectCode")
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil || pid == 0 {
		c.JSON(http.StatusOK, result.Success(gin.H{"list": []any{}, "total": 0}))
		return
	}
	tagCode := c.PostForm("tagCode")
	tagId, err := codecs.DecryptInt64(tagCode)
	if err != nil || tagId == 0 {
		c.JSON(http.StatusOK, result.Success(gin.H{"list": []any{}, "total": 0}))
		return
	}
	db := gorms.New()
	session := db.Session(c.Request.Context())
	query := session.Table("ms_task t").
		Joins("join ms_task_tag_rel r on r.task_id=t.id").
		Where("t.project_code=? and r.tag_id=? and t.deleted=0", pid, tagId)
	var total int64
	_ = query.Count(&total).Error
	var rows []taskRow
	_ = query.Select("t.*").Order("t.id desc").
		Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).
		Scan(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, t := range rows {
		out = append(out, taskToResponse(db, c, t))
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

// taskToTags 查询任务已关联的标签
func (h *HandlerTask) taskToTags(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	var tags []taskTagRow
	_ = db.Table("ms_task_tag tg").
		Joins("join ms_task_tag_rel r on r.tag_id=tg.id").
		Where("r.task_id=? and tg.deleted=0", id).
		Select("tg.*").
		Find(&tags).Error
	out := make([]gin.H, 0, len(tags))
	for _, t := range tags {
		out = append(out, gin.H{
			"code":  codecs.EncryptInt64(t.Id),
			"name":  t.Name,
			"color": t.Color,
		})
	}
	c.JSON(http.StatusOK, result.Success(out))
}

// setTag 为任务设置/取消标签
func (h *HandlerTask) setTag(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	tagCode := c.PostForm("tagCode")
	action := c.PostForm("action") // "add" or "remove"
	taskId, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	tagId, err := codecs.DecryptInt64(tagCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "tagCode无效"))
		return
	}
	db := gorms.GetDB()
	if action == "remove" {
		_ = db.Where("task_id=? and tag_id=?", taskId, tagId).Delete(&taskTagRelRow{}).Error
	} else {
		var cnt int64
		_ = db.Model(&taskTagRelRow{}).Where("task_id=? and tag_id=?", taskId, tagId).Count(&cnt).Error
		if cnt == 0 {
			_ = db.Create(&taskTagRelRow{TaskId: taskId, TagId: tagId}).Error
		}
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// dateTotalForProject 统计项目中各日期的任务数量
func (h *HandlerTask) dateTotalForProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil || pid == 0 {
		c.JSON(http.StatusOK, result.Success([]any{}))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	var rows []struct {
		Date  string `gorm:"column:date"`
		Total int64  `gorm:"column:total"`
		Done  int64  `gorm:"column:done"`
	}
	_ = db.Table("ms_task").
		Select("DATE_FORMAT(FROM_UNIXTIME(create_time/1000),'%Y-%m-%d') as date, count(*) as total, sum(case when done=1 then 1 else 0 end) as done").
		Where("project_code=? and deleted=0 and create_time>0", pid).
		Group("date").
		Order("date asc").
		Scan(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"date":  r.Date,
			"total": r.Total,
			"done":  r.Done,
		})
	}
	c.JSON(http.StatusOK, result.Success(out))
}

// taskLog 查询任务相关的操作日志
func (h *HandlerTask) taskLog(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	taskCode := c.PostForm("taskCode")
	projectCode := c.PostForm("projectCode")
	db := gorms.GetDB().WithContext(c.Request.Context())
	query := db.Model(&taskLogRow{})
	if taskCode != "" {
		if id, err := codecs.DecryptInt64(taskCode); err == nil && id != 0 {
			query = query.Where("event_content like ?", "%taskId:"+strconv.FormatInt(id, 10)+"%")
		}
	} else if projectCode != "" {
		if pid, err := codecs.DecryptInt64(projectCode); err == nil && pid != 0 {
			query = query.Where("project_code=?", pid)
		}
	}
	var total int64
	_ = query.Count(&total).Error
	var rows []taskLogRow
	_ = query.Order("id desc").
		Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).
		Find(&rows).Error

	// 查询所有涉及的成员信息
	memberIds := make([]int64, 0)
	for _, r := range rows {
		memberIds = append(memberIds, r.MemberCode)
	}
	memberMap := make(map[int64]memberRow)
	if len(memberIds) > 0 {
		var members []memberRow
		_ = db.Table("ms_member").Where("id IN ?", memberIds).Find(&members).Error
		for _, m := range members {
			memberMap[m.Id] = m
		}
	}

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		// 获取成员信息
		member, ok := memberMap[r.MemberCode]
		memberInfo := gin.H{"name": "未知用户", "avatar": ""}
		if ok {
			memberInfo = gin.H{
				"name":   member.Name,
				"avatar": member.Avatar,
			}
		}

		// 解析事件类型
		remark, icon, isComment := parseEventType(r.EventType, r.EventContent)

		out = append(out, gin.H{
			"code":        codecs.EncryptInt64(r.Id),
			"member":      memberInfo,
			"remark":      remark,
			"content":     r.EventContent,
			"icon":        icon,
			"is_comment":  isComment,
			"create_time": r.CreateTime,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

// parseEventType 解析事件类型，返回 remark, icon, isComment
func parseEventType(eventType, eventContent string) (string, string, bool) {
	switch eventType {
	case "create", "task:create":
		return "创建了任务", "plus", false
	case "done", "task:done":
		return "完成了任务", "check", false
	case "edit", "task:edit":
		return "编辑了任务", "edit", false
	case "comment", "task:comment":
		return "发表了评论", "message", true
	case "assign", "task:assign":
		return "指派了任务", "user", false
	case "delete", "task:delete":
		return "删除了任务", "delete", false
	case "recovery", "task:recovery":
		return "恢复了任务", "undo", false
	case "move", "task:move":
		return "移动了任务", "swap", false
	case "priority", "task:priority":
		return "修改了优先级", "flag", false
	case "tag", "task:tag":
		return "修改了标签", "tag", false
	case "file", "task:file":
		return "上传了附件", "paper-clip", false
	default:
		// 默认根据内容推断
		if strings.Contains(strings.ToLower(eventType), "comment") {
			return "发表了评论", "message", true
		}
		return "更新了任务", "info-circle", false
	}
}

func (h *HandlerTask) taskWorkTimeList(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	db := gorms.GetDB()
	var rows []taskWorkTimeRow
	_ = db.Where("task_id=?", id).Order("id desc").Find(&rows).Error

	// 查询成员信息
	memberIds := make([]int64, 0)
	for _, r := range rows {
		memberIds = append(memberIds, r.MemberId)
	}
	memberMap := make(map[int64]memberRow)
	if len(memberIds) > 0 {
		var members []memberRow
		_ = db.Table("ms_member").Where("id IN ?", memberIds).Find(&members).Error
		for _, m := range members {
			memberMap[m.Id] = m
		}
	}

	// 转换为前端期待的格式
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		memberInfo := gin.H{"name": "未知用户", "avatar": ""}
		if m, ok := memberMap[r.MemberId]; ok {
			memberInfo = gin.H{"name": m.Name, "avatar": m.Avatar}
		}
		// workTime 字段存的是毫秒，转换为小时
		num := float64(r.WorkTime) / 3600000.0
		out = append(out, gin.H{
			"code":       codecs.EncryptInt64(r.Id),
			"member":     memberInfo,
			"begin_time": r.CreateTime,
			"num":        num,
			"content":    r.Remark,
		})
	}
	c.JSON(http.StatusOK, result.Success(out))
}

func (h *HandlerTask) saveTaskWorkTime(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	memberId := c.GetInt64("memberId")
	workTimeStr := c.PostForm("workTime")
	var workTime int64
	_ = json.Unmarshal([]byte(workTimeStr), &workTime)
	row := &taskWorkTimeRow{
		TaskId:     id,
		MemberId:   memberId,
		WorkTime:   workTime,
		Remark:     c.PostForm("remark"),
		CreateTime: time.Now().UnixMilli(),
	}
	_ = gorms.GetDB().Create(row).Error
	c.JSON(http.StatusOK, result.Success(row))
}

// editTaskWorkTime 编辑工时记录
func (h *HandlerTask) editTaskWorkTime(c *gin.Context) {
	result := &common.Result{}
	idStr := c.PostForm("id")
	if idStr == "" {
		c.JSON(http.StatusOK, result.Fail(400, "id必填"))
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "id无效"))
		return
	}
	updates := map[string]any{}
	if v := c.PostForm("remark"); v != "" {
		updates["remark"] = v
	}
	if v := c.PostForm("workTime"); v != "" {
		var wt int64
		_ = json.Unmarshal([]byte(v), &wt)
		if wt > 0 {
			updates["work_time"] = wt
		}
	}
	if len(updates) == 0 {
		c.JSON(http.StatusOK, result.Success([]int{}))
		return
	}
	if err := gorms.GetDB().Model(&taskWorkTimeRow{}).Where("id=?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// delTaskWorkTime 删除工时记录
func (h *HandlerTask) delTaskWorkTime(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("code")
	if code == "" {
		c.JSON(http.StatusOK, result.Fail(400, "code必填"))
		return
	}
	id, err := codecs.DecryptInt64(code)
	if err != nil || id == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	if err := gorms.GetDB().Where("id=?", id).Delete(&taskWorkTimeRow{}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "删除失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// downloadTemplate 下载任务导入模板（返回CSV格式说明）
func (h *HandlerTask) downloadTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=task_template.csv")
	c.String(http.StatusOK, "任务名称,描述,优先级(0-3),开始时间(yyyy-MM-dd),截止时间(yyyy-MM-dd)\n")
}

// uploadFile 上传文件（保存到 ms_project_event 作为附件记录）
func (h *HandlerTask) uploadFile(c *gin.Context) {
	result := &common.Result{}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "请选择文件"))
		return
	}
	defer file.Close()
	projectCode := c.PostForm("projectCode")
	pid, _ := codecs.DecryptInt64(projectCode)
	memberId := c.GetInt64("memberId")
	// 记录上传事件到 ms_project_event
	row := &taskLogRow{
		ProjectCode:  pid,
		MemberCode:   memberId,
		EventType:    "upload_file",
		EventContent: "上传文件: " + header.Filename,
		CreateTime:   time.Now().UnixMilli(),
	}
	_ = gorms.GetDB().Create(row).Error
	c.JSON(http.StatusOK, result.Success(gin.H{
		"id":        row.Id,
		"file_name": header.Filename,
		"file_size": header.Size,
	}))
}
