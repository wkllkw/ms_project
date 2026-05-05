package invite_link

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	common "test.com/project-common"
)

type HandlerInviteLink struct{}

func New() *HandlerInviteLink {
	return &HandlerInviteLink{}
}

// inviteLinkRow 邀请链接表结构
type inviteLinkRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	ProjectId   int64  `gorm:"column:project_id"`
	InviteCode  string `gorm:"column:invite_code;unique"`
	ExpiredAt   int64  `gorm:"column:expired_at"`
	CreateBy    int64  `gorm:"column:create_by"`
	CreateTime  int64  `gorm:"column:create_time"`
	InviteType  string `gorm:"column:invite_type"`
	SourceCode  int64  `gorm:"column:source_code"`
}

func (*inviteLinkRow) TableName() string {
	return "ms_invite_link"
}

// projectRow 项目表结构（用于查询）
type projectRow struct {
	Id               int64  `gorm:"primaryKey;autoIncrement"`
	Name             string `gorm:"column:name"`
	OrganizationCode int64  `gorm:"column:organization_code"`
}

func (*projectRow) TableName() string {
	return "ms_project"
}

// organizationRow 组织表结构（用于查询）
type organizationRow struct {
	Id   int64  `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"column:name"`
}

func (*organizationRow) TableName() string {
	return "ms_organization"
}

// memberRow 成员表结构（用于查询）
type memberRow struct {
	Id     int64  `gorm:"primaryKey;autoIncrement"`
	Name   string `gorm:"column:name"`
	Avatar string `gorm:"column:avatar"`
	Email  string `gorm:"column:email"`
}

func (*memberRow) TableName() string {
	return "ms_member"
}

// generateInviteCode 生成邀请码
func generateInviteCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// createInvite 创建邀请链接
func (h *HandlerInviteLink) createInvite(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	inviteType := c.PostForm("inviteType") // project 或 organization
	sourceCode := c.PostForm("sourceCode")

	if inviteType == "" || sourceCode == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数不完整"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 检查是否已存在有效的邀请链接
	var existingLink inviteLinkRow
	err := db.Table("ms_invite_link").
		Where("invite_type = ? AND source_code = ? AND create_by = ? AND expired_at > ?", 
			inviteType, sourceCode, memberId, time.Now().UnixMilli()).
		First(&existingLink).Error
	
	if err == nil {
		// 已存在有效链接，直接返回
		c.JSON(http.StatusOK, result.Success(existingLink))
		return
	}

	// 验证资源是否存在
	if inviteType == "project" {
		var project projectRow
		if err := db.Table("ms_project").Where("id = ?", sourceCode).First(&project).Error; err != nil {
			c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "项目不存在"))
			return
		}
	} else if inviteType == "organization" {
		var org organizationRow
		if err := db.Table("ms_organization").Where("id = ?", sourceCode).First(&org).Error; err != nil {
			c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "组织不存在"))
			return
		}
	} else {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "不支持的邀请类型"))
		return
	}

	// 创建新的邀请链接
	inviteCode := generateInviteCode()
	expiredAt := time.Now().Add(24 * time.Hour).UnixMilli() // 24小时后过期

	inviteLink := &inviteLinkRow{
		InviteCode: inviteCode,
		InviteType: inviteType,
		SourceCode: decryptOrParseInt64(sourceCode),
		ExpiredAt:  expiredAt,
		CreateBy:   memberId,
		CreateTime: time.Now().UnixMilli(),
	}

	if err := db.Table("ms_invite_link").Create(inviteLink).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "创建邀请链接失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(inviteLink))
}

// getInviteDetail 获取邀请详情
func (h *HandlerInviteLink) getInviteDetail(c *gin.Context) {
	result := &common.Result{}

	inviteCode := c.PostForm("inviteCode")
	if inviteCode == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "邀请码不能为空"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 查询邀请链接
	var inviteLink inviteLinkRow
	if err := db.Table("ms_invite_link").
		Where("invite_code = ?", inviteCode).
		First(&inviteLink).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "邀请链接不存在或已失效"))
		return
	}

	// 检查是否过期
	if inviteLink.ExpiredAt < time.Now().UnixMilli() {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "邀请链接已过期"))
		return
	}

	// 构建返回数据
	response := gin.H{
		"inviteCode": inviteLink.InviteCode,
		"inviteType": inviteLink.InviteType,
		"expiredAt":  inviteLink.ExpiredAt,
		"createTime": inviteLink.CreateTime,
	}

	// 获取邀请人信息
	var inviter memberRow
	if err := db.Table("ms_member").
		Where("id = ?", inviteLink.CreateBy).
		First(&inviter).Error; err == nil {
		response["member"] = gin.H{
			"id":     inviter.Id,
			"name":   inviter.Name,
			"avatar": inviter.Avatar,
			"email":  inviter.Email,
		}
	}

	// 获取资源详情
	if inviteLink.InviteType == "project" {
		var project projectRow
		if err := db.Table("ms_project").
			Where("id = ?", inviteLink.SourceCode).
			First(&project).Error; err == nil {
			response["name"] = project.Name
			response["sourceDetail"] = gin.H{
				"id":               project.Id,
				"name":             project.Name,
				"organizationCode": project.OrganizationCode,
			}
		}
	} else if inviteLink.InviteType == "organization" {
		var org organizationRow
		if err := db.Table("ms_organization").
			Where("id = ?", inviteLink.SourceCode).
			First(&org).Error; err == nil {
			response["name"] = org.Name
			response["sourceDetail"] = gin.H{
				"id":   org.Id,
				"name": org.Name,
			}
		}
	}

	c.JSON(http.StatusOK, result.Success(response))
}

// parseStringToInt64 辅助函数：字符串转int64
func parseStringToInt64(s string) int64 {
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		}
	}
	return result
}

// decryptOrParseInt64 尝试先解密，失败则提取数字
func decryptOrParseInt64(s string) int64 {
	decrypted, err := codecs.DecryptInt64(s)
	if err == nil && decrypted > 0 {
		return decrypted
	}
	return parseStringToInt64(s)
}
