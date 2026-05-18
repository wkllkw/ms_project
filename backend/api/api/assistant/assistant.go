package assistant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test.com/project-api/api/midd"
	"test.com/project-api/internal/authz"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/router"
)

type RouterAssistant struct{}

func init() {
	log.Println("init assistant router")
	ru := &RouterAssistant{}
	router.Register(ru)
}

func (*RouterAssistant) Route(r *gin.Engine) {
	h := New()
	group := r.Group("/project/assistant")
	group.Use(midd.TokenVerify())
	group.POST("/chat", h.chat)
}

type HandlerAssistant struct{}

func New() *HandlerAssistant {
	return &HandlerAssistant{}
}

const (
	openclawBaseURL = "http://127.0.0.1:35985"
	openclawToken   = "6083c22ccbc4c8776e67a2d1112d555d88b3c42b6bf3dee6"
	maxToolRounds   = 5 // 最大工具调用轮数，防止无限循环
)

// ==================== 请求/响应结构 ====================

type chatMessage struct {
	Role       string                 `json:"role"`
	Content    string                 `json:"content,omitempty"`
	ToolCalls  []toolCall             `json:"tool_calls,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	Name       string                 `json:"name,omitempty"`
}

type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function toolFunction `json:"function"`
}

type toolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// parseMessages 从请求体中解析 messages 数组（兼容 form-urlencoded 和 json）
func parseMessages(c *gin.Context) ([]chatMessage, error) {
	contentType := c.GetHeader("Content-Type")
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bodyBytes) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	defer c.Request.Body.Close()

	var msgs []chatMessage
	if strings.Contains(contentType, "application/json") {
		err = json.Unmarshal(bodyBytes, &msgs)
	} else if strings.Contains(contentType, "form") {
		values, parseErr := url.ParseQuery(string(bodyBytes))
		if parseErr != nil {
			return nil, parseErr
		}
		messagesStr := values.Get("messages")
		if messagesStr != "" {
			err = json.Unmarshal([]byte(messagesStr), &msgs)
		} else {
			return nil, fmt.Errorf("no messages field")
		}
	} else {
		err = json.Unmarshal(bodyBytes, &msgs)
	}

	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// ==================== 工具定义 ====================

// getToolDefinitions 返回 OpenAI function calling 格式的工具定义
func getToolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "list_projects",
				"description": "获取当前用户参与的项目列表，可按类型筛选（active=活跃, archive=已归档, deleted=已删除）",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"select_by": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"active", "archive", "deleted", "collect"},
							"description": "筛选类型：active=活跃项目(默认), archive=已归档, deleted=已删除, collect=已收藏",
						},
						"page": map[string]interface{}{
							"type":        "integer",
							"description": "页码，默认1",
						},
						"page_size": map[string]interface{}{
							"type":        "integer",
							"description": "每页条数，默认10",
						},
					},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "get_project_detail",
				"description": "获取项目详细信息，包含项目名称、描述、进度、创建时间、负责人等",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project_code": map[string]interface{}{
							"type":        "string",
							"description": "项目的加密ID（code）",
						},
					},
					"required": []string{"project_code"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "create_project",
				"description": "创建一个新项目。创建后你将成为项目负责人",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "项目名称",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "项目描述（可选）",
						},
					},
					"required": []string{"name"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "edit_project",
				"description": "编辑项目的名称或描述",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project_code": map[string]interface{}{
							"type":        "string",
							"description": "项目的加密ID",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "新项目名称（可选）",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "新项目描述（可选）",
						},
					},
					"required": []string{"project_code"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "list_tasks",
				"description": "获取指定项目下的任务列表，支持按状态、阶段筛选",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project_code": map[string]interface{}{
							"type":        "string",
							"description": "项目的加密ID",
						},
						"page": map[string]interface{}{
							"type":        "integer",
							"description": "页码，默认1",
						},
						"page_size": map[string]interface{}{
							"type":        "integer",
							"description": "每页条数，默认20",
						},
					},
					"required": []string{"project_code"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "get_my_tasks",
				"description": "获取分配给当前用户的任务列表，可按完成状态筛选",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"done": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"0", "1"},
							"description": "0=未完成任务, 1=已完成任务，不传则返回全部",
						},
						"page": map[string]interface{}{
							"type":        "integer",
							"description": "页码，默认1",
						},
						"page_size": map[string]interface{}{
							"type":        "integer",
							"description": "每页条数，默认20",
						},
					},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "create_task",
				"description": "在指定项目中创建一个新任务。需要提供项目ID和任务名称",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project_code": map[string]interface{}{
							"type":        "string",
							"description": "项目的加密ID",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "任务名称",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "任务描述（可选）",
						},
						"priority": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"0", "1", "2"},
							"description": "优先级：0=普通(默认), 1=紧急, 2=非常紧急",
						},
					},
					"required": []string{"project_code", "name"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "edit_task",
				"description": "编辑任务的名称、描述或优先级",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task_code": map[string]interface{}{
							"type":        "string",
							"description": "任务的加密ID",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "新任务名称（可选）",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "新任务描述（可选）",
						},
						"priority": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"0", "1", "2"},
							"description": "优先级（可选）",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"0", "1", "2", "3", "4"},
							"description": "任务状态（可选）：0=未开始, 1=已完成, 2=进行中, 3=挂起, 4=测试中",
						},
					},
					"required": []string{"task_code"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "complete_task",
				"description": "将任务标记为已完成或重做（取消完成）",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task_code": map[string]interface{}{
							"type":        "string",
							"description": "任务的加密ID",
						},
						"done": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"1", "0"},
							"description": "1=标记完成, 0=重做（取消完成）",
						},
					},
					"required": []string{"task_code", "done"},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "get_project_stats",
				"description": "获取项目统计信息，包括任务总数、完成数、逾期数、成员数等。不传 project_code 则返回全局统计",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project_code": map[string]interface{}{
							"type":        "string",
							"description": "项目的加密ID（可选，不传则返回全局统计）",
						},
					},
				},
			},
		},
		{
			"type": "function",
			"function": map[string]interface{}{
				"name":        "list_project_members",
				"description": "获取项目的成员列表",
				"parameters": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project_code": map[string]interface{}{
							"type":        "string",
							"description": "项目的加密ID",
						},
					},
					"required": []string{"project_code"},
				},
			},
		},
	}
}

// ==================== 工具执行 ====================

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

// executeTool 执行工具调用，返回 JSON 结果
func executeTool(c *gin.Context, toolName string, arguments map[string]interface{}) string {
	memberId := c.GetInt64("memberId")
	orgCode := orgCodeFromContext(c)
	db := gorms.GetDB().WithContext(c.Request.Context())

	switch toolName {
	case "list_projects":
		return execListProjects(db, memberId, orgCode, arguments)
	case "get_project_detail":
		return execGetProjectDetail(db, memberId, arguments)
	case "create_project":
		return execCreateProject(db, memberId, orgCode, arguments)
	case "edit_project":
		return execEditProject(db, memberId, arguments)
	case "list_tasks":
		return execListTasks(db, memberId, arguments)
	case "get_my_tasks":
		return execGetMyTasks(db, memberId, orgCode, arguments)
	case "create_task":
		return execCreateTask(db, memberId, arguments)
	case "edit_task":
		return execEditTask(db, memberId, arguments)
	case "complete_task":
		return execCompleteTask(db, memberId, arguments)
	case "get_project_stats":
		return execGetProjectStats(db, memberId, orgCode, arguments)
	case "list_project_members":
		return execListProjectMembers(db, memberId, arguments)
	default:
		result, _ := json.Marshal(map[string]string{"error": "未知的工具: " + toolName})
		return string(result)
	}
}

// execListProjects 获取项目列表
func execListProjects(db *gorm.DB, memberId int64, orgCode int64, args map[string]interface{}) string {
	selectBy := "active"
	if v, ok := args["select_by"].(string); ok && v != "" {
		selectBy = v
	}
	page := 1
	if v, ok := args["page"].(float64); ok && v > 0 {
		page = int(v)
	}
	pageSize := 10
	if v, ok := args["page_size"].(float64); ok && v > 0 {
		pageSize = int(v)
	}

	base := db.Table("ms_project p").
		Joins("join ms_project_member pm on pm.project_code=p.id").
		Joins("left join ms_project_collection pc on pc.project_code=p.id and pc.member_code=?", memberId).
		Joins("left join ms_project_member opm on opm.project_code=p.id and opm.is_owner=1").
		Joins("left join ms_member owner on owner.id=opm.member_code").
		Where("pm.member_code=?", memberId)
	if orgCode != 0 {
		base = base.Where("p.organization_code=?", orgCode)
	}
	switch selectBy {
	case "collect":
		base = base.Where("pc.id is not null").Where("p.deleted=0").Where("p.archive=0")
	case "archive":
		base = base.Where("p.deleted=0").Where("p.archive=1")
	case "deleted":
		base = base.Where("p.deleted=1")
	default:
		base = base.Where("p.deleted=0").Where("p.archive=0")
	}

	var total int64
	_ = base.Distinct("p.id").Count(&total).Error

	type projectRow struct {
		Id          int64
		Name        string
		Description string
		Schedule    float64
		CreateTime  int64 `gorm:"column:create_time"`
		OwnerName   string `gorm:"column:owner_name"`
	}
	var rows []projectRow
	err := base.Select("p.id, p.name, p.description, p.schedule, p.create_time, coalesce(owner.name,'') as owner_name").
		Order("p.id desc").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error
	if err != nil {
		result, _ := json.Marshal(map[string]string{"error": "查询失败: " + err.Error()})
		return string(result)
	}

	type projItem struct {
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Schedule    float64 `json:"schedule"`
		CreateTime  int64   `json:"create_time"`
		OwnerName   string  `json:"owner_name"`
	}
	list := make([]projItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, projItem{
			Code:        codecs.EncryptInt64(r.Id),
			Name:        r.Name,
			Description: r.Description,
			Schedule:    r.Schedule,
			CreateTime:  r.CreateTime,
			OwnerName:   r.OwnerName,
		})
	}
	result, _ := json.Marshal(map[string]interface{}{
		"total": total,
		"page":  page,
		"list":  list,
	})
	return string(result)
}

// execGetProjectDetail 获取项目详情
func execGetProjectDetail(db *gorm.DB, memberId int64, args map[string]interface{}) string {
	projectCode, _ := args["project_code"].(string)
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		result, _ := json.Marshal(map[string]string{"error": "project_code 无效"})
		return string(result)
	}
	if !authz.IsProjectMember(db, memberId, projectId) {
		result, _ := json.Marshal(map[string]string{"error": "无权限访问此项目"})
		return string(result)
	}

	type projRow struct {
		Id          int64
		Name        string
		Description string
		Schedule    float64
		CreateTime  int64 `gorm:"column:create_time"`
		Private     int
		Archive     int
	}
	var pr projRow
	if err := db.Where("id=?", projectId).First(&pr).Error; err != nil {
		result, _ := json.Marshal(map[string]string{"error": "项目不存在"})
		return string(result)
	}

	// 获取负责人
	var ownerName string
	db.Table("ms_project_member pm").
		Joins("join ms_member m on m.id=pm.member_code").
		Where("pm.project_code=? and pm.is_owner=1", projectId).
		Select("m.name").Scan(&ownerName)

	// 获取统计
	var totalTasks, doneTasks, memberCount int64
	db.Table("ms_task").Where("project_code=? and deleted=0", projectId).Count(&totalTasks)
	db.Table("ms_task").Where("project_code=? and deleted=0 and done=1", projectId).Count(&doneTasks)
	db.Table("ms_project_member").Where("project_code=?", projectId).Count(&memberCount)

	completionRate := 0.0
	if totalTasks > 0 {
		completionRate = math.Round(float64(doneTasks)/float64(totalTasks)*10000) / 100
	}

	result, _ := json.Marshal(map[string]interface{}{
		"code":           codecs.EncryptInt64(pr.Id),
		"name":           pr.Name,
		"description":    pr.Description,
		"schedule":       pr.Schedule,
		"create_time":    pr.CreateTime,
		"owner_name":     ownerName,
		"total_tasks":    totalTasks,
		"done_tasks":     doneTasks,
		"member_count":   memberCount,
		"completion_rate": completionRate,
		"is_archived":    pr.Archive == 1,
	})
	return string(result)
}

// execCreateProject 创建项目
func execCreateProject(db *gorm.DB, memberId int64, orgCode int64, args map[string]interface{}) string {
	name, _ := args["name"].(string)
	if name == "" {
		result, _ := json.Marshal(map[string]string{"error": "项目名称不能为空"})
		return string(result)
	}
	description, _ := args["description"].(string)

	now := time.Now().UnixMilli()
	row := map[string]interface{}{
		"name":              name,
		"description":       description,
		"cover":             "",
		"private":           0,
		"deleted":           0,
		"archive":           0,
		"schedule":          0,
		"create_time":       now,
		"organization_code": orgCode,
	}
	if err := db.Table("ms_project").Create(row).Error; err != nil {
		result, _ := json.Marshal(map[string]string{"error": "创建项目失败: " + err.Error()})
		return string(result)
	}

	// 获取自增ID
	var projectId int64
	db.Table("ms_project").Where("name=? and create_time=?", name, now).Select("id").Scan(&projectId)

	// 添加创建者为项目 owner
	_ = db.Table("ms_project_member").Create(map[string]interface{}{
		"project_code": projectId,
		"member_code":  memberId,
		"join_time":    now,
		"is_owner":     1,
	}).Error

	// 创建默认任务阶段
	stages := []string{"待处理", "进行中", "已完成"}
	for i, stageName := range stages {
		_ = db.Table("ms_task_stages").Create(map[string]interface{}{
			"project_code": projectId,
			"name":         stageName,
			"sort":         i + 1,
			"create_time":  now,
			"deleted":      0,
		}).Error
	}

	result, _ := json.Marshal(map[string]interface{}{
		"code":    codecs.EncryptInt64(projectId),
		"name":    name,
		"message": "项目创建成功！已自动创建3个默认任务阶段：待处理、进行中、已完成",
	})
	return string(result)
}

// execEditProject 编辑项目
func execEditProject(db *gorm.DB, memberId int64, args map[string]interface{}) string {
	projectCode, _ := args["project_code"].(string)
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		result, _ := json.Marshal(map[string]string{"error": "project_code 无效"})
		return string(result)
	}
	if !authz.IsProjectMember(db, memberId, projectId) {
		result, _ := json.Marshal(map[string]string{"error": "无权限操作此项目"})
		return string(result)
	}

	updates := map[string]interface{}{}
	if v, ok := args["name"].(string); ok && v != "" {
		updates["name"] = v
	}
	if v, ok := args["description"].(string); ok {
		updates["description"] = v
	}
	if len(updates) == 0 {
		result, _ := json.Marshal(map[string]string{"error": "没有需要更新的字段"})
		return string(result)
	}
	_ = db.Table("ms_project").Where("id=?", projectId).Updates(updates).Error

	result, _ := json.Marshal(map[string]interface{}{
		"code":    codecs.EncryptInt64(projectId),
		"message": "项目更新成功",
	})
	return string(result)
}

// execListTasks 获取任务列表
func execListTasks(db *gorm.DB, memberId int64, args map[string]interface{}) string {
	projectCode, _ := args["project_code"].(string)
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		result, _ := json.Marshal(map[string]string{"error": "project_code 无效"})
		return string(result)
	}
	if !authz.IsProjectMember(db, memberId, projectId) {
		result, _ := json.Marshal(map[string]string{"error": "无权限访问此项目"})
		return string(result)
	}

	page := 1
	if v, ok := args["page"].(float64); ok && v > 0 {
		page = int(v)
	}
	pageSize := 20
	if v, ok := args["page_size"].(float64); ok && v > 0 {
		pageSize = int(v)
	}

	var total int64
	db.Table("ms_task").Where("project_code=? and deleted=0", projectId).Count(&total)

	type taskRow struct {
		Id          int64
		Name        string
		Description string
		Priority    int8
		Done        int8
		Status      int8
		CreateTime  int64  `gorm:"column:create_time"`
		AssignTo    int64  `gorm:"column:assign_to"`
		StageCode   int64  `gorm:"column:stage_code"`
		EndTime     int64  `gorm:"column:end_time"`
	}
	var rows []taskRow
	db.Table("ms_task").
		Where("project_code=? and deleted=0", projectId).
		Order("id desc").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows)

	type taskItem struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Priority    string `json:"priority_text"`
		Done        bool   `json:"done"`
		Status      string `json:"status_text"`
		AssignName  string `json:"assign_name"`
		StageName   string `json:"stage_name"`
		EndTime     int64  `json:"end_time"`
		CreateTime  int64  `json:"create_time"`
	}
	list := make([]taskItem, 0, len(rows))
	for _, t := range rows {
		priText := "普通"
		if t.Priority == 1 {
			priText = "紧急"
		} else if t.Priority >= 2 {
			priText = "非常紧急"
		}
		statusText := "未开始"
		switch t.Status {
		case 1:
			statusText = "已完成"
		case 2:
			statusText = "进行中"
		case 3:
			statusText = "挂起"
		case 4:
			statusText = "测试中"
		}
		var assignName string
		if t.AssignTo != 0 {
			db.Table("ms_member").Where("id=?", t.AssignTo).Select("name").Scan(&assignName)
		}
		var stageName string
		if t.StageCode != 0 {
			db.Table("ms_task_stages").Where("id=?", t.StageCode).Select("name").Scan(&stageName)
		}
		list = append(list, taskItem{
			Code:        codecs.EncryptInt64(t.Id),
			Name:        t.Name,
			Description: t.Description,
			Priority:    priText,
			Done:        t.Done == 1,
			Status:      statusText,
			AssignName:  assignName,
			StageName:   stageName,
			EndTime:     t.EndTime,
			CreateTime:  t.CreateTime,
		})
	}

	result, _ := json.Marshal(map[string]interface{}{
		"total": total,
		"page":  page,
		"list":  list,
	})
	return string(result)
}

// execGetMyTasks 获取我的任务
func execGetMyTasks(db *gorm.DB, memberId int64, orgCode int64, args map[string]interface{}) string {
	page := 1
	if v, ok := args["page"].(float64); ok && v > 0 {
		page = int(v)
	}
	pageSize := 20
	if v, ok := args["page_size"].(float64); ok && v > 0 {
		pageSize = int(v)
	}

	query := db.Table("ms_task t").
		Where("t.deleted=0").
		Where("t.assign_to=? or t.owner_code=? or t.member_code=?", memberId, memberId, memberId)
	if orgCode != 0 {
		subQuery := db.Table("ms_project").Select("id").Where("organization_code=? AND deleted=0", orgCode)
		query = query.Where("t.project_code IN (?)", subQuery)
	}
	if v, ok := args["done"].(string); ok && (v == "0" || v == "1") {
		d, _ := strconv.ParseInt(v, 10, 8)
		query = query.Where("t.done=?", d)
	}

	var total int64
	_ = query.Count(&total).Error

	type taskRow struct {
		Id          int64
		Name        string
		Priority    int8
		Done        int8
		EndTime     int64 `gorm:"column:end_time"`
		ProjectCode int64 `gorm:"column:project_code"`
	}
	var rows []taskRow
	_ = query.Select("t.id, t.name, t.priority, t.done, t.end_time, t.project_code").
		Order("t.id desc").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error

	type taskItem struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Priority    string `json:"priority_text"`
		Done        bool   `json:"done"`
		EndTime     int64  `json:"end_time"`
		ProjectName string `json:"project_name"`
		ProjectCode string `json:"project_code"`
	}
	list := make([]taskItem, 0, len(rows))
	for _, t := range rows {
		priText := "普通"
		if t.Priority == 1 {
			priText = "紧急"
		} else if t.Priority >= 2 {
			priText = "非常紧急"
		}
		var projectName string
		db.Table("ms_project").Where("id=?", t.ProjectCode).Select("name").Scan(&projectName)
		list = append(list, taskItem{
			Code:        codecs.EncryptInt64(t.Id),
			Name:        t.Name,
			Priority:    priText,
			Done:        t.Done == 1,
			EndTime:     t.EndTime,
			ProjectName: projectName,
			ProjectCode: codecs.EncryptInt64(t.ProjectCode),
		})
	}

	result, _ := json.Marshal(map[string]interface{}{
		"total": total,
		"page":  page,
		"list":  list,
	})
	return string(result)
}

// execCreateTask 创建任务
func execCreateTask(db *gorm.DB, memberId int64, args map[string]interface{}) string {
	projectCode, _ := args["project_code"].(string)
	name, _ := args["name"].(string)
	if projectCode == "" || name == "" {
		result, _ := json.Marshal(map[string]string{"error": "项目ID和任务名称不能为空"})
		return string(result)
	}
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil || pid == 0 {
		result, _ := json.Marshal(map[string]string{"error": "project_code 无效"})
		return string(result)
	}
	if !authz.IsProjectMember(db, memberId, pid) {
		result, _ := json.Marshal(map[string]string{"error": "无权限操作此项目"})
		return string(result)
	}

	// 获取第一个阶段作为默认阶段
	var stageId int64
	db.Table("ms_task_stages").Where("project_code=? and deleted=0", pid).
		Order("sort asc, id asc").Limit(1).Select("id").Scan(&stageId)
	if stageId == 0 {
		result, _ := json.Marshal(map[string]string{"error": "项目没有任务阶段，请先在项目页面创建看板阶段"})
		return string(result)
	}

	priority := int8(0)
	if v, ok := args["priority"].(string); ok {
		if p, err := strconv.ParseInt(v, 10, 8); err == nil {
			priority = int8(p)
		}
	}
	description, _ := args["description"].(string)

	var maxSort int
	_ = db.Table("ms_task").Where("stage_code=? and deleted=0", stageId).
		Select("coalesce(max(sort),0)").Scan(&maxSort).Error

	row := map[string]interface{}{
		"project_code": pid,
		"name":         name,
		"description":  description,
		"priority":     priority,
		"create_time":  time.Now().UnixMilli(),
		"member_code":  memberId,
		"owner_code":   memberId,
		"assign_to":    0,
		"stage_code":   stageId,
		"sort":         maxSort + 1,
		"deleted":      0,
		"private":      0,
		"done":         0,
	}
	if err := db.Table("ms_task").Create(row).Error; err != nil {
		result, _ := json.Marshal(map[string]string{"error": "创建任务失败: " + err.Error()})
		return string(result)
	}

	// 获取自增ID
	var taskId int64
	db.Table("ms_task").Where("project_code=? and name=? and member_code=?", pid, name, memberId).
		Order("id desc").Limit(1).Select("id").Scan(&taskId)

	result, _ := json.Marshal(map[string]interface{}{
		"code":    codecs.EncryptInt64(taskId),
		"name":    name,
		"message": "任务创建成功！",
	})
	return string(result)
}

// execEditTask 编辑任务
func execEditTask(db *gorm.DB, memberId int64, args map[string]interface{}) string {
	taskCode, _ := args["task_code"].(string)
	taskId, err := codecs.DecryptInt64(taskCode)
	if err != nil || taskId == 0 {
		result, _ := json.Marshal(map[string]string{"error": "task_code 无效"})
		return string(result)
	}

	// 查询任务所属项目
	var projectId int64
	db.Table("ms_task").Where("id=?", taskId).Select("project_code").Scan(&projectId)
	if projectId == 0 {
		result, _ := json.Marshal(map[string]string{"error": "任务不存在"})
		return string(result)
	}
	if !authz.IsProjectMember(db, memberId, projectId) {
		result, _ := json.Marshal(map[string]string{"error": "无权限操作此任务"})
		return string(result)
	}

	updates := map[string]interface{}{}
	if v, ok := args["name"].(string); ok && v != "" {
		updates["name"] = v
	}
	if v, ok := args["description"].(string); ok {
		updates["description"] = v
	}
	if v, ok := args["priority"].(string); ok {
		if p, err := strconv.ParseInt(v, 10, 8); err == nil {
			updates["priority"] = int8(p)
		}
	}
	if v, ok := args["status"].(string); ok {
		if s, err := strconv.ParseInt(v, 10, 8); err == nil {
			updates["status"] = int8(s)
		}
	}
	if len(updates) == 0 {
		result, _ := json.Marshal(map[string]string{"error": "没有需要更新的字段"})
		return string(result)
	}
	_ = db.Table("ms_task").Where("id=?", taskId).Updates(updates).Error

	result, _ := json.Marshal(map[string]interface{}{
		"code":    codecs.EncryptInt64(taskId),
		"message": "任务更新成功",
	})
	return string(result)
}

// execCompleteTask 完成/重做任务
func execCompleteTask(db *gorm.DB, memberId int64, args map[string]interface{}) string {
	taskCode, _ := args["task_code"].(string)
	doneStr, _ := args["done"].(string)
	taskId, err := codecs.DecryptInt64(taskCode)
	if err != nil || taskId == 0 {
		result, _ := json.Marshal(map[string]string{"error": "task_code 无效"})
		return string(result)
	}

	var projectId int64
	db.Table("ms_task").Where("id=?", taskId).Select("project_code").Scan(&projectId)
	if projectId == 0 {
		result, _ := json.Marshal(map[string]string{"error": "任务不存在"})
		return string(result)
	}
	if !authz.IsProjectMember(db, memberId, projectId) {
		result, _ := json.Marshal(map[string]string{"error": "无权限操作此任务"})
		return string(result)
	}

	done := int8(0)
	if doneStr == "1" {
		done = 1
	}
	updates := map[string]interface{}{"done": done}
	if done == 1 {
		updates["done_time"] = time.Now().UnixMilli()
	} else {
		updates["done_time"] = int64(0)
	}
	_ = db.Table("ms_task").Where("id=?", taskId).Updates(updates).Error

	action := "已标记为完成"
	if done == 0 {
		action = "已重做（取消完成）"
	}
	result, _ := json.Marshal(map[string]interface{}{
		"code":    codecs.EncryptInt64(taskId),
		"message": action,
	})
	return string(result)
}

// execGetProjectStats 获取项目统计
func execGetProjectStats(db *gorm.DB, memberId int64, orgCode int64, args map[string]interface{}) string {
	projectCode, _ := args["project_code"].(string)
	if projectCode != "" {
		projectId, err := codecs.DecryptInt64(projectCode)
		if err != nil || projectId == 0 {
			result, _ := json.Marshal(map[string]string{"error": "project_code 无效"})
			return string(result)
		}
		if !authz.IsProjectMember(db, memberId, projectId) {
			result, _ := json.Marshal(map[string]string{"error": "无权限访问此项目"})
			return string(result)
		}
		var totalTasks, doneTasks, overdueTasks, memberCount int64
		now := time.Now().UnixMilli()
		db.Table("ms_task").Where("project_code=? and deleted=0", projectId).Count(&totalTasks)
		db.Table("ms_task").Where("project_code=? and deleted=0 and done=1", projectId).Count(&doneTasks)
		db.Table("ms_task").Where("project_code=? and deleted=0 and done=0 and end_time>0 and end_time<?", projectId, now).Count(&overdueTasks)
		db.Table("ms_project_member").Where("project_code=?", projectId).Count(&memberCount)
		completionRate := 0.0
		if totalTasks > 0 {
			completionRate = math.Round(float64(doneTasks)/float64(totalTasks)*10000) / 100
		}
		result, _ := json.Marshal(map[string]interface{}{
			"project_code":    projectCode,
			"total_tasks":     totalTasks,
			"done_tasks":      doneTasks,
			"undone_tasks":    totalTasks - doneTasks,
			"overdue_tasks":   overdueTasks,
			"member_count":    memberCount,
			"completion_rate": completionRate,
		})
		return string(result)
	}

	// 全局统计
	var projectCount, taskCount, doneCount, overdueCount int64
	projectQuery := db.Table("ms_project_member pm").
		Joins("join ms_project p on p.id=pm.project_code").
		Where("pm.member_code=? and p.deleted=0", memberId)
	if orgCode != 0 {
		projectQuery = projectQuery.Where("p.organization_code=?", orgCode)
	}
	projectQuery.Count(&projectCount)

	taskQuery := db.Table("ms_task t").
		Joins("join ms_project_member pm on pm.project_code=t.project_code").
		Joins("join ms_project p on p.id=pm.project_code").
		Where("pm.member_code=? and t.deleted=0", memberId)
	if orgCode != 0 {
		taskQuery = taskQuery.Where("p.organization_code=?", orgCode)
	}
	taskQuery.Count(&taskCount)

	doneQuery := db.Table("ms_task t").
		Joins("join ms_project_member pm on pm.project_code=t.project_code").
		Joins("join ms_project p on p.id=pm.project_code").
		Where("pm.member_code=? and t.deleted=0 and t.done=1", memberId)
	if orgCode != 0 {
		doneQuery = doneQuery.Where("p.organization_code=?", orgCode)
	}
	doneQuery.Count(&doneCount)

	now := time.Now().UnixMilli()
	overdueQuery := db.Table("ms_task t").
		Joins("join ms_project_member pm on pm.project_code=t.project_code").
		Joins("join ms_project p on p.id=pm.project_code").
		Where("pm.member_code=? and t.deleted=0 and t.done=0 and t.end_time>0 and t.end_time<?", memberId, now)
	if orgCode != 0 {
		overdueQuery = overdueQuery.Where("p.organization_code=?", orgCode)
	}
	overdueQuery.Count(&overdueCount)

	completionRate := 0.0
	if taskCount > 0 {
		completionRate = math.Round(float64(doneCount)/float64(taskCount)*10000) / 100
	}

	result, _ := json.Marshal(map[string]interface{}{
		"project_count":   projectCount,
		"total_tasks":     taskCount,
		"done_tasks":      doneCount,
		"undone_tasks":    taskCount - doneCount,
		"overdue_tasks":   overdueCount,
		"completion_rate": completionRate,
	})
	return string(result)
}

// execListProjectMembers 获取项目成员
func execListProjectMembers(db *gorm.DB, memberId int64, args map[string]interface{}) string {
	projectCode, _ := args["project_code"].(string)
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		result, _ := json.Marshal(map[string]string{"error": "project_code 无效"})
		return string(result)
	}
	if !authz.IsProjectMember(db, memberId, projectId) {
		result, _ := json.Marshal(map[string]string{"error": "无权限访问此项目"})
		return string(result)
	}

	type memberRow struct {
		Id     int64  `gorm:"column:id"`
		Name   string `gorm:"column:name"`
		Avatar string `gorm:"column:avatar"`
		Email  string `gorm:"column:email"`
		IsOwner int   `gorm:"column:is_owner"`
	}
	var rows []memberRow
	db.Table("ms_project_member pm").
		Joins("join ms_member m on m.id=pm.member_code").
		Where("pm.project_code=?", projectId).
		Select("m.id, m.name, m.avatar, m.email, pm.is_owner").
		Order("pm.is_owner desc, pm.id asc").
		Scan(&rows)

	type memberItem struct {
		Code    string `json:"code"`
		Name    string `json:"name"`
		Avatar  string `json:"avatar"`
		Email   string `json:"email"`
		IsOwner bool   `json:"is_owner"`
	}
	list := make([]memberItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, memberItem{
			Code:    codecs.EncryptInt64(r.Id),
			Name:    r.Name,
			Avatar:  r.Avatar,
			Email:   r.Email,
			IsOwner: r.IsOwner == 1,
		})
	}

	result, _ := json.Marshal(map[string]interface{}{
		"total": len(list),
		"list":  list,
	})
	return string(result)
}

// ==================== 系统提示词 ====================

func getSystemPrompt(memberName string) string {
	return fmt.Sprintf(`你是项目管理系统的 AI 助手。你的名字叫"小助手"，你可以帮助用户管理项目和任务。

当前用户：%s

你可以执行以下操作：
1. **项目管理**：查看项目列表、项目详情、创建项目、编辑项目、查看项目统计和成员
2. **任务管理**：查看项目任务、查看我的任务、创建任务、编辑任务、完成任务
3. **数据分析**：查看项目统计数据、完成率、逾期任务等

使用说明：
- 当用户说"我的项目"或"查看项目"时，调用 list_projects 工具
- 当用户说"创建项目"时，调用 create_project 工具
- 当用户说"创建任务"时，需要先知道在哪个项目下创建，如果用户没有指定项目，请先调用 list_projects 让用户选择
- 任务优先级：0=普通, 1=紧急, 2=非常紧急
- 任务状态：0=未开始, 1=已完成, 2=进行中, 3=挂起, 4=测试中

重要规则：
1. 在执行写操作（创建、编辑、完成）前，先用自然语言确认用户的意图
2. 对于删除操作，务必再次确认
3. 返回数据时，用简洁清晰的方式展示，不要直接返回原始 JSON
4. 如果工具返回错误，友好地告知用户并建议解决方案
5. 保持回复简洁，重点突出关键信息`, memberName)
}

// ==================== 核心聊天逻辑 ====================

func (h *HandlerAssistant) chat(c *gin.Context) {
	msgs, err := parseMessages(c)
	if err != nil || len(msgs) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	memberName, _ := c.Get("memberName")
	memberNameStr := "用户"
	if v, ok := memberName.(string); ok && v != "" {
		memberNameStr = v
	}

	// 构建消息列表（加入系统提示）
	messages := make([]chatMessage, 0, len(msgs)+1)
	messages = append(messages, chatMessage{
		Role:    "system",
		Content: getSystemPrompt(memberNameStr),
	})
	// 只保留 user 和 assistant 的消息
	for _, m := range msgs {
		if m.Role == "user" || m.Role == "assistant" {
			messages = append(messages, chatMessage{
				Role:    m.Role,
				Content: m.Content,
			})
		}
	}

	// 构建请求体
	body := map[string]interface{}{
		"model":    "openclaw",
		"messages": messages,
		"tools":    getToolDefinitions(),
		"stream":   false,
	}

	// Function calling 循环
	for round := 0; round < maxToolRounds; round++ {
		bodyBytes, _ := json.Marshal(body)
		apiURL := fmt.Sprintf("%s/v1/chat/completions", openclawBaseURL)

		httpReq, _ := http.NewRequest("POST", apiURL, bytes.NewReader(bodyBytes))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+openclawToken)

		httpClient := &http.Client{Timeout: 120 * time.Second}
		resp, err := httpClient.Do(httpReq)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "AI 服务连接失败或超时"})
			return
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result map[string]interface{}
		if err := json.Unmarshal(respBody, &result); err != nil {
			log.Printf("[assistant] JSON unmarshal error, body preview: %.200s", string(respBody))
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "AI 响应解析失败"})
			return
		}

		choices, _ := result["choices"].([]interface{})
		if len(choices) == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 500, "msg": "AI 无响应"})
			return
		}

		choice, _ := choices[0].(map[string]interface{})
		message, _ := choice["message"].(map[string]interface{})

		// 检查是否有工具调用
		toolCallsRaw, hasToolCalls := message["tool_calls"].([]interface{})
		if !hasToolCalls || len(toolCallsRaw) == 0 {
			// 没有工具调用，直接返回 AI 的文本回复
			content, _ := message["content"].(string)
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"data": map[string]interface{}{
					"choices": []interface{}{
						map[string]interface{}{
							"message": map[string]interface{}{
								"role":    "assistant",
								"content": content,
							},
						},
					},
				},
				"msg": "ok",
			})
			return
		}

		// 有工具调用，执行工具并继续循环
		// 先将 assistant 的消息（含 tool_calls）加入历史
		assistantMsg := chatMessage{
			Role:      "assistant",
			Content:   "",
			ToolCalls: make([]toolCall, 0),
		}
		if content, ok := message["content"].(string); ok {
			assistantMsg.Content = content
		}
		for _, tcRaw := range toolCallsRaw {
			tc, _ := tcRaw.(map[string]interface{})
			fn, _ := tc["function"].(map[string]interface{})
			args, _ := fn["arguments"].(string)
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, toolCall{
				ID:   fmt.Sprintf("%v", tc["id"]),
				Type: fmt.Sprintf("%v", tc["type"]),
				Function: toolFunction{
					Name:      fmt.Sprintf("%v", fn["name"]),
					Arguments: args,
				},
			})
		}
		messages = append(messages, assistantMsg)

		// 执行每个工具调用
		for _, tc := range assistantMsg.ToolCalls {
			var args map[string]interface{}
			json.Unmarshal([]byte(tc.Function.Arguments), &args)

			log.Printf("[AI Assistant] 执行工具: %s, 参数: %s", tc.Function.Name, tc.Function.Arguments)

			toolResult := executeTool(c, tc.Function.Name, args)

			log.Printf("[AI Assistant] 工具结果: %s", toolResult)

			messages = append(messages, chatMessage{
				Role:       "tool",
				Content:    toolResult,
				ToolCallID: tc.ID,
			})
		}

		// 更新请求体中的 messages
		body["messages"] = messages
	}

	// 超过最大轮数，返回提示
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "抱歉，操作过于复杂，已达到最大处理轮数。请尝试简化你的请求。",
					},
				},
			},
		},
		"msg": "ok",
	})
}
