package organization

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test.com/project-api/api/rpc"
	"test.com/project-api/internal/authz"
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

	// 清除组织角色授权
	if err := tx.Exec("DELETE FROM ms_organization_auth WHERE member_code = ? AND organization_code = ?", memberId, orgCodeInt).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "退出失败"))
		return
	}

	// 清除组织成员账户关联
	if err := tx.Exec("DELETE FROM ms_member_account WHERE member_code = ? AND organization_code = ?", memberId, orgCodeInt).Error; err != nil {
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

// ===== 组织级角色管理 =====

// orgAuthRow 组织角色授权表
type orgAuthRow struct {
	Id               int64 `gorm:"primaryKey;autoIncrement"`
	OrganizationCode int64 `gorm:"column:organization_code"`
	MemberCode       int64 `gorm:"column:member_code"`
	AuthId           int64 `gorm:"column:auth_id"`
	CreateTime       int64 `gorm:"column:create_time"`
}

func (*orgAuthRow) TableName() string { return "ms_organization_auth" }

// orgMemberListReq 组织成员列表查询参数
type orgMemberListReq struct {
	OrganizationCode string `form:"organizationCode" json:"organizationCode"`
	Keyword          string `form:"keyword" json:"keyword"`
	SearchType       string `form:"searchType" json:"searchType"`
	DepartmentCode   string `form:"departmentCode" json:"departmentCode"`
	Page             int64  `form:"page" json:"page"`
	PageSize         int64  `form:"pageSize" json:"pageSize"`
}

// listMembersWithAuth 获取组织成员列表及其角色
func (h *HandlerOrganization) listMembersWithAuth(c *gin.Context) {
	result := &common.Result{}
	db := gorms.GetDB().WithContext(c.Request.Context())

	orgCode := decryptOrgCode(c.PostForm("organizationCode"))
	if orgCode == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "缺少组织参数"))
		return
	}

	// 分页参数
	page := int64(1)
	pageSize := int64(20)
	if p := c.PostForm("page"); p != "" {
		page, _ = strconv.ParseInt(p, 10, 64)
	}
	if ps := c.PostForm("pageSize"); ps != "" {
		pageSize, _ = strconv.ParseInt(ps, 10, 64)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 查询组织成员（通过 ms_member_account）
	type memberResult struct {
		Id         int64  `gorm:"column:id"`
		Account    string `gorm:"column:account"`
		Name       string `gorm:"column:name"`
		Status     int    `gorm:"column:status"`
		Avatar     string `gorm:"column:avatar"`
		Email      string `gorm:"column:email"`
		Mobile     string `gorm:"column:mobile"`
		CreateTime int64  `gorm:"column:create_time"`
		IsOwner    int    `gorm:"column:is_owner"`
	}

	query := db.Table("ms_member AS m").
		Select("m.id, m.account, m.name, m.status, m.avatar, m.email, m.mobile, m.create_time, ma.is_owner").
		Joins("JOIN ms_member_account AS ma ON ma.member_code = m.id AND ma.organization_code = ?", orgCode)

	keyword := c.PostForm("keyword")
	if keyword != "" {
		query = query.Where("m.name LIKE ? OR m.account LIKE ? OR m.email LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	searchType := c.PostForm("searchType")
	switch searchType {
	case "4": // 按部门筛选
		departmentCode := c.PostForm("departmentCode")
		if departmentCode != "" {
			depId, err := codecs.DecryptInt64(departmentCode)
			if err == nil {
				query = query.Where("m.id IN (SELECT member_id FROM ms_department_member WHERE department_id=?)", depId)
			}
		}
	case "3": // 停用的成员
		query = query.Where("m.status = 0")
	}

	var total int64
	_ = query.Count(&total).Error

	var members []memberResult
	_ = query.Order("ma.is_owner DESC, m.id DESC").
		Limit(int(pageSize)).Offset(int((page - 1) * pageSize)).
		Scan(&members).Error

	// 为每个成员查询部门和角色
	out := make([]gin.H, 0, len(members))
	for _, m := range members {
		// 查询部门
		type deptRow struct {
			Id   int64  `gorm:"column:id"`
			Name string `gorm:"column:name"`
		}
		var depts []deptRow
		db.Table("ms_department AS d").
			Select("d.id, d.name").
			Joins("JOIN ms_department_member AS dm ON dm.department_id = d.id").
			Where("dm.member_id = ? AND d.organization_code = ? AND d.deleted = 0", m.Id, orgCode).
			Scan(&depts)

		deptNames := make([]string, 0, len(depts))
		deptCodes := make([]string, 0, len(depts))
		for _, d := range depts {
			deptNames = append(deptNames, d.Name)
			deptCodes = append(deptCodes, codecs.EncryptInt64(d.Id))
		}

		// 查询组织级角色
		var orgAuth orgAuthRow
		authId := int64(0)
		if err := db.Where("member_code=? AND organization_code=?", m.Id, orgCode).First(&orgAuth).Error; err == nil {
			authId = orgAuth.AuthId
		}

		out = append(out, gin.H{
			"id":            m.Id,
			"code":          codecs.EncryptInt64(m.Id),
			"name":          m.Name,
			"account":       m.Account,
			"email":         m.Email,
			"mobile":        m.Mobile,
			"avatar":        m.Avatar,
			"status":        m.Status,
			"create_time":   m.CreateTime,
			"is_owner":      m.IsOwner,
			"departments":   deptNames,
			"departmentCodes": deptCodes,
			"authorize":     authId,
		})
	}

	// 获取可用角色列表
	authList := authListForOrg(db)

	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":     out,
		"total":    total,
		"authList": authList,
	}))
}

