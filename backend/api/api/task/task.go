package task

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test.com/project-api/internal/authz"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	ws "test.com/project-api/api/websocket"
	common "test.com/project-common"
)

type HandlerTask struct {
	dbOverride *gorm.DB // 测试时可注入外部DB，nil时使用h.db()
}

func New() *HandlerTask {
	return &HandlerTask{}
}

// NewWithDB 创建可注入数据库连接的 Handler（用于测试）
func NewWithDB(db *gorm.DB) *HandlerTask {
	return &HandlerTask{dbOverride: db}
}

// db 返回当前使用的数据库连接（原生 *gorm.DB）
func (h *HandlerTask) db() *gorm.DB {
	if h.dbOverride != nil {
		return h.dbOverride
	}
	return gorms.GetDB()
}

// dbConn 返回 GormConn 包装（用于 taskToResponse 等需要 GormConn 的场景）
func (h *HandlerTask) dbConn() *gorms.GormConn {
	return gorms.NewWithDB(h.db())
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
	VersionCode  int64
	Sort         int
	Deleted      int8
	Private      int8
	Done         int8
	DoneTime     int64 `gorm:"column:done_time;default:0"`
	LikeCount    int `gorm:"column:like"`
	Star         int
	WorkTime     int64 `gorm:"column:work_time;default:0"` // 预估工时（分钟）
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

func statusText(status int8) string {
	switch status {
	case 0:
		return "未开始"
	case 1:
		return "已完成"
	case 2:
		return "进行中"
	case 3:
		return "挂起"
	case 4:
		return "测试中"
	default:
		return "未开始"
	}
}

func taskExecuteInfo(status int8) gin.H {
	colorMap := map[int8]string{
		0: "muted",
		1: "complete",
		2: "primary",
		3: "danger",
		4: "warning",
	}
	color := "muted"
	if c, ok := colorMap[status]; ok {
		color = c
	}
	return gin.H{
		"color": color,
		"name":  statusText(status),
	}
}

func taskToResponse(db *gorms.GormConn, c *gin.Context, t taskRow) gin.H {
	var projectInfo any
	var projectName string
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
			projectName = pr.Name
		}
	}
	// Query stage name
	var stageName string
	if t.StageCode != 0 {
		var st struct {
			Name string
		}
		_ = db.Session(c.Request.Context()).Table("ms_task_stages").Select("name").Where("id=?", t.StageCode).First(&st).Error
		stageName = st.Name
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
	// Query tags for this task (nested structure for frontend compatibility)
	var tagRows []taskTagRow
	_ = db.Session(c.Request.Context()).Table("ms_task_tag tg").
		Joins("join ms_task_tag_rel r on r.tag_id=tg.id").
		Where("r.task_id=? and tg.deleted=0", t.Id).
		Select("tg.*").
		Find(&tagRows).Error
	tags := make([]gin.H, 0, len(tagRows))
	for _, tg := range tagRows {
		tags = append(tags, gin.H{
			"code": codecs.EncryptInt64(tg.Id),
			"tag": gin.H{
				"color": tg.Color,
				"name":  tg.Name,
			},
		})
	}
	// Priority text mapping
	priText := ""
	switch t.Priority {
	case 0:
		priText = "普通"
	case 1:
		priText = "紧急"
	case 2:
		priText = "非常紧急"
	default:
		priText = "普通"
	}
	// Count child tasks
	var childDoneCount int64
	var childTotalCount int64
	if t.ParentTaskId == 0 {
		_ = db.Session(c.Request.Context()).Model(&taskRow{}).Where("parent_task_id=? and deleted=0", t.Id).Count(&childTotalCount).Error
		_ = db.Session(c.Request.Context()).Model(&taskRow{}).Where("parent_task_id=? and deleted=0 and done=1", t.Id).Count(&childDoneCount).Error
	}
	childCount := []int{int(childTotalCount), int(childDoneCount)}
	// Check hasUnDone
	hasUnDone := childTotalCount > 0 && childDoneCount < childTotalCount
	// Check hasSource
	var sourceCount int64
	_ = db.Session(c.Request.Context()).Table("ms_source_link").Where("task_code=?", t.Id).Count(&sourceCount).Error
	hasSource := sourceCount > 0
	// Check hasComment
	var commentCount int64
	_ = db.Session(c.Request.Context()).Table("ms_task_comment").Where("task_id=?", t.Id).Count(&commentCount).Error
	hasComment := commentCount > 0
	// Sum work_time (分钟)
	var totalWorkTimeMinutes int64
	_ = db.Session(c.Request.Context()).Table("ms_task_work_time").Where("task_id=?", t.Id).Select("coalesce(sum(work_time),0)").Scan(&totalWorkTimeMinutes).Error
	// Parent task info
	var parentTask any
	var parentTasks []gin.H
	if t.ParentTaskId != 0 {
		var pt taskRow
		_ = db.Session(c.Request.Context()).Where("id=?", t.ParentTaskId).First(&pt).Error
		if pt.Id != 0 {
			parentTask = gin.H{
				"code": codecs.EncryptInt64(pt.Id),
				"name": pt.Name,
			}
			parentTasks = append(parentTasks, parentTask.(gin.H))
		}
	}
	return gin.H{
		"id":             t.Id,
		"id_num":         t.Id,
		"code":           codecs.EncryptInt64(t.Id),
		"name":           t.Name,
		"description":    t.Description,
		"done":           t.Done == 1,
		"canRead":        true,
		"status":            t.Status,
		"statusText":        statusText(t.Status),
		"execute_state":     t.Status,
		"task_execute_name": statusText(t.Status),
		"task_execute":      taskExecuteInfo(t.Status),
		"pri":            t.Priority,
		"priText":        priText,
		"begin_time":     t.BeginTime,
		"end_time":       t.EndTime,
		"executor":       executor,
		"childCount":     childCount,
		"tags":           tags,
		"hasUnDone":      hasUnDone,
		"hasSource":      hasSource,
		"hasComment":     hasComment,
		"like":           t.LikeCount,
		"star":           t.Star,
		"project_code":   codecs.EncryptInt64(t.ProjectCode),
		"stage_code":     codecs.EncryptInt64(t.StageCode),
		"projectInfo":    projectInfo,
		"projectName":    projectName,
		"stageName":      stageName,
		"deleted":        t.Deleted,
		"openBeginTime":  t.BeginTime > 0,
		"liked":          false,
		"stared":         t.Star > 0,
		"work_time":      float64(t.WorkTime) / 60.0,       // 预估工时（小时）
		"total_work_time": float64(totalWorkTimeMinutes) / 60.0, // 实际工时（小时）
		"sort":           t.Sort,
		"private":        t.Private,
		"parentTask":     parentTask,
		"parentTasks":    parentTasks,
	}
}

