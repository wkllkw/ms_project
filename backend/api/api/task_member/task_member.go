package task_member

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	common "test.com/project-common"
)

type HandlerTaskMember struct{}

func New() *HandlerTaskMember {
	return &HandlerTaskMember{}
}

type taskMemberRow struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	TaskId     int64 `gorm:"column:task_id"`
	MemberId   int64 `gorm:"column:member_id"`
	IsExecutor int8  `gorm:"column:is_executor"`
	IsOwner    int8  `gorm:"column:is_owner"`
	JoinTime   int64 `gorm:"column:join_time"`
}

func (*taskMemberRow) TableName() string { return "ms_task_member" }

type memberRow struct {
	Id     int64  `gorm:"primaryKey;autoIncrement"`
	Name   string `gorm:"column:name"`
	Avatar string `gorm:"column:avatar"`
	Email  string `gorm:"column:email"`
}

func (*memberRow) TableName() string { return "ms_member" }

type memberAccountRow struct {
	Id             int64  `gorm:"primaryKey;autoIncrement"`
	MemberCode     int64  `gorm:"column:member_code"`
	OrganizationCode int64 `gorm:"column:organization_code"`
	Name           string `gorm:"column:name"`
	Avatar         string `gorm:"column:avatar"`
	Email          string `gorm:"column:email"`
	Status         int8   `gorm:"column:status"`
	Authorize      string `gorm:"column:authorize"`
	Code           string `gorm:"column:code"`
}

func (*memberAccountRow) TableName() string { return "ms_member_account" }

type taskRow struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"column:project_code"`
	Deleted     int8  `gorm:"column:deleted"`
}

func (*taskRow) TableName() string { return "ms_task" }

type projectMemberRow struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"column:project_code"`
	MemberCode  int64 `gorm:"column:member_code"`
}

func (*projectMemberRow) TableName() string { return "ms_project_member" }

// list 获取任务成员列表
func (h *HandlerTaskMember) list(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")

	if taskCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode必填"))
		return
	}

	tid, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}

	db := gorms.GetDB()
	var members []taskMemberRow
	err = db.Where("task_id = ?", tid).Order("is_owner DESC, is_executor DESC").Find(&members).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "查询失败"))
		return
	}

	orgCode := c.GetInt64("organizationCode")
	list := make([]gin.H, 0, len(members))

	for _, m := range members {
		// 获取成员基本信息
		var member memberRow
		db.Where("id = ?", m.MemberId).First(&member)

		// 获取成员账户信息
		var memberAccount memberAccountRow
		db.Where("member_code = ? AND organization_code = ?", m.MemberId, orgCode).First(&memberAccount)

		list = append(list, gin.H{
			"memberCode":       codecs.EncryptInt64(member.Id),
			"name":             member.Name,
			"avatar":           member.Avatar,
			"memberAccountCode": memberAccount.Code,
			"status":           memberAccount.Status,
			"authorize":        memberAccount.Authorize,
			"isExecutor":       m.IsExecutor,
			"isOwner":          m.IsOwner,
		})
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"list": list}))
}

// searchInviteMember 搜索可邀请的成员
func (h *HandlerTaskMember) searchInviteMember(c *gin.Context) {
	result := &common.Result{}
	keyword := c.PostForm("keyword")
	projectCode := c.PostForm("projectCode")

	if projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "请先选择项目"))
		return
	}

	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}

	if keyword == "" {
		c.JSON(http.StatusOK, result.Success([]gin.H{}))
		return
	}

	db := gorms.GetDB()
	orgCode := c.GetInt64("organizationCode")

	// 获取项目成员
	var projectMembers []projectMemberRow
	db.Where("project_code = ?", pid).Find(&projectMembers)
	projectMemberMap := make(map[int64]bool)
	for _, pm := range projectMembers {
		projectMemberMap[pm.MemberCode] = true
	}

	// 从当前组织查询成员
	var memberAccounts []memberAccountRow
	db.Where("name LIKE ? AND organization_code = ?", "%"+keyword+"%", orgCode).Find(&memberAccounts)

	tempList := make(map[int64]gin.H)

	for _, ma := range memberAccounts {
		tempList[ma.MemberCode] = gin.H{
			"memberCode": codecs.EncryptInt64(ma.MemberCode),
			"avatar":     ma.Avatar,
			"name":       ma.Name,
			"email":      ma.Email,
			"joined":     projectMemberMap[ma.MemberCode],
		}
	}

	// 从平台查询（按邮箱）
	var members []memberRow
	db.Where("email LIKE ?", "%"+keyword+"%").Find(&members)
	for _, m := range members {
		if _, exists := tempList[m.Id]; !exists {
			tempList[m.Id] = gin.H{
				"memberCode": codecs.EncryptInt64(m.Id),
				"avatar":     m.Avatar,
				"name":       m.Name,
				"email":      m.Email,
				"joined":     projectMemberMap[m.Id],
			}
		}
	}

	// 转换为数组
	list := make([]gin.H, 0, len(tempList))
	for _, item := range tempList {
		list = append(list, item)
	}

	c.JSON(http.StatusOK, result.Success(list))
}

