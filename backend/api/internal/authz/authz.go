package authz

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/internal/menus"
	"test.com/project-api/pkg/codecs"
	projectMenu "test.com/project-api/pkg/model/menu"
)

func IsProjectMember(db *gorm.DB, memberId, projectId int64) bool {
	var count int64
	_ = db.Table("ms_project_member").Where("member_code=? AND project_code=?", memberId, projectId).Count(&count).Error
	return count > 0
}

func IsProjectOwner(db *gorm.DB, memberId, projectId int64) bool {
	var count int64
	_ = db.Table("ms_project_member").Where("member_code=? AND project_code=? AND is_owner=1", memberId, projectId).Count(&count).Error
	return count > 0
}

func CanOperateProject(c *gin.Context, projectCode string) (int64, bool) {
	memberId := c.GetInt64("memberId")
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		return 0, false
	}
	db := getDB(c)
	if !IsProjectMember(db, memberId, projectId) {
		return projectId, false
	}
	return projectId, true
}

func CanManageProject(c *gin.Context, projectCode string) (int64, bool) {
	memberId := c.GetInt64("memberId")
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		return 0, false
	}
	db := gorms.GetDB()
	if !IsProjectOwner(db, memberId, projectId) {
		return projectId, false
	}
	return projectId, true
}

// getDB 优先取 Gin context 中注入的测试 DB，否则用生产 DB
func getDB(c *gin.Context) *gorm.DB {
	if db, ok := c.Keys["test_db"].(*gorm.DB); ok && db != nil {
		return db
	}
	return gorms.GetDB()
}

func CanOperateTask(c *gin.Context, taskCode string) (int64, int64, bool) {
	memberId := c.GetInt64("memberId")
	taskId, err := codecs.DecryptInt64(taskCode)
	if err != nil || taskId == 0 {
		return 0, 0, false
	}
	db := getDB(c)
	var row struct {
		ProjectCode int64 `gorm:"column:project_code"`
	}
	_ = db.Table("ms_task").Select("project_code").Where("id=?", taskId).Scan(&row).Error
	if row.ProjectCode == 0 {
		return taskId, 0, false
	}
	if !IsProjectMember(db, memberId, row.ProjectCode) {
		return taskId, row.ProjectCode, false
	}
	return taskId, row.ProjectCode, true
}

// ===== 角色权限相关 =====

// authRow 权限角色表
type authRow struct {
	Id        int64  `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"column:title"`
	Status    int    `gorm:"column:status"`
	IsDefault int    `gorm:"column:is_default"`
}

func (*authRow) TableName() string { return "ms_project_auth" }

// authNodeRow 权限节点关联表
type authNodeRow struct {
	Id     int64  `gorm:"primaryKey;autoIncrement"`
	AuthId int64  `gorm:"column:auth_id;index"`
	Node   string `gorm:"column:node"`
}

func (*authNodeRow) TableName() string { return "ms_project_auth_node" }

// projectMemberRow 项目成员表（用于读取 authorize 字段）
type projectMemberRow struct {
	Id             int64  `gorm:"primaryKey;autoIncrement"`
	MemberCode     int64  `gorm:"column:member_code"`
	IsOwner        int8   `gorm:"column:is_owner"`
	Authorize      string `gorm:"column:authorize"`
}

func (*projectMemberRow) TableName() string { return "ms_project_member" }

// GetUserAuthId 获取用户的权限角色ID
// 优先从 ms_project_member.authorize 读取（取 is_owner=1 的记录，否则取第一条有 authorize 值的记录）
// 如果为空则使用默认角色
func GetUserAuthId(db *gorm.DB, memberId int64) int64 {
	// 优先查找用户作为 Owner 的成员记录
	var pm projectMemberRow
	err := db.Where("member_code=? AND is_owner=1", memberId).First(&pm).Error
	if err == nil && pm.Authorize != "" {
		authId, _ := strconv.ParseInt(pm.Authorize, 10, 64)
		if authId > 0 {
			return authId
		}
	}
	// 其次查找任意一条有 authorize 值的成员记录
	err = db.Where("member_code=? AND authorize != '' AND authorize != '0'", memberId).First(&pm).Error
	if err == nil && pm.Authorize != "" {
		authId, _ := strconv.ParseInt(pm.Authorize, 10, 64)
		if authId > 0 {
			return authId
		}
	}
	// 使用默认角色
	var auth authRow
	if err := db.Where("is_default=1 AND status=1").First(&auth).Error; err == nil {
		return auth.Id
	}
	return 0
}

// GetUserNodes 获取用户的权限节点列表
func GetUserNodes(db *gorm.DB, memberId int64) []string {
	authId := GetUserAuthId(db, memberId)
	if authId == 0 {
		return []string{}
	}
	var rows []authNodeRow
	db.Where("auth_id=?", authId).Find(&rows)
	nodes := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.Node != "" {
			nodes = append(nodes, r.Node)
		}
	}
	return nodes
}

// HasNode 检查用户是否拥有指定节点权限
func HasNode(db *gorm.DB, memberId int64, node string) bool {
	if node == "" || node == "#" {
		return true
	}
	nodes := GetUserNodes(db, memberId)
	for _, n := range nodes {
		if n == node {
			return true
		}
	}
	return false
}

// FilterMenusByNodes 根据用户权限节点过滤菜单树
// node 为 "#" 或空的菜单对所有人可见
// 父菜单如果有任一子菜单可见则保留
func FilterMenusByNodes(tree []*projectMenu.Menu, allowedNodes map[string]bool) []*projectMenu.Menu {
	result := make([]*projectMenu.Menu, 0, len(tree))
	for _, item := range tree {
		filtered := filterMenuItem(item, allowedNodes)
		if filtered != nil {
			result = append(result, filtered)
		}
	}
	return result
}