// @Summary 任务列表查询
// @Description 根据项目编码分页查询任务列表，支持按状态/阶段/优先级等条件筛选
// @Tags task
// @Accept x-www-form-urlencoded
// @Produce json
// @Param projectCode formData string true "项目加密ID"
// @Param deleted formData string false "是否已删除(0/1)"
// @Param stageCode formData string false "看板阶段编码"
// @Param status formData int false "任务状态"
// @Param keyword formData string false "关键词搜索"
// @Success 200 {object} common.Result "成功返回任务列表和分页信息"
// @Failure 400 {object} common.Result "参数错误"
// @Security ApiKeyAuth
// @Router /project/task [post]
func (h *HandlerTask) list(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	projectCode := c.PostForm("projectCode")
	deleted := c.PostForm("deleted")
	db := h.dbConn()
	session := db.Session(c.Request.Context())
	query := session.Model(&taskRow{})
	orgCode := orgCodeFromContext(c)
	if projectCode != "" {
		pid, err := codecs.DecryptInt64(projectCode)
		if err == nil {
			// 数据隔离：检查当前用户是否是项目成员
			memberId := c.GetInt64("memberId")
			if !authz.IsProjectMember(session, memberId, pid) {
				c.JSON(http.StatusOK, result.Success(gin.H{"list": []any{}, "total": 0}))
				return
			}
			// 组织过滤：验证项目属于当前组织
			if orgCode != 0 {
				var projOrg int64
				session.Table("ms_project").Where("id=?", pid).Select("organization_code").Scan(&projOrg)
				if projOrg != orgCode {
					c.JSON(http.StatusOK, result.Success(gin.H{"list": []any{}, "total": 0}))
					return
				}
			}
			query = query.Where("project_code=?", pid)
		}
	} else {
		// 没有指定项目时，只返回用户参与的、当前组织下的项目中的任务
		memberId := c.GetInt64("memberId")
		subQuery := session.Table("ms_project_member pm").Select("pm.project_code").
			Joins("join ms_project p on p.id=pm.project_code").
			Where("pm.member_code=?", memberId)
		if orgCode != 0 {
			subQuery = subQuery.Where("p.organization_code=?", orgCode)
		}
		query = query.Where("project_code IN (?)", subQuery)
	}
	// pcode: 父任务加密ID，用于查询子任务
	if pcode := c.PostForm("pcode"); pcode != "" {
		pid, err := codecs.DecryptInt64(pcode)
		if err == nil {
			query = query.Where("parent_task_id=?", pid)
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
	orgCode := orgCodeFromContext(c)
	db := h.dbConn()
	session := db.Session(c.Request.Context())
	query := session.Model(&taskRow{}).Where("deleted=0").Where("assign_to=? or owner_code=? or member_code=?", memberId, memberId, memberId)
	// 组织过滤：只显示当前组织下的任务
	if orgCode != 0 {
		subQuery := session.Table("ms_project").Select("id").Where("organization_code=? AND deleted=0", orgCode)
		query = query.Where("project_code IN (?)", subQuery)
	}
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

// @Summary 创建新任务
// @Description 在指定项目中创建新任务，支持设置优先级/截止时间/执行者等
// @Tags task
// @Accept x-www-form-urlencoded
// @Produce json
// @Param name formData string true "任务名称"
// @Param projectCode formData string true "项目加密ID"
// @Param description formData string false "任务描述"
// @Param priority formData int false "优先级(0-3)"
// @Param end_time formData int64 false "截止时间戳(毫秒)"
// @Param assign_to formData string false "执行者账号"
// @Param stage_code formData int64 false "看板阶段编码"
// @Success 200 {object} common.Result "返回新创建的任务ID"
// @Failure 400 {object} common.Result "参数错误"
// @Security ApiKeyAuth
// @Router /project/task/save [post]
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
	// 权限校验：只有项目成员才能创建任务
	memberId := c.GetInt64("memberId")
	db := h.db()
	if !authz.IsProjectMember(db, memberId, pid) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此项目"))
		return
	}
	var assignTo int64
	if assignToCode != "" {
		assignTo, _ = codecs.DecryptInt64(assignToCode)
	}
	var endTime int64
	if v := c.PostForm("end_time"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04", v, time.Local); err == nil {
			endTime = t.UnixMilli()
		}
	}
	var beginTime int64
	if v := c.PostForm("begin_time"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04", v, time.Local); err == nil {
			beginTime = t.UnixMilli()
		}
	}
	// pcode: 父任务加密ID，用于创建子任务
	var parentTaskId int64
	if v := c.PostForm("pcode"); v != "" {
		parentTaskId, _ = codecs.DecryptInt64(v)
	}
	var maxSort int
	_ = db.Model(&taskRow{}).Where("stage_code=? and deleted=0", sid).Select("coalesce(max(sort),0)").Scan(&maxSort).Error
	row := &taskRow{
		ProjectCode:  pid,
		Name:         name,
		Description:  c.PostForm("description"),
		Priority:     0,
		BeginTime:    beginTime,
		EndTime:      endTime,
		ParentTaskId: parentTaskId,
		CreateTime:   time.Now().UnixMilli(),
		MemberCode:   memberId,
		OwnerCode:    memberId,
		AssignTo:     assignTo,
		StageCode:    sid,
		Sort:         maxSort + 1,
		Deleted:      0,
		Private:      0,
		Done:         0,
	}
	if err := db.Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(taskToResponse(h.dbConn(), c, *row)))
}

