package index

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	ws "test.com/project-api/api/websocket"
	"test.com/project-api/internal/authz"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	common "test.com/project-common"
)

type HandlerIndex struct{}

func New() *HandlerIndex {
	return &HandlerIndex{}
}

type memberRow struct {
	Id            int64  `gorm:"primaryKey;autoIncrement"`
	Account       string `gorm:"column:account"`
	Password      string `gorm:"column:password"`
	Name          string `gorm:"column:name"`
	Mobile        string `gorm:"column:mobile"`
	Realname      string `gorm:"column:realname"`
	CreateTime    int64  `gorm:"column:create_time"`
	Status        int8   `gorm:"column:status"`
	LastLoginTime int64  `gorm:"column:last_login_time"`
	Sex           int8   `gorm:"column:sex"`
	Avatar        string `gorm:"column:avatar"`
	Idcard        string `gorm:"column:idcard"`
	Province      int    `gorm:"column:province"`
	City          int    `gorm:"column:city"`
	Area          int    `gorm:"column:area"`
	Email         string `gorm:"column:email"`
}

func (*memberRow) TableName() string { return "ms_member" }

type organizationRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"column:name"`
	Avatar      string `gorm:"column:avatar"`
	Description string `gorm:"column:description"`
}

func (*organizationRow) TableName() string { return "ms_organization" }

type projectMemberRow struct {
	Id             int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode    int64 `gorm:"column:project_code"`
	MemberCode     int64 `gorm:"column:member_code"`
	JoinTime       int64 `gorm:"column:join_time"`
	IsOwner        int8  `gorm:"column:is_owner"`
	Authorize      string `gorm:"column:authorize"`
	OrganizationCode int64 `gorm:"column:organization_code"`
}

func (*projectMemberRow) TableName() string { return "ms_project_member" }

type menuRow struct {
	Id       int64  `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"column:name"`
	Title    string `gorm:"column:title"`
	Icon     string `gorm:"column:icon"`
	ParentId int64  `gorm:"column:parent_id"`
	Sort     int    `gorm:"column:sort"`
	Status   int8   `gorm:"column:status"`
}

func (*menuRow) TableName() string { return "ms_project_menu" }

type departmentRow struct {
	Id               int64  `gorm:"primaryKey;autoIncrement"`
	Name             string `gorm:"column:name"`
	ParentId         int64  `gorm:"column:parent_id"`
	OrganizationCode int64  `gorm:"column:organization_code"`
	Sort             int    `gorm:"column:sort"`
}

func (*departmentRow) TableName() string { return "ms_department" }

type departmentMemberRow struct {
	Id           int64 `gorm:"primaryKey;autoIncrement"`
	DepartmentId int64 `gorm:"column:department_id"`
	MemberId     int64 `gorm:"column:member_id"`
}

func (*departmentMemberRow) TableName() string { return "ms_department_member" }

// index 获取用户菜单（简化版）
func (h *HandlerIndex) index(c *gin.Context) {
	result := &common.Result{}
	_ = c.GetInt64("memberId")       // memberId 暂未使用
	_ = c.GetInt64("organizationCode") // orgCode 暂未使用

	db := gorms.GetDB()

	// 获取用户菜单权限
	var menus []menuRow
	// 这里简化处理，实际应该根据用户权限查询菜单
	db.Where("status = 1").Order("sort ASC").Find(&menus)

	menuList := make([]gin.H, 0, len(menus))
	for _, menu := range menus {
		menuList = append(menuList, gin.H{
			"code":     codecs.EncryptInt64(menu.Id),
			"name":     menu.Name,
			"title":    menu.Title,
			"icon":     menu.Icon,
			"parentId": menu.ParentId,
			"sort":     menu.Sort,
		})
	}

	// 如果没有菜单数据，返回示例菜单
	if len(menuList) == 0 {
		menuList = []gin.H{
			{"code": "1", "name": "home", "title": "首页", "icon": "home", "parentId": 0, "sort": 0},
			{"code": "2", "name": "project", "title": "项目", "icon": "project", "parentId": 0, "sort": 1},
		}
	}

	c.JSON(http.StatusOK, result.Success(menuList))
}

// menus 获取用户菜单（详细版）
func (h *HandlerIndex) menus(c *gin.Context) {
	result := &common.Result{}
	_ = c.GetInt64("memberId")       // memberId 暂未使用
	_ = c.GetInt64("organizationCode") // orgCode 暂未使用

	db := gorms.GetDB()

	var menus []menuRow
	db.Where("status = 1").Order("sort ASC").Find(&menus)

	menuList := buildMenuTree(menus, 0)

	c.JSON(http.StatusOK, result.Success(gin.H{
		"menus":       menus,
		"menusFormat": menuList,
	}))
}

// buildMenuTree 构建菜单树
func buildMenuTree(menus []menuRow, parentId int64) []gin.H {
	tree := make([]gin.H, 0)
	for _, menu := range menus {
		if menu.ParentId == parentId {
			node := gin.H{
				"code":     codecs.EncryptInt64(menu.Id),
				"name":     menu.Name,
				"title":    menu.Title,
				"icon":     menu.Icon,
				"parentId": menu.ParentId,
				"sort":     menu.Sort,
				"children": buildMenuTree(menus, menu.Id),
			}
			tree = append(tree, node)
		}
	}
	return tree
}

// changeCurrentOrganization 切换当前组织
func (h *HandlerIndex) changeCurrentOrganization(c *gin.Context) {
	result := &common.Result{}
	organizationCode := c.PostForm("organizationCode")

	if organizationCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "请选择组织"))
		return
	}

	orgId, err := codecs.DecryptInt64(organizationCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "organizationCode无效"))
		return
	}

	memberId := c.GetInt64("memberId")
	db := gorms.GetDB()

	// 检查用户是否属于该组织
	var projectMember projectMemberRow
	err = db.Where("member_code = ? AND organization_code = ?", memberId, orgId).First(&projectMember).Error
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, result.Fail(403, "您不属于该组织"))
		return
	}

	// 获取组织信息
	var org organizationRow
	db.Where("id = ?", orgId).First(&org)

	// 获取用户部门信息
	var deptMembers []departmentMemberRow
	db.Where("member_id = ?", memberId).Find(&deptMembers)

	departments := make([]string, 0)
	for _, dm := range deptMembers {
		var dept departmentRow
		db.Where("id = ? AND organization_code = ?", dm.DepartmentId, orgId).First(&dept)
		if dept.Name != "" {
			departments = append(departments, dept.Name)
		}
	}

	// 获取用户信息
	var member memberRow
	db.Where("id = ?", memberId).First(&member)

	// 返回更新后的用户信息
	memberInfo := gin.H{
		"code":             codecs.EncryptInt64(member.Id),
		"name":             member.Name,
		"avatar":           member.Avatar,
		"email":            member.Email,
		"mobile":           member.Mobile,
		"position":         "", // 职位信息可以从 projectMember 获取
		"department":       joinStrings(departments, " - "),
		"is_owner":         projectMember.IsOwner,
		"authorize":        projectMember.Authorize,
		"organization_code": organizationCode,
	}

	// 更新上下文（在实际应用中，这里应该更新session或token）
	c.Set("organizationCode", orgId)

	// 获取菜单列表
	var menus []menuRow
	db.Where("status = 1").Order("sort ASC").Find(&menus)
	menuList := make([]gin.H, 0, len(menus))
	for _, menu := range menus {
		menuList = append(menuList, gin.H{
			"code":     codecs.EncryptInt64(menu.Id),
			"name":     menu.Name,
			"title":    menu.Title,
			"icon":     menu.Icon,
			"parentId": menu.ParentId,
			"sort":     menu.Sort,
		})
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"menuList": menuList,
		"member":   memberInfo,
	}))
}

// systemConfig 获取系统配置信息
func (h *HandlerIndex) systemConfig(c *gin.Context) {
	result := &common.Result{}

	// 由于没有 system_config 表，返回默认配置
	config := gin.H{
		"app_name":    "Pear项目管理",
		"app_version": "1.0.0",
		"miitbeian":  "",
		"site_copy":   "© 2024 Pear Project",
		"site_name":   "Pear",
	}

	c.JSON(http.StatusOK, result.Success(config))
}

// info 获取个人信息
func (h *HandlerIndex) info(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	db := gorms.GetDB()
	var member memberRow
	err := db.Where("id = ?", memberId).First(&member).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "用户不存在"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":         codecs.EncryptInt64(member.Id),
		"account":      member.Account,
		"name":         member.Name,
		"mobile":       member.Mobile,
		"email":        member.Email,
		"realname":     member.Realname,
		"avatar":       member.Avatar,
		"idcard":       member.Idcard,
		"sex":          member.Sex,
		"status":       member.Status,
		"createTime":   member.CreateTime,
		"lastLoginTime": member.LastLoginTime,
	}))
}

// editPersonal 编辑个人资料
func (h *HandlerIndex) editPersonal(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	mobile := c.PostForm("mobile")
	email := c.PostForm("email")
	name := c.PostForm("name")
	realname := c.PostForm("realname")
	avatar := c.PostForm("avatar")
	idcard := c.PostForm("idcard")

	db := gorms.GetDB()
	updates := map[string]interface{}{}

	if mobile != "" {
		updates["mobile"] = mobile
	}
	if email != "" {
		updates["email"] = email
	}
	if name != "" {
		updates["name"] = name
	}
	if realname != "" {
		updates["realname"] = realname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if idcard != "" {
		updates["idcard"] = idcard
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "没有需要更新的内容"))
		return
	}

	if err := db.Model(&memberRow{}).Where("id = ?", memberId).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(nil))
}

// editPassword 修改密码
func (h *HandlerIndex) editPassword(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	password := c.PostForm("password")
	newPassword := c.PostForm("newPassword")
	confirmPassword := c.PostForm("confirmPassword")

	if password == "" || newPassword == "" || confirmPassword == "" {
		c.JSON(http.StatusOK, result.Fail(400, "参数不完整"))
		return
	}

	if len(newPassword) < 6 || len(confirmPassword) < 6 {
		c.JSON(http.StatusOK, result.Fail(400, "密码必须包含6个字符"))
		return
	}

	if newPassword != confirmPassword {
		c.JSON(http.StatusOK, result.Fail(400, "两次新密码不匹配"))
		return
	}

	db := gorms.GetDB()
	var member memberRow
	err := db.Where("id = ?", memberId).First(&member).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "用户不存在"))
		return
	}

	// 验证原密码（MD5加密后比对，与注册逻辑一致）
	oldMd5 := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	if oldMd5 != member.Password {
		c.JSON(http.StatusOK, result.Fail(400, "原密码不正确"))
		return
	}

	// 更新密码（MD5加密存储，与注册逻辑一致）
	newMd5 := fmt.Sprintf("%x", md5.Sum([]byte(newPassword)))
	if err := db.Model(&memberRow{}).Where("id = ?", memberId).Update("password", newMd5).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "密码修改失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(nil))
}

// uploadAvatar 上传头像
func (h *HandlerIndex) uploadAvatar(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "请选择头像文件"))
		return
	}

	// 保存文件（这里简化处理，实际应该上传到OSS或文件服务器）
	// 生成文件名
	filename := time.Now().Format("20060102150405") + "_" + file.Filename
	filepath := "/uploads/avatars/" + filename

	// 保存文件
	if err := c.SaveUploadedFile(file, "."+filepath); err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "上传失败"))
		return
	}

	// 更新用户头像
	db := gorms.GetDB()
	if err := db.Model(&memberRow{}).Where("id = ?", memberId).Update("avatar", filepath).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新头像失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"url": filepath,
	}))
}

// uploadImg 上传编辑器图片
func (h *HandlerIndex) uploadImg(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "请选择图片文件"))
		return
	}

	// 创建上传目录
	uploadDir := "uploads/editor"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建上传目录失败"))
		return
	}

	// 生成文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), memberId, ext)
	filePath := uploadDir + "/" + filename

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "上传失败"))
		return
	}

	// 生成访问URL
	imgUrl := "/" + strings.ReplaceAll(filePath, "\\", "/")

	c.JSON(http.StatusOK, result.Success(gin.H{
		"url": imgUrl,
	}))
}

// bindClientId 绑定 WebSocket client_id 到用户（前端 socket.vue 调用）
func (h *HandlerIndex) bindClientId(c *gin.Context) {
	result := &common.Result{}
	clientID := c.PostForm("client_id")
	uid := c.PostForm("uid")

	if clientID == "" || uid == "" {
		c.JSON(http.StatusOK, result.Fail(400, "参数不能为空"))
		return
	}

	if ws.Manager.BindUser(clientID, uid) {
		c.JSON(http.StatusOK, result.Success(gin.H{"bound": true}))
	} else {
		c.JSON(http.StatusOK, result.Fail(404, "WebSocket连接不存在"))
	}
}

// joinStrings 连接字符串
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// nodes 获取当前用户的权限节点列表
func (h *HandlerIndex) nodes(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB()
	nodes := authz.GetUserNodes(db, memberId)
	if nodes == nil {
		nodes = []string{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"nodes": nodes,
	}))
}
