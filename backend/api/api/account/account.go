package account

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerAccount struct {
}

func New() *HandlerAccount {
	return &HandlerAccount{}
}

type memberRow struct {
	Id            int64 `gorm:"primaryKey;autoIncrement"`
	Account       string
	Password      string
	Name          string
	Mobile        string
	Email         string
	Status        int
	Avatar        string
	CreateTime    int64
	LastLoginTime int64
}

func (*memberRow) TableName() string { return "ms_member" }

func (h *HandlerAccount) list(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	db := gorms.GetDB()
	query := db.Model(&memberRow{})
	// 数据隔离：按组织过滤成员列表
	memberId := c.GetInt64("memberId")
	orgCodeStr, _ := c.Get("organizationCode")
	var orgCode int64
	if orgStr, ok := orgCodeStr.(string); ok && orgStr != "" {
		orgCode, _ = codecs.DecryptInt64(orgStr)
	}

	// 判断当前用户是否为非个人组织的拥有者/管理员
	isOrgAdmin := false
	if orgCode > 0 {
		var personal int64
		_ = db.Table("ms_organization").Select("personal").Where("id=?", orgCode).Scan(&personal).Error
		if personal == 0 {
			// 非个人组织，检查是否是拥有者
			var ownerCount int64
			_ = db.Table("ms_member_account").Where("member_code=? AND organization_code=? AND is_owner=1", memberId, orgCode).Count(&ownerCount).Error
			if ownerCount > 0 {
				isOrgAdmin = true
			}
		}
	}

	if isOrgAdmin {
		// 组织管理员：显示同组织所有成员 + 仅属于个人组织的用户（新注册未被分配的用户）
		// 1. 同组织的成员
		var orgMemberIds []int64
		_ = db.Table("ms_member_account").Select("DISTINCT member_code").Where("organization_code=?", orgCode).Scan(&orgMemberIds).Error
		// 2. 仅属于个人组织的用户（没有加入任何非个人组织的用户）
		var personalOnlyMemberIds []int64
		_ = db.Raw(`
			SELECT DISTINCT ma.member_code
			FROM ms_member_account ma
			JOIN ms_organization o ON o.id = ma.organization_code
			WHERE o.personal = 1
			  AND ma.member_code NOT IN (
			    SELECT DISTINCT ma2.member_code
			    FROM ms_member_account ma2
			    JOIN ms_organization o2 ON o2.id = ma2.organization_code
			    WHERE o2.personal = 0
			  )
		`).Scan(&personalOnlyMemberIds).Error
		// 合并两个列表
		allIds := append(orgMemberIds, personalOnlyMemberIds...)
		if len(allIds) > 0 {
			query = query.Where("id IN ?", allIds)
		}
	} else {
		// 普通用户：按组织成员过滤
		var orgMemberIds []int64
		if orgCode > 0 {
			_ = db.Table("ms_member_account").Select("DISTINCT member_code").Where("organization_code=?", orgCode).Scan(&orgMemberIds).Error
		}
		if len(orgMemberIds) > 0 {
			query = query.Where("id IN ?", orgMemberIds)
		} else {
			// 没有组织成员记录，退回项目成员过滤
			var userProjectIds []int64
			_ = db.Table("ms_project_member").Select("DISTINCT project_code").Where("member_code=?", memberId).Scan(&userProjectIds).Error
			if len(userProjectIds) > 0 {
				query = query.Where("id IN (SELECT DISTINCT member_code FROM ms_project_member WHERE project_code IN ?)", userProjectIds)
			} else {
				query = query.Where("id=?", memberId)
			}
		}
	}
	if keyword := c.PostForm("keyword"); keyword != "" {
		query = query.Where("name like ? or account like ? or email like ? or mobile like ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 处理 searchType 筛选
	searchType := c.PostForm("searchType")
	switch searchType {
	case "4": // 按部门筛选
		departmentCode := c.PostForm("departmentCode")
		if departmentCode != "" {
			depId, err := codecs.DecryptInt64(departmentCode)
			if err == nil {
				query = query.Where("id IN (SELECT member_id FROM ms_department_member WHERE department_id=?)", depId)
			}
		}
	case "2": // 未分配部门的成员
		query = query.Where("id NOT IN (SELECT member_id FROM ms_department_member)")
	case "3": // 停用的成员
		query = query.Where("status = 0")
	case "1": // 新加入的成员（最近7天）
		sevenDaysAgo := time.Now().AddDate(0, 0, -7).UnixMilli()
		query = query.Where("create_time >= ?", sevenDaysAgo)
	}

	var total int64
	_ = query.Count(&total).Error
	var list []memberRow
	_ = query.Order("id desc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&list).Error

	// 为每个成员查询部门信息和权限角色
	out := make([]gin.H, 0, len(list))
	for _, m := range list {
		// 查询成员所属部门
		var deptNames []string
		type deptRow struct {
			Name string
		}
		var depts []deptRow
		db.Table("ms_department").Select("ms_department.name").
			Joins("JOIN ms_department_member ON ms_department_member.department_id = ms_department.id").
			Where("ms_department_member.member_id = ? AND ms_department.deleted = 0", m.Id).
			Find(&depts)
		for _, d := range depts {
			deptNames = append(deptNames, d.Name)
		}
		deptStr := ""
		for i, dn := range deptNames {
			if i > 0 {
				deptStr += ", "
			}
			deptStr += dn
		}
		// 查询成员的权限角色
		authId := getMemberAuthId(db, m.Id)

		out = append(out, gin.H{
			"id":          m.Id,
			"code":        codecs.EncryptInt64(m.Id),
			"name":        m.Name,
			"account":     m.Account,
			"email":       m.Email,
			"mobile":      m.Mobile,
			"avatar":      m.Avatar,
			"status":      m.Status,
			"create_time": m.CreateTime,
			"departments": deptStr,
			"authorize":   authId,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total, "authList": authListFromDB(db)}))
}

func (h *HandlerAccount) allList(c *gin.Context) {
	result := &common.Result{}
	db := gorms.GetDB()
	memberId := c.GetInt64("memberId")
	// 数据隔离：优先按部门过滤，只返回与当前用户同部门的成员
	var deptIds []int64
	_ = db.Table("ms_department_member").Select("DISTINCT department_id").Where("member_id=?", memberId).Scan(&deptIds).Error
	var list []memberRow
	if len(deptIds) > 0 {
		_ = db.Where("id IN (SELECT DISTINCT member_id FROM ms_department_member WHERE department_id IN ?)", deptIds).
			Order("id desc").Limit(300).Find(&list).Error
	} else {
		// 没有部门的用户，退回到项目成员过滤
		var userProjectIds []int64
		_ = db.Table("ms_project_member").Select("DISTINCT project_code").Where("member_code=?", memberId).Scan(&userProjectIds)
		if len(userProjectIds) > 0 {
			_ = db.Where("id IN (SELECT DISTINCT member_code FROM ms_project_member WHERE project_code IN ?)", userProjectIds).
				Order("id desc").Limit(300).Find(&list).Error
		} else {
			list = []memberRow{}
		}
	}
	out := make([]gin.H, 0, len(list))
	for _, m := range list {
		out = append(out, gin.H{
			"id":     m.Id,
			"code":   codecs.EncryptInt64(m.Id),
			"name":   m.Name,
			"avatar": m.Avatar,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": len(out)}))
}

func (h *HandlerAccount) forbid(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("accountCode")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "accountCode无效"))
		return
	}
	_ = gorms.GetDB().Model(&memberRow{}).Where("id=?", id).Update("status", 0).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAccount) resume(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("accountCode")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "accountCode无效"))
		return
	}
	_ = gorms.GetDB().Model(&memberRow{}).Where("id=?", id).Update("status", 1).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAccount) add(c *gin.Context) {
	result := &common.Result{}
	row := &memberRow{
		Account:    c.PostForm("account"),
		Password:   c.PostForm("password"),
		Name:       c.PostForm("name"),
		Mobile:     c.PostForm("mobile"),
		Email:      c.PostForm("email"),
		Status:     1,
		Avatar:     c.PostForm("avatar"),
		CreateTime: time.Now().UnixMilli(),
	}
	if err := gorms.GetDB().Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"code": codecs.EncryptInt64(row.Id)}))
}