// @Summary 编辑任务
// @Description 修改已有任务的名称、描述、优先级、截止时间等信息
// @Tags task
// @Accept x-www-form-urlencoded
// @Produce json
// @Param code formData string true "任务加密ID"
// @Param name formData string false "任务名称"
// @Param description formData string false "任务描述"
// @Param priority formData int false "优先级(0-3)"
// @Param end_time formData int64 false "截止时间戳(毫秒)"
// @Param assign_to formData string false "执行者账号"
// @Success 200 {object} common.Result "返回编辑后的任务数据"
// @Failure 400 {object} common.Result "参数错误"
// @Security ApiKeyAuth
// @Router /project/task/edit [post]
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
	if v := c.PostForm("priority"); v != "" {
		if p, err := strconv.ParseInt(v, 10, 8); err == nil {
			updates["priority"] = int8(p)
		}
	}
	if v := c.PostForm("status"); v != "" {
		if s, err := strconv.ParseInt(v, 10, 8); err == nil {
			updates["status"] = int8(s)
		}
	}
	if v := c.PostForm("assign_to"); v != "" {
		if a, err := strconv.ParseInt(v, 10, 64); err == nil {
			updates["assign_to"] = a
		}
	}
	// begin_time: 前端传 "YYYY-MM-DD HH:mm" 格式字符串，转为 Unix 毫秒时间戳
	if beginTimeVals, ok := c.Request.PostForm["begin_time"]; ok {
		beginTimeStr := beginTimeVals[0]
		if beginTimeStr != "" {
			if t, err := time.ParseInLocation("2006-01-02 15:04", beginTimeStr, time.Local); err == nil {
				updates["begin_time"] = t.UnixMilli()
			}
		} else {
			updates["begin_time"] = int64(0)
		}
	}
	// end_time: 前端传 "YYYY-MM-DD HH:mm" 格式字符串，转为 Unix 毫秒时间戳
	if endTimeVals, ok := c.Request.PostForm["end_time"]; ok {
		endTimeStr := endTimeVals[0]
		if endTimeStr != "" {
			if t, err := time.ParseInLocation("2006-01-02 15:04", endTimeStr, time.Local); err == nil {
				updates["end_time"] = t.UnixMilli()
			}
		} else {
			updates["end_time"] = int64(0)
		}
	}
	// work_time: 前端传小时，后端转分钟存储
	if v := c.PostForm("work_time"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			updates["work_time"] = int64(f * 60)
		}
	}
	db := h.db()
	if err := db.Model(&taskRow{}).Where("id=?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"code": codecs.EncryptInt64(id), "projectCode": codecs.EncryptInt64(projectId)}))
}