// setMemberAuth 设置组织成员的角色
func (h *HandlerOrganization) setMemberAuth(c *gin.Context) {
	result := &common.Result{}
	db := gorms.GetDB().WithContext(c.Request.Context())

	orgCodeStr := c.PostForm("organizationCode")
	memberCodeStr := c.PostForm("memberCode")
	authIdStr := c.PostForm("authId")

	orgCode := decryptOrgCode(orgCodeStr)
	if orgCode == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "organizationCode无效"))
		return
	}

	memberCode := decryptOrgCode(memberCodeStr)
	if memberCode == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "memberCode无效"))
		return
	}

	// 不能给自己设置角色（防止权限提升）
	memberId := c.GetInt64("memberId")
	if memberCode == memberId {
		c.JSON(http.StatusOK, result.Fail(400, "不能修改自己的角色"))
		return
	}

	// 验证目标用户是否属于该组织
	var mCount int64
	db.Table("ms_member_account").Where("member_code=? AND organization_code=?", memberCode, orgCode).Count(&mCount)
	if mCount == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "该用户不属于此组织"))
		return
	}

	if authIdStr == "" || authIdStr == "0" {
		// 移除角色
		db.Where("member_code=? AND organization_code=?", memberCode, orgCode).Delete(&orgAuthRow{})
		c.JSON(http.StatusOK, result.Success([]int{}))
		return
	}

	authId, err := strconv.ParseInt(authIdStr, 10, 64)
	if err != nil || authId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "authId无效"))
		return
	}

	// 验证角色是否存在
	var authCount int64
	db.Table("ms_project_auth").Where("id=? AND status=1", authId).Count(&authCount)
	if authCount == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "角色不存在或已禁用"))
		return
	}

	// Upsert：存在则更新，不存在则创建
	var existing orgAuthRow
	err = db.Where("member_code=? AND organization_code=?", memberCode, orgCode).First(&existing).Error
	if err == nil {
		db.Model(&orgAuthRow{}).Where("id=?", existing.Id).Update("auth_id", authId)
	} else {
		db.Create(&orgAuthRow{
			OrganizationCode: orgCode,
			MemberCode:       memberCode,
			AuthId:           authId,
			CreateTime:       time.Now().UnixMilli(),
		})
	}

	c.JSON(http.StatusOK, result.Success([]int{}))
}

// removeMemberAuth 移除组织成员的角色（恢复为默认）
func (h *HandlerOrganization) removeMemberAuth(c *gin.Context) {
	result := &common.Result{}
	db := gorms.GetDB().WithContext(c.Request.Context())

	orgCode := decryptOrgCode(c.PostForm("organizationCode"))
	memberCode := decryptOrgCode(c.PostForm("memberCode"))

	if orgCode == 0 || memberCode == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "参数无效"))
		return
	}

	// 不能移除自己
	memberId := c.GetInt64("memberId")
	if memberCode == memberId {
		c.JSON(http.StatusOK, result.Fail(400, "不能修改自己的角色"))
		return
	}

	db.Where("member_code=? AND organization_code=?", memberCode, orgCode).Delete(&orgAuthRow{})
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// getMemberAuth 获取指定成员在当前组织的角色
func (h *HandlerOrganization) getMemberAuth(c *gin.Context) {
	result := &common.Result{}
	db := gorms.GetDB().WithContext(c.Request.Context())

	orgCode := decryptOrgCode(c.PostForm("organizationCode"))
	memberCode := decryptOrgCode(c.PostForm("memberCode"))

	if orgCode == 0 || memberCode == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "参数无效"))
		return
	}

	authId := authz.GetUserOrgAuthId(db, memberCode, orgCode)
	nodes := authz.GetUserOrgNodes(db, memberCode, orgCode)

	// 获取角色名称
	roleName := ""
	if authId > 0 {
		type roleRow struct {
			Title string `gorm:"column:title"`
		}
		var r roleRow
		db.Table("ms_project_auth").Select("title").Where("id=?", authId).Scan(&r)
		roleName = r.Title
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"authId":   authId,
		"roleName": roleName,
		"nodes":    nodes,
	}))
}

// authListForOrg 获取组织可用的角色列表
func authListForOrg(db *gorm.DB) []gin.H {
	type authRow struct {
		Id        int64  `gorm:"column:id"`
		Title     string `gorm:"column:title"`
		Desc      string `gorm:"column:desc"`
		Status    int    `gorm:"column:status"`
		IsDefault int    `gorm:"column:is_default"`
	}
	var rows []authRow
	db.Table("ms_project_auth").Where("status=1").Order("id ASC").Scan(&rows)
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":         r.Id,
			"title":      r.Title,
			"desc":       r.Desc,
			"is_default": r.IsDefault,
		})
	}
	return out
}