func (h *HandlerAccount) edit(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("code")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	updates := map[string]any{}
	for _, k := range []string{"name", "mobile", "email", "avatar"} {
		if v := c.PostForm(k); v != "" {
			updates[k] = v
		}
	}
	_ = gorms.GetDB().Model(&memberRow{}).Where("id=?", id).Updates(updates).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAccount) auth(c *gin.Context) {
	result := &common.Result{}
	idStr := c.PostForm("id")
	authStr := c.PostForm("auth")
	orgCodeStr := c.PostForm("organizationCode")
	if idStr == "" {
		c.JSON(http.StatusOK, result.Fail(400, "id必填"))
		return
	}
	// id 可能是加密的 code，也可能是原始数字 ID
	var id int64
	if decId, err := codecs.DecryptInt64(idStr); err == nil && decId > 0 {
		id = decId
	} else {
		// 尝试作为原始数字解析
		parsedId, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || parsedId == 0 {
			c.JSON(http.StatusOK, result.Fail(400, "id无效"))
			return
		}
		id = parsedId
	}
	db := gorms.GetDB().WithContext(c.Request.Context())

	// 如果指定了 organizationCode，则设置组织级角色
	if orgCodeStr != "" {
		orgCode, err := codecs.DecryptInt64(orgCodeStr)
		if err != nil || orgCode == 0 {
			orgCode, _ = strconv.ParseInt(orgCodeStr, 10, 64)
		}
		if orgCode > 0 {
			if authStr == "" || authStr == "0" {
				// 移除组织角色
				db.Table("ms_organization_auth").Where("member_code=? AND organization_code=?", id, orgCode).Delete(nil)
			} else {
				authId, _ := strconv.ParseInt(authStr, 10, 64)
				// Upsert
				var count int64
				db.Table("ms_organization_auth").Where("member_code=? AND organization_code=?", id, orgCode).Count(&count)
				if count > 0 {
					db.Table("ms_organization_auth").Where("member_code=? AND organization_code=?", id, orgCode).
						Update("auth_id", authId)
				} else {
					db.Table("ms_organization_auth").Create(map[string]interface{}{
						"organization_code": orgCode,
						"member_code":       id,
						"auth_id":           authId,
						"create_time":       time.Now().UnixMilli(),
					})
				}
			}
			c.JSON(http.StatusOK, result.Success([]int{}))
			return
		}
	}

	// 否则保持原有全局项目级授权行为（向后兼容）
	_ = db.Table("ms_project_member").Where("member_code=?", id).Update("authorize", authStr).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAccount) del(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("accountCode")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "accountCode无效"))
		return
	}
	_ = gorms.GetDB().Where("id=?", id).Delete(&memberRow{}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAccount) read(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("code")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	db := gorms.GetDB()
	var m memberRow
	if err := db.Where("id=?", id).First(&m).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "账号不存在"))
		return
	}
	memberCode := codecs.EncryptInt64(m.Id)
	c.JSON(http.StatusOK, result.Success(gin.H{
		"id":          m.Id,
		"code":        memberCode,
		"member_code": memberCode,
		"name":        m.Name,
		"account":     m.Account,
		"email":       m.Email,
		"mobile":      m.Mobile,
		"avatar":      m.Avatar,
		"status":      m.Status,
		"create_time": m.CreateTime,
		"position":    "",
		"departments": "",
		"description": "",
	}))
}