// @Summary 任务详情
// @Description 获取单个任务的完整详情（含评论、工时记录、附件等）
// @Tags task
// @Accept x-www-form-urlencoded
// @Produce json
// @Param code formData string true "任务加密ID"
// @Success 200 {object} common.Result "返回任务完整数据"
// @Failure 404 {object} common.Result "任务不存在"
// @Security ApiKeyAuth
// @Router /project/task/read [post]
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
	db := h.dbConn()
	var row taskRow
	if err := db.Session(c.Request.Context()).Where("id=?", id).First(&row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	memberId := c.GetInt64("memberId")
	if !authz.IsProjectMember(db.Session(c.Request.Context()), memberId, row.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限访问此任务"))
		return
	}
	var comments []taskCommentRow
	_ = db.Session(c.Request.Context()).Where("task_id=?", row.Id).Order("id desc").Find(&comments).Error
	commentOut := make([]gin.H, 0, len(comments))
	// Batch query comment member info
	commentMemberIds := make([]int64, 0)
	for _, cm := range comments {
		commentMemberIds = append(commentMemberIds, cm.MemberId)
	}
	commentMemberMap := make(map[int64]memberRow)
	if len(commentMemberIds) > 0 {
		var cMembers []memberRow
		_ = db.Session(c.Request.Context()).Where("id IN ?", commentMemberIds).Find(&cMembers).Error
		for _, m := range cMembers {
			commentMemberMap[m.Id] = m
		}
	}
	for _, cm := range comments {
		memberInfo := gin.H{"name": "未知用户", "avatar": "", "code": ""}
		if m, ok := commentMemberMap[cm.MemberId]; ok {
			memberInfo = gin.H{
				"name":   m.Name,
				"avatar": m.Avatar,
				"code":   codecs.EncryptInt64(m.Id),
			}
		}
		commentOut = append(commentOut, gin.H{
			"id":          cm.Id,
			"task_code":   taskCode,
			"comment":     cm.Comment,
			"create_time": cm.CreateTime,
			"member":      memberInfo,
			"memberCode":  codecs.EncryptInt64(cm.MemberId),
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
	db := h.db()
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	memberId := c.GetInt64("memberId")
	var t taskRow
	if err := db.Where("id=?", id).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
	updates := map[string]any{"done": done}
	if done == 1 {
		updates["done_time"] = time.Now().UnixMilli()
	} else {
		updates["done_time"] = int64(0)
	}
	if err := db.Model(&taskRow{}).Where("id=?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}
	// 记录事件日志
	{
		eventType := "task:done"
		remark := "完成了任务"
		if done == 0 {
			eventType = "task:redo"
			remark = "重做了任务"
		}
		_ = db.Create(&taskLogRow{
			ProjectCode:  t.ProjectCode,
			MemberCode:   memberId,
			TaskId:       t.Id,
			EventType:    eventType,
			EventContent: remark + " " + t.Name,
			CreateTime:   time.Now().UnixMilli(),
		}).Error
		// 如果执行者不是操作者，通知执行者
		if t.AssignTo != 0 && t.AssignTo != memberId {
			action := "task:done"
			if done == 0 {
				action = "task:redo"
			}
			sendTaskNotifyToUser(t.AssignTo, action, gin.H{
				"taskCode":    taskCode,
				"projectCode": codecs.EncryptInt64(t.ProjectCode),
				"taskName":    t.Name,
				"done":        done,
			})
		}
	}
	// 广播任务完成状态变更
	broadcastTaskChange("task:done", gin.H{
		"taskCode": taskCode,
		"taskId":   id,
		"done":     done,
	})
	// 触发工作流规则
	triggerWorkflowRules(db, t.ProjectCode, t.StageCode, t.Id, "task_done", done)
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

	if err := h.db().Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"deleted": 1}).Error; err != nil {
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

	if err := h.db().Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"deleted": 0}).Error; err != nil {
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

	if err := h.db().Where("id=?", id).Delete(&taskRow{}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "删除失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTask) recycleBatch(c *gin.Context) {
	result := &common.Result{}
	db := h.db().WithContext(c.Request.Context())
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

// batchDone 批量标记任务完成/重做
func (h *HandlerTask) batchDone(c *gin.Context) {
	result := &common.Result{}
	raw := c.PostForm("taskCodes")
	doneStr := c.PostForm("done")
	var codes []string
	_ = json.Unmarshal([]byte(raw), &codes)
	ids := make([]int64, 0, len(codes))
	for _, code := range codes {
		id, _, ok := authz.CanOperateTask(c, code)
		if ok {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "没有可操作的任务"))
		return
	}
	done := int8(0)
	if doneStr == "1" || doneStr == "true" {
		done = 1
	}
	db := h.db()
	batchUpdates := map[string]any{"done": done}
	if done == 1 {
		batchUpdates["done_time"] = time.Now().UnixMilli()
	} else {
		batchUpdates["done_time"] = int64(0)
	}
	if err := db.Model(&taskRow{}).Where("id in ?", ids).Updates(batchUpdates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "批量操作失败"))
		return
	}
	// 记录事件日志
	memberId := c.GetInt64("memberId")
	for _, id := range ids {
		var t taskRow
		if err := db.Where("id=?", id).First(&t).Error; err == nil {
			eventType := "task:done"
			remark := "完成了任务"
			if done == 0 {
				eventType = "task:redo"
				remark = "重做了任务"
			}
			_ = db.Create(&taskLogRow{
				ProjectCode:  t.ProjectCode,
				MemberCode:   memberId,
				TaskId:       t.Id,
				EventType:    eventType,
				EventContent: remark + " " + t.Name,
				CreateTime:   time.Now().UnixMilli(),
			}).Error
		}
	}
	// WebSocket 广播通知
	broadcastTaskChange("batchDone", gin.H{
		"taskIds": ids,
		"done":    done,
	})
	c.JSON(http.StatusOK, result.Success(gin.H{"count": len(ids)}))
}

func (h *HandlerTask) batchAssignTask(c *gin.Context) {
	result := &common.Result{}
	raw := c.PostForm("taskCodes")
	executorCode := c.PostForm("executorCode")
	var codes []string
	_ = json.Unmarshal([]byte(raw), &codes)
	memberId := c.GetInt64("memberId")
	db := h.db()
	ids := make([]int64, 0, len(codes))
	for _, code := range codes {
		id, err := codecs.DecryptInt64(code)
		if err == nil {
			// 数据隔离：检查当前用户是否是任务所属项目的成员
			var t taskRow
			if err := db.Where("id=?", id).First(&t).Error; err == nil {
				if authz.IsProjectMember(db, memberId, t.ProjectCode) {
					ids = append(ids, id)
				}
			}
		}
	}
	execId, err := codecs.DecryptInt64(executorCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "executorCode无效"))
		return
	}
	if len(ids) > 0 {
		_ = h.db().Model(&taskRow{}).Where("id in ?", ids).Updates(map[string]any{"assign_to": execId}).Error
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
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	db := h.db()
	memberId := c.GetInt64("memberId")
	var t taskRow
	if err := db.Where("id=?", id).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
	private := int8(0)
	if privateStr == "1" || privateStr == "true" {
		private = 1
	}
	_ = db.Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"private": private}).Error
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
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	db := h.db()
	memberId := c.GetInt64("memberId")
	var t taskRow
	if err := db.Where("id=?", id).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
	like := 0
	if likeStr == "1" || likeStr == "true" {
		like = 1
	}
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
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	db := h.db()
	memberId := c.GetInt64("memberId")
	var t taskRow
	if err := db.Where("id=?", id).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
	star := 0
	if starStr == "1" || starStr == "true" {
		star = 1
	}
	_ = db.Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"star": star}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// @Summary 创建任务评论
