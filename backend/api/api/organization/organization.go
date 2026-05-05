package organization

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/api/rpc"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	common "test.com/project-common"
	"test.com/project-grpc/user/login"
)

type HandlerOrganization struct{}

func New() *HandlerOrganization {
	return &HandlerOrganization{}
}

// organizationRow 组织表结构
type organizationRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"column:name"`
	Avatar      string `gorm:"column:avatar"`
	Description string `gorm:"column:description"`
	MemberId    int64  `gorm:"column:member_id"`
	CreateTime  int64  `gorm:"column:create_time"`
	Personal    int    `gorm:"column:personal"`
	Address     string `gorm:"column:address"`
	Province    int    `gorm:"column:province"`
	City        int    `gorm:"column:city"`
	Area        int    `gorm:"column:area"`
}

func (*organizationRow) TableName() string {
	return "ms_organization"
}

func (h *HandlerOrganization) loadOrganizations(memberId int64) ([]gin.H, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rsp, err := rpc.LoginServiceClient.MyOrgList(ctx, &login.UserMessage{MemId: memberId})
	if err != nil {
		return nil, err
	}

	orgList := make([]gin.H, 0, len(rsp.OrganizationList))
	for _, org := range rsp.OrganizationList {
		if org == nil {
			continue
		}
		orgList = append(orgList, gin.H{
			"id":          org.Code,
			"code":        org.Code,
			"name":        org.Name,
			"avatar":      org.Avatar,
			"description": org.Description,
			"owner_code":  org.OwnerCode,
			"create_time": org.CreateTime,
			"personal":    org.Personal,
			"address":     org.Address,
			"province":    org.Province,
			"city":        org.City,
			"area":        org.Area,
		})
	}

	return orgList, nil
}

// isOrgMember 检查用户是否是组织成员（创建者或通过部门加入）
func isOrgMember(db *gorms.GormConn, memberId int64, organizationCode int64) bool {
	// 检查是否是组织创建者
	var org organizationRow
	if err := db.Session(context.Background()).Where("id=? AND member_id=?", organizationCode, memberId).First(&org).Error; err == nil {
		return true
	}
	// 检查是否通过部门加入
	var count int64
	db.Session(context.Background()).
		Table("ms_department_member").
		Joins("JOIN ms_department ON ms_department.id = ms_department_member.department_id").
		Where("ms_department_member.member_id=? AND ms_department.organization_code=? AND ms_department.deleted=0", memberId, organizationCode).
		Count(&count)
	return count > 0
}

// createOrganization 创建组织
func (h *HandlerOrganization) createOrganization(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	name := c.PostForm("name")
	address := c.PostForm("address")
	description := c.PostForm("description")

	if name == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "组织名称不能为空"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 创建组织
	org := &organizationRow{
		Name:        name,
		Address:     address,
		Description: description,
		MemberId:    memberId,
		CreateTime:  time.Now().UnixMilli(),
		Personal:    1,
	}

	if err := db.Table("ms_organization").Create(org).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "创建组织失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(org))
}

// getOrganizationList 获取用户的组织列表
func (h *HandlerOrganization) getOrganizationList(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	orgList, err := h.loadOrganizations(memberId)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "查询组织失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  orgList,
		"total": len(orgList),
	}))
}

// getOrgList 获取组织列表（简化版）
func (h *HandlerOrganization) getOrgList(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	orgList, err := h.loadOrganizations(memberId)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "查询失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(orgList))
}

// editOrganization 编辑组织
func (h *HandlerOrganization) editOrganization(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	organizationCode := c.PostForm("organizationCode")
	name := c.PostForm("name")
	address := c.PostForm("address")
	description := c.PostForm("description")

	if organizationCode == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "请选择组织"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 解析 organizationCode
	orgCodeInt := decryptOrgCode(organizationCode)

	// 验证用户是否有权限编辑该组织（创建者或组织成员）
	var org organizationRow
	if err := db.Table("ms_organization").Where("id = ?", orgCodeInt).First(&org).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "组织不存在"))
		return
	}
	if org.MemberId != memberId && !isOrgMember(nil, memberId, orgCodeInt) {
		c.JSON(http.StatusOK, result.Fail(http.StatusForbidden, "无权限编辑该组织"))
		return
	}

	// 更新组织信息
	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if address != "" {
		updates["address"] = address
	}
	if description != "" {
		updates["description"] = description
	}

	if len(updates) > 0 {
		if err := db.Table("ms_organization").
			Where("id = ?", orgCodeInt).
			Updates(updates).Error; err != nil {
			c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "更新失败"))
			return
		}
	}

	c.JSON(http.StatusOK, result.Success(""))
}

// deleteOrganization 删除组织
func (h *HandlerOrganization) deleteOrganization(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	organizationCode := c.PostForm("organizationCode")
	if organizationCode == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "请选择组织"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	orgCodeInt := decryptOrgCode(organizationCode)

	// 验证用户是否是组织所有者
	var org organizationRow
	if err := db.Table("ms_organization").
		Where("id = ?", orgCodeInt).
		First(&org).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "组织不存在"))
		return
	}

	if org.MemberId != memberId {
		c.JSON(http.StatusOK, result.Fail(http.StatusForbidden, "无权限删除该组织"))
		return
	}

	// 删除组织
	if err := db.Table("ms_organization").
		Where("id = ?", orgCodeInt).
		Delete(nil).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "删除失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(""))
}

// quitOrganization 退出组织
func (h *HandlerOrganization) quitOrganization(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	organizationCode := c.PostForm("organizationCode")
	if organizationCode == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "请选择组织"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	orgCodeInt := decryptOrgCode(organizationCode)

	// 检查是否属于该组织（创建者不能退出自己的组织）
	var org organizationRow
	if err := db.Table("ms_organization").Where("id = ?", orgCodeInt).First(&org).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "组织不存在"))
		return
	}
	if org.MemberId == memberId {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "创建者不能退出自己的组织"))
		return
	}

	// 开启事务
	tx := db.Begin()

	// 退出该组织下的部门
	if err := tx.Exec("DELETE FROM ms_department_member WHERE member_id = ? AND department_id IN (SELECT id FROM ms_department WHERE organization_code = ?)", memberId, orgCodeInt).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "退出失败"))
		return
	}

	// 退出项目成员
	if err := tx.Exec("DELETE FROM ms_project_member WHERE member_code = ? AND project_code IN (SELECT id FROM ms_project WHERE organization_code = ?)", memberId, orgCodeInt).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "退出失败"))
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, result.Success(""))
}

// decryptOrgCode 辅助函数：解密组织代码
func decryptOrgCode(v interface{}) int64 {
	switch t := v.(type) {
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