func filterMenuItem(item *projectMenu.Menu, allowedNodes map[string]bool) *projectMenu.Menu {
	// node 为 "#" 或空表示不绑定权限，所有人可见
	if item.Node == "" || item.Node == "#" {
		// 仍然需要过滤子菜单
		if len(item.Children) > 0 {
			filteredChildren := FilterMenusByNodes(item.Children, allowedNodes)
			copy := *item
			copy.Children = filteredChildren
			return &copy
		}
		return item
	}

	// 有子菜单时，只要任一子菜单可见则保留
	if len(item.Children) > 0 {
		filteredChildren := FilterMenusByNodes(item.Children, allowedNodes)
		if len(filteredChildren) > 0 {
			copy := *item
			copy.Children = filteredChildren
			return &copy
		}
		// 子菜单都不可见，但自身有权限则保留
		if allowedNodes[item.Node] {
			copy := *item
			copy.Children = []*projectMenu.Menu{}
			return &copy
		}
		return nil
	}

	// 叶子节点：检查是否有权限
	if allowedNodes[item.Node] {
		return item
	}
	return nil
}

// FilterProjectMenusByNodes 对 ProjectMenu 列表过滤（用于 BuildTree 之前）
func FilterProjectMenusByNodes(pms []*menus.ProjectMenu, allowedNodes map[string]bool) []*menus.ProjectMenu {
	result := make([]*menus.ProjectMenu, 0, len(pms))
	for _, pm := range pms {
		// node 为 "#" 或空不绑定权限，不过滤
		if pm.Node == "" || pm.Node == "#" || allowedNodes[pm.Node] {
			result = append(result, pm)
		}
	}
	return result
}

// IsAdmin 判断用户是否为管理员（拥有所有项目 is_owner=1 的记录）
func IsAdmin(db *gorm.DB, memberId int64) bool {
	var count int64
	db.Table("ms_project_member").Where("member_code=? AND is_owner=1", memberId).Count(&count)
	return count > 0
}

// ===== 组织级权限 =====

// orgAuthRow 组织角色授权表
type orgAuthRow struct {
	Id               int64 `gorm:"primaryKey;autoIncrement"`
	OrganizationCode int64 `gorm:"column:organization_code"`
	MemberCode       int64 `gorm:"column:member_code"`
	AuthId           int64 `gorm:"column:auth_id"`
}

func (*orgAuthRow) TableName() string { return "ms_organization_auth" }

// GetUserOrgAuthId 获取用户在指定组织中的权限角色ID
// 优先查 ms_organization_auth，如果没有则判断是否为组织创建者（创建者给予 org_owner 角色），
// 最后回退到项目级默认角色
func GetUserOrgAuthId(db *gorm.DB, memberId, orgCode int64) int64 {
	// 1. 从组织角色表查询
	var oa orgAuthRow
	err := db.Where("member_code=? AND organization_code=?", memberId, orgCode).First(&oa).Error
	if err == nil && oa.AuthId > 0 {
		return oa.AuthId
	}
	// 2. 判断是否为组织创建者
	var ownerCount int64
	db.Table("ms_organization").Where("id=? AND member_id=?", orgCode, memberId).Count(&ownerCount)
	if ownerCount > 0 {
		// 组织创建者使用默认角色
		var auth authRow
		if err := db.Where("is_default=1 AND status=1").First(&auth).Error; err == nil {
			return auth.Id
		}
	}
	// 3. 判断是否为 ms_member_account 中 is_owner=1 的组织管理员
	var accountOwnerCount int64
	db.Table("ms_member_account").Where("member_code=? AND organization_code=? AND is_owner=1", memberId, orgCode).Count(&accountOwnerCount)
	if accountOwnerCount > 0 {
		var auth authRow
		if err := db.Where("is_default=1 AND status=1").First(&auth).Error; err == nil {
			return auth.Id
		}
	}
	return 0
}

// GetUserOrgNodes 获取用户在指定组织中的权限节点列表
func GetUserOrgNodes(db *gorm.DB, memberId, orgCode int64) []string {
	authId := GetUserOrgAuthId(db, memberId, orgCode)
	if authId == 0 {
		return []string{}
	}
	var rows []authNodeRow
	db.Where("auth_id=?", authId).Find(&rows)
	nodes := make([]string, 0, len(rows))
	for _, r := range rows {
		if r.Node != "" {
			nodes = append(nodes, r.Node)
		}
	}
	return nodes
}

// HasOrgNode 检查用户是否拥有指定组织级节点权限
func HasOrgNode(db *gorm.DB, memberId, orgCode int64, node string) bool {
	if node == "" || node == "#" {
		return true
	}
	nodes := GetUserOrgNodes(db, memberId, orgCode)
	for _, n := range nodes {
		if n == node {
			return true
		}
	}
	return false
}

// GetAllNodes 获取用户的所有权限节点（项目级 + 组织级合并）
func GetAllNodes(db *gorm.DB, memberId, orgCode int64) []string {
	projectNodes := GetUserNodes(db, memberId)
	orgNodes := GetUserOrgNodes(db, memberId, orgCode)
	// 合并去重
	merged := make(map[string]bool)
	for _, n := range projectNodes {
		merged[n] = true
	}
	for _, n := range orgNodes {
		merged[n] = true
	}
	result := make([]string, 0, len(merged))
	for n := range merged {
		result = append(result, n)
	}
	return result
}