// @Description 在指定任务下添加评论
// @Tags task
// @Accept x-www-form-urlencoded
// @Produce json
// @Param code formData string true "任务加密ID"
// @Param content formData string true "评论内容"
// @Success 200 {object} common.Result "返回新评论ID"
// @Failure 400 {object} common.Result "参数错误"
// @Security ApiKeyAuth
// @Router /project/task/createComment [post]
func (h *HandlerTask) createComment(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	comment := c.PostForm("comment")
	mentionsRaw := c.PostForm("mentions")
	id, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}
	memberId := c.GetInt64("memberId")
	db := h.db()
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	var t taskRow
	if err := db.Where("id=?", id).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
	row := &taskCommentRow{TaskId: id, MemberId: memberId, Comment: comment, CreateTime: time.Now().UnixMilli()}
	if err := db.Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "评论创建失败"))
		return
	}
	// 记录评论事件日志（t 已在上方查询）
	_ = db.Create(&taskLogRow{
		ProjectCode:  t.ProjectCode,
		MemberCode:   memberId,
		TaskId:       t.Id,
		EventType:    "task:comment",
		EventContent: comment,
		CreateTime:   time.Now().UnixMilli(),
	}).Error
	// 解析 @mentions 并发送通知
	var mentions []string
	if mentionsRaw != "" {
		_ = json.Unmarshal([]byte(mentionsRaw), &mentions)
	}
	if len(mentions) > 0 {
		// 查询被 @ 成员的 ID 并发送通知
		for _, mName := range mentions {
			var m memberRow
			if err := db.Where("name=?", mName).First(&m).Error; err == nil && m.Id != 0 {
				// 跳过@自己
				if m.Id == memberId {
					continue
				}
				sendTaskNotifyToUser(m.Id, "task:mention", gin.H{
					"taskCode":    taskCode,
					"taskId":      id,
					"projectCode": codecs.EncryptInt64(t.ProjectCode),
					"comment":     comment,
					"fromUser":    memberId,
				})
			}
		}
	}
	// 给任务执行者发送评论通知（如果执行者不是评论者本人）
	if t.AssignTo != 0 && t.AssignTo != memberId {
		sendTaskNotifyToUser(t.AssignTo, "task:comment", gin.H{
			"taskCode":    taskCode,
			"projectCode": codecs.EncryptInt64(t.ProjectCode),
			"comment":     comment,
			"fromUser":    memberId,
		})
	}
	// 广播评论变更
	broadcastTaskChange("task:comment", gin.H{
		"taskCode": taskCode,
		"taskId":   id,
		"comment":  comment,
	})
	c.JSON(http.StatusOK, result.Success(gin.H{"id": row.Id}))
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
	db := h.db()
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	memberId := c.GetInt64("memberId")
	var moved taskRow
	if err := db.Where("id=?", movedId).First(&moved).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, moved.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
	var list []taskRow
	if err := db.Where("stage_code=? and deleted=0", stageId).Order("sort asc, id asc").Find(&list).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "排序失败"))
		return
	}
	moved.StageCode = stageId
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
	// 触发工作流规则（任务移动到新阶段）
	triggerWorkflowRules(db, moved.ProjectCode, stageId, movedId, "task_moved", 0)
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
	TaskId       int64 `gorm:"column:task_id;default:0"`
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
	id, _, ok := authz.CanOperateTask(c, taskCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
	execId, err := codecs.DecryptInt64(executorCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "executorCode无效"))
		return
	}
	if err := h.db().Model(&taskRow{}).Where("id=?", id).Updates(map[string]any{"assign_to": execId}).Error; err != nil {
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
			_ = h.db().Select("project_code").Where("id=?", taskId).First(&task).Error
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

	// 数据隔离：检查当前用户是否是项目成员
	memberId := c.GetInt64("memberId")
	db := h.db().WithContext(c.Request.Context())
	if !authz.IsProjectMember(db, memberId, pid) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限访问此项目"))
		return
	}

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
	// 数据隔离：检查当前用户是否是项目成员
	memberId := c.GetInt64("memberId")
	if !authz.IsProjectMember(h.db(), memberId, pid) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限访问此项目"))
		return
	}
	db := h.dbConn()
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
	db := h.db().WithContext(c.Request.Context())
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	memberId := c.GetInt64("memberId")
	var t taskRow
	if err := db.Where("id=?", id).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限访问此任务"))
		return
	}
	var tags []taskTagRow
	_ = db.Table("ms_task_tag tg").
		Joins("join ms_task_tag_rel r on r.tag_id=tg.id").
		Where("r.task_id=? and tg.deleted=0", id).
		Select("tg.*").
		Find(&tags).Error
	out := make([]gin.H, 0, len(tags))
	for _, t := range tags {
		out = append(out, gin.H{
			"code": codecs.EncryptInt64(t.Id),
			"tag": gin.H{
				"color": t.Color,
				"name":  t.Name,
			},
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
	db := h.db()
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	memberId := c.GetInt64("memberId")
	var t taskRow
	if err := db.Where("id=?", taskId).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
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
	// 数据隔离：检查当前用户是否是项目成员
	memberId := c.GetInt64("memberId")
	db := h.db().WithContext(c.Request.Context())
	if !authz.IsProjectMember(db, memberId, pid) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限访问此项目"))
		return
	}
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
	onlyComment := c.PostForm("comment")
	showAll := c.PostForm("all")
	db := h.db().WithContext(c.Request.Context())
	query := db.Model(&taskLogRow{})
	memberId := c.GetInt64("memberId")
	if taskCode != "" {
		if id, err := codecs.DecryptInt64(taskCode); err == nil && id != 0 {
			query = query.Where("task_id=?", id)
			// 数据隔离：检查当前用户是否是任务所属项目的成员
			var t taskRow
			if err := db.Where("id=?", id).First(&t).Error; err == nil {
				if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
					c.JSON(http.StatusOK, result.Success(gin.H{"list": []any{}, "total": 0}))
					return
				}
			}
		}
	} else if projectCode != "" {
		if pid, err := codecs.DecryptInt64(projectCode); err == nil && pid != 0 {
			// 数据隔离：检查当前用户是否是项目成员
			if !authz.IsProjectMember(db, memberId, pid) {
				c.JSON(http.StatusOK, result.Success(gin.H{"list": []any{}, "total": 0}))
				return
			}
			query = query.Where("project_code=?", pid)
		}
	} else {
		// 未指定任务和项目时，只返回用户参与的项目日志
		query = query.Where("project_code IN (SELECT project_code FROM ms_project_member WHERE member_code=?)", memberId)
	}
	// 支持仅查看评论
	if onlyComment == "1" {
		query = query.Where("event_type=?", "task:comment")
	}
	// showAll=1 时不分页
	if showAll != "1" {
		var total int64
		_ = query.Count(&total).Error
		_ = total // 保留总数供后面使用
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
	db := h.db()
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	memberId := c.GetInt64("memberId")
	var t taskRow
	if err := db.Where("id=?", id).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限访问此任务"))
		return
	}
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
		// workTime 字段存的是分钟，转换为小时
		num := float64(r.WorkTime) / 60.0
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
	// 数据隔离：检查当前用户是否是任务所属项目的成员
	db := h.db()
	var t taskRow
	if err := db.Where("id=?", id).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
	// 前端传 num（小时值）和 timeUnit（可选，默认 hour），转为分钟存储
	numStr := c.PostForm("num")
	if numStr == "" {
		numStr = c.PostForm("workTime") // 兼容旧接口
	}
	timeUnit := c.PostForm("timeUnit")
	if timeUnit == "" {
		timeUnit = "hour"
	}
	var numFloat float64
	if f, err := strconv.ParseFloat(numStr, 64); err == nil {
		// 按单位转为分钟
		switch timeUnit {
		case "day":
			numFloat = f * 8 * 60 // 1天=8小时
		case "hour":
			numFloat = f * 60
		case "minute":
			numFloat = f
		default:
			numFloat = f * 60
		}
	}
	workTimeMinutes := int64(numFloat)
	row := &taskWorkTimeRow{
		TaskId:     id,
		MemberId:   memberId,
		WorkTime:   workTimeMinutes,
		Remark:     c.PostForm("remark"),
		CreateTime: time.Now().UnixMilli(),
	}
	_ = h.db().Create(row).Error
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
	wtId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || wtId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "id无效"))
		return
	}
	db := h.db()
	// 数据隔离：检查当前用户是否是工时关联任务所属项目的成员
	memberId := c.GetInt64("memberId")
	var wt taskWorkTimeRow
	if err := db.Where("id=?", wtId).First(&wt).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "工时记录不存在"))
		return
	}
	var t taskRow
	if err := db.Where("id=?", wt.TaskId).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
	updates := map[string]any{}
	if v := c.PostForm("remark"); v != "" {
		updates["remark"] = v
	}
	// 前端传 num（数值）和 timeUnit（单位），转为分钟存储
	numStr := c.PostForm("num")
	if numStr == "" {
		numStr = c.PostForm("workTime") // 兼容旧接口
	}
	if numStr != "" {
		timeUnit := c.PostForm("timeUnit")
		if timeUnit == "" {
			timeUnit = "hour"
		}
		if f, err := strconv.ParseFloat(numStr, 64); err == nil && f > 0 {
			var minutes float64
			switch timeUnit {
			case "day":
				minutes = f * 8 * 60
			case "hour":
				minutes = f * 60
			case "minute":
				minutes = f
			default:
				minutes = f * 60
			}
			updates["work_time"] = int64(minutes)
		}
	}
	if len(updates) == 0 {
		c.JSON(http.StatusOK, result.Success([]int{}))
		return
	}
	if err := db.Model(&taskWorkTimeRow{}).Where("id=?", wtId).Updates(updates).Error; err != nil {
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
	wtId, err := codecs.DecryptInt64(code)
	if err != nil || wtId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	db := h.db()
	// 数据隔离：检查当前用户是否是工时关联任务所属项目的成员
	memberId := c.GetInt64("memberId")
	var wt taskWorkTimeRow
	if err := db.Where("id=?", wtId).First(&wt).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "工时记录不存在"))
		return
	}
	var t taskRow
	if err := db.Where("id=?", wt.TaskId).First(&t).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "任务不存在"))
		return
	}
	if !authz.IsProjectMember(db, memberId, t.ProjectCode) {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此任务"))
		return
	}
	if err := db.Where("id=?", wtId).Delete(&taskWorkTimeRow{}).Error; err != nil {
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

// uploadFile 上传任务附件（保存到本地磁盘 + ms_file 表记录）
// @Summary 上传任务附件
// @Description 上传文件作为任务附件，保存到本地磁盘并记录到数据库
// @Tags task
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "附件文件(最大30MB)"
// @Param projectCode formData string true "项目加密ID"
// @Success 200 {object} common.Result "返回文件信息(file_name/file_size/file_url/file_type)"
// @Failure 400 {object} common.Result "参数错误或文件过大"
// @Security ApiKeyAuth
// @Router /project/task/uploadFile [post]
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

	// 数据隔离：检查当前用户是否是项目成员
	if pid != 0 {
		db := h.db()
		if !authz.IsProjectMember(db, memberId, pid) {
			c.JSON(http.StatusOK, result.Fail(403, "无权限操作此项目"))
			return
		}
	}

	// 限制文件大小 30MB
	if header.Size > 30*1024*1024 {
		c.JSON(http.StatusOK, result.Fail(400, "文件大小不能超过30MB"))
		return
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	fileType := mapFileType(ext)

	// 创建上传目录（按日期分目录存储）
	uploadDir := filepath.Join("uploads", "files", time.Now().Format("2006/01/02"))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建上传目录失败: "+err.Error()))
		return
	}

	// 生成唯一文件名
	fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), memberId, ext)
	savePath := filepath.Join(uploadDir, fileName)

	// 写入磁盘
	dst, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建文件失败"))
		return
	}
	defer dst.Close()

	fileSize, err := io.Copy(dst, file)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "保存文件失败"))
		return
	}

	// 生成访问URL
	fileUrl := "/" + strings.ReplaceAll(savePath, "\\", "/")

	// 记录上传事件到 ms_project_event
	row := &taskLogRow{
		ProjectCode:  pid,
		MemberCode:   memberId,
		TaskId:       0,
		EventType:    "upload_file",
		EventContent: "上传文件: " + header.Filename,
		CreateTime:   time.Now().UnixMilli(),
	}
	_ = h.db().Create(row).Error

	c.JSON(http.StatusOK, result.Success(gin.H{
		"id":         row.Id,
		"file_name":  header.Filename,
		"file_size":  fileSize,
		"file_url":   fileUrl,
		"file_path":  savePath,
		"file_type":  fileType,
	}))
}

