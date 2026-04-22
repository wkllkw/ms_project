package menu

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/internal/menus"
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
	c.JSON(http.StatusOK, result.Success(menus.BuildTree(pms)))
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
	for _, k := range []string{"title", "icon", "url", "file_path", "params", "node", "values"} {
		if v := c.PostForm(k); v != "" {
			if k == "file_path" {
				updates["file_path"] = v
			} else {
				updates[k] = v
			}
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
	var childIDs []int64
	_ = db.Model(&menus.ProjectMenu{}).Select("id").Where("pid=?", id).Scan(&childIDs).Error
	var grandChildIDs []int64
	if len(childIDs) > 0 {
		_ = db.Model(&menus.ProjectMenu{}).Select("id").Where("pid in ?", childIDs).Scan(&grandChildIDs).Error
	}
	ids := make([]int64, 0, 1+len(childIDs)+len(grandChildIDs))
	ids = append(ids, id)
	ids = append(ids, childIDs...)
	ids = append(ids, grandChildIDs...)
	if err := db.Where("id in ?", ids).Delete(&menus.ProjectMenu{}).Error; err != nil {
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
