package task_workflow

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	common "test.com/project-common"
)

type HandlerTaskWorkflow struct {
}

func New() *HandlerTaskWorkflow {
	return &HandlerTaskWorkflow{}
}

type taskWorkflowRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64  `gorm:"column:project_code"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	Rules       string `gorm:"column:rules"`
	Sort        int    `gorm:"column:sort"`
	CreateTime  int64  `gorm:"column:create_time"`
	UpdateTime  int64  `gorm:"column:update_time"`
}

func (*taskWorkflowRow) TableName() string { return "ms_task_workflow" }

// list 获取工作流列表
func (h *HandlerTaskWorkflow) list(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")

	if projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode不能为空"))
		return
	}

	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil || pid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var rows []taskWorkflowRow
	db.Where("project_code=?", pid).Order("sort asc, id asc").Find(&rows)

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"code":        codecs.EncryptInt64(r.Id),
			"name":        r.Name,
			"description": r.Description,
			"rules":       r.Rules,
			"sort":        r.Sort,
			"createTime":  r.CreateTime,
			"projectCode": projectCode,
		})
	}

	c.JSON(http.StatusOK, result.Success(out))
}

// save 创建工作流
func (h *HandlerTaskWorkflow) save(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	name := c.PostForm("name")
	description := c.PostForm("description")
	rules := c.PostForm("rules")

	if projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode不能为空"))
		return
	}

	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil || pid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}

	if name == "" {
		c.JSON(http.StatusOK, result.Fail(400, "名称不能为空"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 获取当前最大排序
	var maxSort int
	db.Model(&taskWorkflowRow{}).Where("project_code=?", pid).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort)

	now := time.Now().UnixMilli()
	row := &taskWorkflowRow{
		ProjectCode: pid,
		Name:        name,
		Description: description,
		Rules:       rules,
		Sort:        maxSort + 1,
		CreateTime:  now,
		UpdateTime:  now,
	}

	if err := db.Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":        codecs.EncryptInt64(row.Id),
		"name":        row.Name,
		"description": row.Description,
		"rules":       row.Rules,
		"sort":        row.Sort,
		"createTime":  row.CreateTime,
		"projectCode": projectCode,
	}))
}

// edit 编辑工作流
func (h *HandlerTaskWorkflow) edit(c *gin.Context) {
	result := &common.Result{}
	workflowCode := c.PostForm("workflowCode")
	name := c.PostForm("name")
	description := c.PostForm("description")
	rules := c.PostForm("rules")

	if workflowCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "workflowCode不能为空"))
		return
	}

	wid, err := codecs.DecryptInt64(workflowCode)
	if err != nil || wid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "workflowCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	updates := map[string]any{
		"update_time": time.Now().UnixMilli(),
	}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}
	if rules != "" {
		updates["rules"] = rules
	}

	if err := db.Model(&taskWorkflowRow{}).Where("id=?", wid).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "编辑失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": workflowCode}))
}

// del 删除工作流
func (h *HandlerTaskWorkflow) del(c *gin.Context) {
	result := &common.Result{}
	workflowCode := c.PostForm("workflowCode")

	if workflowCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "workflowCode不能为空"))
		return
	}

	wid, err := codecs.DecryptInt64(workflowCode)
	if err != nil || wid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "workflowCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	db.Delete(&taskWorkflowRow{}, wid)

	c.JSON(http.StatusOK, result.Success(gin.H{"code": workflowCode}))
}

// getTaskWorkflowRules 获取工作流规则
func (h *HandlerTaskWorkflow) getTaskWorkflowRules(c *gin.Context) {
	result := &common.Result{}
	workflowCode := c.PostForm("workflowCode")

	if workflowCode == "" {
		// 返回默认规则模板
		defaultRules := []gin.H{
			{
				"name":        "状态流转规则",
				"description": "定义任务状态之间的流转规则",
				"rules": []gin.H{
					{"from": "待办", "to": "进行中", "condition": "assign_to != null"},
					{"from": "进行中", "to": "已完成", "condition": "done == true"},
					{"from": "进行中", "to": "待办", "condition": "assign_to == null"},
				},
			},
			{
				"name":        "自动分配规则",
				"description": "根据条件自动分配任务",
				"rules": []gin.H{
					{"condition": "priority == 2", "action": "assign_to_creator"},
					{"condition": "priority == 3", "action": "notify_all_members"},
				},
			},
			{
				"name":        "通知规则",
				"description": "定义任务变更时的通知规则",
				"rules": []gin.H{
					{"event": "task_created", "notify": "project_members"},
					{"event": "task_completed", "notify": "task_creator"},
					{"event": "task_overdue", "notify": "assignee"},
				},
			},
		}
		c.JSON(http.StatusOK, result.Success(defaultRules))
		return
	}

	wid, err := codecs.DecryptInt64(workflowCode)
	if err != nil || wid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "workflowCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var workflow taskWorkflowRow
	if err := db.Where("id=?", wid).First(&workflow).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "工作流不存在"))
		return
	}

	// 解析规则并返回
	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":        workflowCode,
		"name":        workflow.Name,
		"description": workflow.Description,
		"rules":       workflow.Rules,
	}))
}