// mapFileType 根据扩展名返回文件类型
func mapFileType(ext string) string {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg"}
	docExts := []string{".doc", ".docx", ".pdf", ".txt", ".xls", ".xlsx", ".ppt", ".pptx"}
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv"}
	audioExts := []string{".mp3", ".wav", ".flac", ".aac", ".ogg"}

	for _, e := range imageExts {
		if ext == e { return "image" }
	}
	for _, e := range docExts {
		if ext == e { return "doc" }
	}
	for _, e := range videoExts {
		if ext == e { return "video" }
	}
	for _, e := range audioExts {
		if ext == e { return "audio" }
	}
	return "other"
}

// ==================== WebSocket 推送 & 通知辅助 ====================

// notifyRow 通知记录（复用 ms_notify 表）
type notifyRow struct {
	Id         int64  `gorm:"primaryKey;autoIncrement"`
	MemberCode int64  `gorm:"column:member_code"`
	Title      string `gorm:"column:title"`
	Content    string `gorm:"column:content"`
	Type       int8   `gorm:"column:type"`
	IsRead     int8   `gorm:"column:is_read"`
	CreateTime int64  `gorm:"column:create_time"`
	Action     string `gorm:"column:action"`
	SendData   string `gorm:"column:send_data"`
}

func (*notifyRow) TableName() string { return "ms_notify" }