func (h *HandlerAccount) syncDetail(c *gin.Context) {
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
	db := gorms.GetDB().WithContext(c.Request.Context())
	var m memberRow
	if err := db.Where("id=?", id).First(&m).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "账号不存在"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"id":          m.Id,
		"code":        codecs.EncryptInt64(m.Id),
		"name":        m.Name,
		"account":     m.Account,
		"email":       m.Email,
		"mobile":      m.Mobile,
		"avatar":      m.Avatar,
		"status":      m.Status,
		"create_time": m.CreateTime,
	}))
}

func (h *HandlerAccount) joinByInviteLink(c *gin.Context) {
	result := &common.Result{}
	inviteCode := c.PostForm("inviteCode")
	if inviteCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "inviteCode必填"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// authListRow 权限角色表
type authListRow struct {
	Id        int64  `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"column:title"`
	Desc      string `gorm:"column:desc"`
	Status    int    `gorm:"column:status"`
	IsDefault int    `gorm:"column:is_default"`
}

func (*authListRow) TableName() string { return "ms_project_auth" }

// authListFromDB 从数据库查询所有启用的权限角色
func authListFromDB(db *gorm.DB) []gin.H {
	var rows []authListRow
	_ = db.Where("status=1").Order("id asc").Find(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":    r.Id,
			"title": r.Title,
		})
	}
	return out
}

// projectMemberAuthRow 用于查询成员的权限角色
type projectMemberAuthRow struct {
	Authorize string `gorm:"column:authorize"`
}

// getMemberAuthId 获取成员的权限角色ID
func getMemberAuthId(db *gorm.DB, memberId int64) int64 {
	// 优先查找 is_owner=1 的记录
	var pm projectMemberAuthRow
	err := db.Table("ms_project_member").Select("authorize").
		Where("member_code=? AND is_owner=1 AND authorize != '' AND authorize != '0'", memberId).
		Scan(&pm).Error
	if err == nil && pm.Authorize != "" {
		authId, _ := strconv.ParseInt(pm.Authorize, 10, 64)
		if authId > 0 {
			return authId
		}
	}
	// 其次查找任意有 authorize 的记录
	err = db.Table("ms_project_member").Select("authorize").
		Where("member_code=? AND authorize != '' AND authorize != '0'", memberId).
		Scan(&pm).Error
	if err == nil && pm.Authorize != "" {
		authId, _ := strconv.ParseInt(pm.Authorize, 10, 64)
		if authId > 0 {
			return authId
		}
	}
	return 0
}