// inviteMember 邀请成员加入任务
func (h *HandlerTaskMember) inviteMember(c *gin.Context) {
	result := &common.Result{}
	memberCode := c.PostForm("memberCode")
	taskCode := c.PostForm("taskCode")

	if memberCode == "" || taskCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "数据异常"))
		return
	}

	mid, err := codecs.DecryptInt64(memberCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "memberCode无效"))
		return
	}

	tid, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}

	db := gorms.GetDB()

	// 检查任务是否存在
	var task taskRow
	err = db.Where("id = ? AND deleted = 0", tid).First(&task).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "该任务已失效"))
		return
	}

	// 检查是否已加入
	var existMember taskMemberRow
	err = db.Where("task_id = ? AND member_id = ?", tid, mid).First(&existMember).Error
	if err == nil {
		c.JSON(http.StatusOK, result.Success(nil))
		return
	}

	// 添加成员
	taskMember := &taskMemberRow{
		TaskId:     tid,
		MemberId:   mid,
		IsExecutor: 0,
		IsOwner:    0,
		JoinTime:   time.Now().UnixMilli(),
	}

	if err := db.Create(taskMember).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "邀请失败"))
		return
	}

	// 同时加入项目
	var projectMember projectMemberRow
	err = db.Where("project_code = ? AND member_code = ?", task.ProjectCode, mid).First(&projectMember).Error
	if err == gorm.ErrRecordNotFound {
		projectMember = projectMemberRow{
			ProjectCode: task.ProjectCode,
			MemberCode:  mid,
		}
		db.Create(&projectMember)
	}

	c.JSON(http.StatusOK, result.Success(nil))
}

// inviteMemberBatch 批量邀请成员
func (h *HandlerTaskMember) inviteMemberBatch(c *gin.Context) {
	result := &common.Result{}
	memberCodesJSON := c.PostForm("memberCodes")
	taskCode := c.PostForm("taskCode")

	if memberCodesJSON == "" || taskCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "数据异常"))
		return
	}

	tid, err := codecs.DecryptInt64(taskCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}

	// 解析 memberCodes JSON 数组
	memberCodeStrs := []string{}
	if err := json.Unmarshal([]byte(memberCodesJSON), &memberCodeStrs); err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "memberCodes格式错误"))
		return
	}

	db := gorms.GetDB()

	// 检查任务是否存在
	var task taskRow
	err = db.Where("id = ? AND deleted = 0", tid).First(&task).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "该任务已失效"))
		return
	}

	// 检查是否包含 "all"
	isAll := false
	for _, code := range memberCodeStrs {
		if code == "all" {
			isAll = true
			break
		}
	}

	if isAll {
		// 获取所有项目成员
		var projectMembers []projectMemberRow
		db.Where("project_code = ?", task.ProjectCode).Find(&projectMembers)
		memberCodeStrs = []string{}
		for _, pm := range projectMembers {
			memberCodeStrs = append(memberCodeStrs, codecs.EncryptInt64(pm.MemberCode))
		}
	}

	// 获取任务创建者
	var owner taskMemberRow
	db.Where("task_id = ? AND is_owner = 1", tid).First(&owner)

	// 批量处理
	for _, memberCodeStr := range memberCodeStrs {
		mid, err := codecs.DecryptInt64(memberCodeStr)
		if err != nil {
			continue
		}

		// 创建者不能被移除
		if owner.MemberId == mid {
			continue
		}

		var existMember taskMemberRow
		err = db.Where("task_id = ? AND member_id = ?", tid, mid).First(&existMember).Error

		if err == gorm.ErrRecordNotFound {
			// 不存在，添加
			taskMember := &taskMemberRow{
				TaskId:     tid,
				MemberId:   mid,
				IsExecutor: 0,
				IsOwner:    0,
				JoinTime:   time.Now().UnixMilli(),
			}
			db.Create(taskMember)

			// 同时加入项目
			var projectMember projectMemberRow
			err = db.Where("project_code = ? AND member_code = ?", task.ProjectCode, mid).First(&projectMember).Error
			if err == gorm.ErrRecordNotFound {
				projectMember = projectMemberRow{
					ProjectCode: task.ProjectCode,
					MemberCode:  mid,
				}
				db.Create(&projectMember)
			}
		} else if err == nil && !isAll {
			// 已存在且不是"全部成员"模式，则移除
			db.Where("task_id = ? AND member_id = ?", tid, mid).Delete(&taskMemberRow{})
		}
	}

	c.JSON(http.StatusOK, result.Success(nil))
}