// broadcastTaskChange 通过 WebSocket 广播任务变更事件
func broadcastTaskChange(action string, data interface{}) {
	msg := ws.Message{Action: action, Data: data}
	ws.Manager.Broadcast(msg)
}

// sendTaskNotifyToUser 向指定用户发送任务通知（WebSocket + 通知记录）
func sendTaskNotifyToUser(userId int64, action string, data interface{}) {
	// WebSocket 推送
	userIDStr := strconv.FormatInt(userId, 10)
	msg := ws.Message{Action: action, Data: data}
	ws.Manager.SendToUser(userIDStr, msg)

	// 写入通知记录
	title := "任务通知"
	contentText := ""
	switch action {
	case "task:mention":
		title = "你在评论中被提及"
		contentText = extractComment(data)
	case "task:comment":
		title = "任务有新评论"
		contentText = extractComment(data)
	case "task:done":
		title = "任务已完成"
		contentText = extractTaskName(data)
	case "task:redo":
		title = "任务被重做"
		contentText = extractTaskName(data)
	case "task:assign":
		title = "任务已指派给你"
		contentText = extractTaskName(data)
	}
	if contentText == "" {
		dataJSON, _ := json.Marshal(data)
		contentText = string(dataJSON)
	}
	dataJSON, _ := json.Marshal(data)
	_ = gorms.GetDB().Create(&notifyRow{
		MemberCode: userId,
		Title:      title,
		Content:    contentText,
		Type:       1, // notice
		IsRead:     0,
		CreateTime: time.Now().UnixMilli(),
		Action:     action,
		SendData:   string(dataJSON),
	}).Error
}

// extractComment 从 data 中提取评论内容（兼容 gin.H 和 map[string]interface{}）
func extractComment(data interface{}) string {
	switch d := data.(type) {
	case gin.H:
		if c, _ := d["comment"].(string); c != "" {
			return c
		}
	case map[string]interface{}:
		if c, _ := d["comment"].(string); c != "" {
			return c
		}
	}
	return ""
}

// extractTaskName 从 data 中提取任务名称（兼容 gin.H 和 map[string]interface{}）
func extractTaskName(data interface{}) string {
	switch d := data.(type) {
	case gin.H:
		if n, _ := d["taskName"].(string); n != "" {
			return n
		}
	case map[string]interface{}:
		if n, _ := d["taskName"].(string); n != "" {
			return n
		}
	}
	return ""
}

