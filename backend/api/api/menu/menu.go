package menu

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/authz"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/internal/menus"
	"test.com/project-api/pkg/codecs"
	projectMenu "test.com/project-api/pkg/model/menu"
	common "test.com/project-common"
)

type HandlerMenu struct {
}

func New() *HandlerMenu {
	return &HandlerMenu{}
}

func (h *HandlerMenu) menu(c *gin.Context) {
	result := &common.Result{}
	if err := menus.SeedDefaultIfEmpty(c.Request.Context()); err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "初始化菜单失败"))
		return
	}
	pms, err := menus.FindMenus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "查询失败"))
		return
	}

	// 根据用户角色过滤菜单（合并项目级和组织级权限）
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB()

	// 获取当前组织编码
	orgCodeStr, _ := c.Get("organizationCode")
	var orgCode int64
	if orgStr, ok := orgCodeStr.(string); ok && orgStr != "" {
		orgCode, _ = codecs.DecryptInt64(orgStr)
	}

	nodes := authz.GetAllNodes(db, memberId, orgCode)
	allowedNodes := make(map[string]bool, len(nodes))
	for _, n := range nodes {
		allowedNodes[n] = true
	}

	// 如果没有任何权限节点（未分配角色或默认角色无节点），只返回不绑定权限的菜单（node为空或"#"）
	if len(allowedNodes) == 0 {
		filtered := authz.FilterProjectMenusByNodes(pms, allowedNodes)
		tree := menus.BuildTree(filtered)
		c.JSON(http.StatusOK, result.Success(gin.H{
			"menus": tree,
			"nodes": nodes,
		}))
		return
	}

	// 先过滤扁平列表，再构建树
	filtered := authz.FilterProjectMenusByNodes(pms, allowedNodes)
	tree := menus.BuildTree(filtered)

	// 二次过滤树：确保父节点即使 node 不在权限中，只要有可见子节点也保留
	tree = authz.FilterMenusByNodes(tree, allowedNodes)

	// 同时返回用户的权限节点列表，供前端路由守卫使用
	c.JSON(http.StatusOK, result.Success(gin.H{
		"menus": tree,
		"nodes": nodes,
	}))
}

func parseBoolInt(v string) int {
	if v == "1" || v == "true" {
		return 1
	}
	return 0
}

func (h *HandlerMenu) menuAdd(c *gin.Context) {
	result := &common.Result{}
	pm := &menus.ProjectMenu{
		Pid:        parseInt64(c.PostForm("pid")),
		Title:      c.PostForm("title"),
		Icon:       c.PostForm("icon"),
		Url:        c.PostForm("url"),
		FilePath:   c.PostForm("file_path"),
		Params:     c.PostForm("params"),
		Node:       c.PostForm("node"),
		Sort:       int(parseInt64(c.PostForm("sort"))),
		Status:     1,
		CreateBy:   c.GetInt64("memberId"),
		IsInner:    parseBoolInt(c.PostForm("is_inner")),
		Values:     c.PostForm("values"),
		ShowSlider: parseBoolInt(c.PostForm("show_slider")),
	}
	if pm.Sort == 0 {
		pm.Sort = 1
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	if err := db.Create(pm).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	m := &projectMenu.Menu{
		Id:         pm.Id,
		Pid:        pm.Pid,
		Title:      pm.Title,
		Icon:       pm.Icon,
		Url:        pm.Url,
		FilePath:   pm.FilePath,
		Params:     pm.Params,
		Node:       pm.Node,
		Sort:       int32(pm.Sort),
		Status:     int32(pm.Status),
		CreateBy:   pm.CreateBy,
		IsInner:    int32(pm.IsInner),
		Values:     pm.Values,
		ShowSlider: int32(pm.ShowSlider),
		StatusText: "使用中",
		InnerText:  map[int]string{0: "导航", 1: "内页"}[pm.IsInner],
		FullUrl:    pm.Url,
		Children:   []*projectMenu.Menu{},
	}
	if (pm.Params != "" && pm.Values != "") || pm.Values != "" {
		m.FullUrl = pm.Url + "/" + pm.Values
	}
	c.JSON(http.StatusOK, result.Success(m))
}

func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

func (h *HandlerMenu) menuEdit(c *gin.Context) {
	result := &common.Result{}
	id := parseInt64(c.PostForm("id"))
	if id == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "id必填"))
		return
	}
	updates := map[string]any{}
	// 使用 c.Request.ParseMultipartForm 确保 multipart 表单中所有字段都被解析
	// 允许空字符串通过，这样用户可以清空已有值的字段
	_ = c.Request.ParseMultipartForm(32 << 20)
	for _, k := range []string{"title", "icon", "url", "file_path", "params", "node", "values"} {
		if v, ok := c.GetPostForm(k); ok {
			updates[k] = v
		}
	}
	if sortStr := c.PostForm("sort"); sortStr != "" {
		updates["sort"] = int(parseInt64(sortStr))
	}
	if innerStr := c.PostForm("is_inner"); innerStr != "" {
		updates["is_inner"] = parseBoolInt(innerStr)
	}
	if sliderStr := c.PostForm("show_slider"); sliderStr != "" {
		updates["show_slider"] = parseBoolInt(sliderStr)
	}
	if pidStr := c.PostForm("pid"); pidStr != "" {
		updates["pid"] = parseInt64(pidStr)
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	if err := db.Model(&menus.ProjectMenu{}).Where("id=?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerMenu) menuDel(c *gin.Context) {
	result := &common.Result{}
	id := parseInt64(c.PostForm("id"))
	if id == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "id必填"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	// 递归收集所有子孙ID，支持任意层级
	allIDs := []int64{id}
	queue := []int64{id}
	for len(queue) > 0 {
		var childIDs []int64
		_ = db.Model(&menus.ProjectMenu{}).Select("id").Where("pid in ?", queue).Scan(&childIDs).Error
		if len(childIDs) == 0 {
			break
		}
		allIDs = append(allIDs, childIDs...)
		queue = childIDs
	}
	if err := db.Where("id in ?", allIDs).Delete(&menus.ProjectMenu{}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "删除失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerMenu) menuForbid(c *gin.Context) {
	result := &common.Result{}
	id := parseInt64(c.PostForm("id"))
	if id == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "id必填"))
		return
	}
	_ = gorms.GetDB().WithContext(c.Request.Context()).Model(&menus.ProjectMenu{}).Where("id=?", id).Update("status", 0).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerMenu) menuResume(c *gin.Context) {
	result := &common.Result{}
	id := parseInt64(c.PostForm("id"))
	if id == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "id必填"))
		return
	}
	_ = gorms.GetDB().WithContext(c.Request.Context()).Model(&menus.ProjectMenu{}).Where("id=?", id).Update("status", 1).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}
