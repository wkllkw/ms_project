package node

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	common "test.com/project-common"
)

type HandlerNode struct {
}

func New() *HandlerNode {
	return &HandlerNode{}
}

type nodeItem struct {
	Node     string      `json:"node"`
	Title    string      `json:"title"`
	PNode    string      `json:"pnode,omitempty"`
	IsLogin  bool        `json:"is_login,omitempty"`
	IsAuth   bool        `json:"is_auth,omitempty"`
	Children []*nodeItem `json:"children,omitempty"`
}

func baseNodes() []nodeItem {
	return []nodeItem{
		{Node: "#", Title: "不绑定权限(#)"},
		{Node: "home", Title: "首页"},
		{Node: "project.list", Title: "项目列表"},
		{Node: "project.template", Title: "项目模板"},
		{Node: "project.analysis", Title: "数据分析"},
		{Node: "project.archive", Title: "归档项目"},
		{Node: "project.recycle", Title: "回收站"},
		{Node: "project.manage", Title: "项目管理"},
		{Node: "project.delete", Title: "项目删除"},
		{Node: "project.member", Title: "项目成员管理"},
		{Node: "organization.manage", Title: "组织管理"},
		{Node: "calendar", Title: "日程"},
		{Node: "notify.notice", Title: "通知列表"},
		{Node: "notify.system", Title: "系统消息"},
		{Node: "members.index", Title: "成员列表"},
		{Node: "task:create", Title: "创建任务"},
		{Node: "task:edit", Title: "编辑任务"},
		{Node: "task:assign", Title: "分配任务"},
		{Node: "task:delete", Title: "删除任务"},
		{Node: "file:upload", Title: "上传文件"},
		{Node: "file:delete", Title: "删除文件"},
		{Node: "system.account", Title: "账号管理"},
		{Node: "system.menu", Title: "菜单管理"},
		{Node: "system.node", Title: "节点管理"},
		{Node: "system.account.auth", Title: "权限管理"},
	}
}

func (h *HandlerNode) list(c *gin.Context) {
	result := &common.Result{}
	nodes := []nodeItem{
		{
			Node:  "project",
			Title: "项目权限",
			Children: []*nodeItem{
				{Node: "home", Title: "首页", PNode: "project"},
				{Node: "project.list", Title: "项目列表", PNode: "project"},
				{Node: "project.template", Title: "项目模板", PNode: "project"},
				{Node: "project.analysis", Title: "数据分析", PNode: "project"},
				{Node: "project.archive", Title: "归档项目", PNode: "project"},
				{Node: "project.recycle", Title: "回收站", PNode: "project"},
				{Node: "project.manage", Title: "项目管理", PNode: "project"},
				{Node: "project.delete", Title: "项目删除", PNode: "project"},
				{Node: "project.member", Title: "项目成员管理", PNode: "project"},
				{Node: "calendar", Title: "日程", PNode: "project"},
				{Node: "notify.notice", Title: "通知列表", PNode: "project"},
				{Node: "notify.system", Title: "系统消息", PNode: "project"},
				{Node: "members.index", Title: "成员列表", PNode: "project"},
				{Node: "project.manage", Title: "组织管理", PNode: "project"},
			},
		},
		{
			Node:  "task",
			Title: "任务权限",
			Children: []*nodeItem{
				{Node: "task:create", Title: "创建任务", PNode: "task"},
				{Node: "task:edit", Title: "编辑任务", PNode: "task"},
				{Node: "task:assign", Title: "分配任务", PNode: "task"},
				{Node: "task:delete", Title: "删除任务", PNode: "task"},
			},
		},
		{
			Node:  "file",
			Title: "文件权限",
			Children: []*nodeItem{
				{Node: "file:upload", Title: "上传文件", PNode: "file"},
				{Node: "file:delete", Title: "删除文件", PNode: "file"},
			},
		},
		{
			Node:  "system",
			Title: "系统权限",
			Children: []*nodeItem{
				{Node: "system.account", Title: "账号管理", PNode: "system"},
				{Node: "system.menu", Title: "菜单管理", PNode: "system"},
				{Node: "system.node", Title: "节点管理", PNode: "system"},
				{Node: "system.account.auth", Title: "权限管理", PNode: "system"},
			},
		},
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"nodes": nodes}))
}

func (h *HandlerNode) allList(c *gin.Context) {
	result := &common.Result{}
	keyword := strings.TrimSpace(c.PostForm("node"))
	out := make([]nodeItem, 0)
	for _, n := range baseNodes() {
		if keyword == "" || strings.Contains(n.Node, keyword) || strings.Contains(n.Title, keyword) {
			out = append(out, n)
		}
	}
	c.JSON(http.StatusOK, result.Success(out))
}

// nodeAuthRow 节点授权关联表
type nodeAuthRow struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	AuthId int64
	Node   string
}

func (*nodeAuthRow) TableName() string { return "ms_project_auth_node" }

// savedNodeRow 已保存的节点表（复用 ms_project_auth_node， auth_id=0 表示全局节点）
func (h *HandlerNode) save(c *gin.Context) {
	result := &common.Result{}
	nodeStr := c.PostForm("node")
	titleStr := c.PostForm("title")
	if nodeStr == "" {
		c.JSON(http.StatusOK, result.Fail(400, "node不能为空"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	// 检查是否已存在
	var cnt int64
	_ = db.Model(&nodeAuthRow{}).Where("auth_id=0 and node=?", nodeStr).Count(&cnt).Error
	if cnt == 0 {
		_ = db.Create(&nodeAuthRow{AuthId: 0, Node: nodeStr}).Error
	}
	_ = titleStr
	_ = time.Now()
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerNode) clear(c *gin.Context) {
	result := &common.Result{}
	nodeStr := c.PostForm("node")
	db := gorms.GetDB().WithContext(c.Request.Context())
	if nodeStr != "" {
		_ = db.Where("auth_id=0 and node=?", nodeStr).Delete(&nodeAuthRow{}).Error
	} else {
		// 清除所有全局节点
		_ = db.Where("auth_id=0").Delete(&nodeAuthRow{}).Error
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}