// workflowRuleRow 工作流规则数据行
type workflowRuleRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64  `gorm:"column:project_code"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	Rules       string `gorm:"column:rules"`
	Sort        int    `gorm:"column:sort"`
	CreateTime  int64  `gorm:"column:create_time"`
	UpdateTime  int64  `gorm:"column:update_time"`
}

func (*workflowRuleRow) TableName() string { return "ms_task_workflow" }

// workflowRule 解析后的工作流规则
type workflowRule struct {
	TaskWorkflowName string `json:"taskWorkflowName"`
	FirstObj         string `json:"firstObj"`           // 触发的任务列表 code
	FirstAction      struct {
		Action int    `json:"action"` // 0=增加/移动, 1=完成, 2=重做, 3=设置执行人
		Value  string `json:"value"`
	} `json:"firstAction"`
	FirstResult struct {
		Action int    `json:"action"` // 0=自动流转到, 3=默认指派给
		Value  string `json:"value"`
	} `json:"firstResult"`
	LastResult struct {
		Action int    `json:"action"`
		Value  string `json:"value"`
	} `json:"lastResult"`
	State struct {
		Action int `json:"action"`
		Value  int `json:"value"` // -1=不做修改, 1=已完成, 2=未完成
	} `json:"state"`
}

// triggerWorkflowRules 工作流规则自动触发引擎
// eventType: "task_done" (完成/重做), "task_moved" (移动到新阶段), "task_created" (新增)
func triggerWorkflowRules(db *gorm.DB, projectCode int64, stageCode int64, taskId int64, eventType string, done int8) {
	// 查询项目下的所有工作流规则
	var workflows []workflowRuleRow
	if err := db.Where("project_code=?", projectCode).Find(&workflows).Error; err != nil || len(workflows) == 0 {
		return
	}

	stageCodeStr := codecs.EncryptInt64(stageCode)

	for _, wf := range workflows {
		if wf.Rules == "" {
			continue
		}

		var rule workflowRule
		if err := json.Unmarshal([]byte(wf.Rules), &rule); err != nil {
			continue
		}

		// 检查触发条件是否匹配
		if !matchWorkflowTrigger(rule, stageCodeStr, eventType, done) {
			continue
		}

		// 执行工作流结果
		executeWorkflowResult(db, rule, taskId, projectCode)
	}
}

// matchWorkflowTrigger 检查工作流规则触发条件是否匹配
func matchWorkflowTrigger(rule workflowRule, stageCodeStr string, eventType string, done int8) bool {
	// 检查任务列表是否匹配
	if rule.FirstObj != "" && rule.FirstObj != stageCodeStr {
		return false
	}

	// 检查触发动作是否匹配
	action := rule.FirstAction.Action
	switch eventType {
	case "task_done":
		if done == 1 && action != 1 { // 完成
			return false
		}
		if done == 0 && action != 2 { // 重做
			return false
		}
	case "task_moved":
		if action != 0 { // 增加/移动
			return false
		}
	case "task_created":
		if action != 0 {
			return false
		}
	default:
		return false
	}

	return true
}

// executeWorkflowResult 执行工作流规则的结果动作
func executeWorkflowResult(db *gorm.DB, rule workflowRule, taskId int64, projectCode int64) {
	resultAction := rule.FirstResult.Action

	switch resultAction {
	case 0:
		// 自动流转到指定任务列表
		if rule.FirstResult.Value != "" {
			targetStageId, err := codecs.DecryptInt64(rule.FirstResult.Value)
			if err == nil && targetStageId != 0 {
				_ = db.Model(&taskRow{}).Where("id=?", taskId).Update("stage_code", targetStageId)
			}
		}
		// 如果设置了最终执行者
		if rule.LastResult.Value != "" {
			assignToId, err := codecs.DecryptInt64(rule.LastResult.Value)
			if err == nil && assignToId != 0 {
				_ = db.Model(&taskRow{}).Where("id=?", taskId).Update("assign_to", assignToId)
			}
		}
	case 3:
		// 默认指派给指定执行者
		if rule.FirstResult.Value != "" {
			assignToId, err := codecs.DecryptInt64(rule.FirstResult.Value)
			if err == nil && assignToId != 0 {
				_ = db.Model(&taskRow{}).Where("id=?", taskId).Update("assign_to", assignToId)
			}
		}
		// 如果设置了流转目标
		if rule.LastResult.Value != "" {
			targetStageId, err := codecs.DecryptInt64(rule.LastResult.Value)
			if err == nil && targetStageId != 0 {
				_ = db.Model(&taskRow{}).Where("id=?", taskId).Update("stage_code", targetStageId)
			}
		}
	}

	// 修改任务状态
	if rule.State.Value > 0 {
		if rule.State.Value == 1 {
			// 标记为已完成
			_ = db.Model(&taskRow{}).Where("id=?", taskId).Updates(map[string]any{"done": 1})
		} else if rule.State.Value == 2 {
			// 标记为未完成
			_ = db.Model(&taskRow{}).Where("id=?", taskId).Updates(map[string]any{"done": 0})
		}
	}
}

// orgCodeFromContext 从 gin 上下文中获取解密后的 organizationCode
func orgCodeFromContext(c *gin.Context) int64 {
	orgVal, _ := c.Get("organizationCode")
	switch t := orgVal.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case string:
		decrypted, err := codecs.DecryptInt64(t)
		if err == nil && decrypted > 0 {
			return decrypted
		}
		i, _ := strconv.ParseInt(t, 10, 64)
		return i
	default:
		return 0
	}
}
