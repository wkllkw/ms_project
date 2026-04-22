package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	common "test.com/project-common"
)

type HandlerAuth struct {
}

func New() *HandlerAuth {
	return &HandlerAuth{}
}

type authRow struct {
	Id            int64  `gorm:"primaryKey;autoIncrement"`
	Title         string `gorm:"column:title"`
	Organization  int64  `gorm:"column:organization"`
	Status        int    `gorm:"column:status"`
	IsDefault     int    `gorm:"column:is_default"`
	CreateAt      int64  `gorm:"column:create_at;index"`
	Description   string `gorm:"column:description"`
}

func (*authRow) TableName() string { return "ms_project_auth" }

type authNodeRow struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	AuthId int64 `gorm:"column:auth_id;index"`
	Node   string `gorm:"column:node"`
}

func (*authNodeRow) TableName() string { return "ms_project_auth_node" }

func (h *HandlerAuth) list(c *gin.Context) {
	result := &common.Result{}
	db := gorms.GetDB()
	organization := c.PostForm("organization")
	var rows []authRow
	query := db.Model(&authRow{})
	if organization != "" {
		query = query.Where("organization=?", organization)
	}
	_ = query.Order("id desc").Find(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":          r.Id,
			"title":       r.Title,
			"organization": r.Organization,
			"status":      r.Status,
			"is_default":  r.IsDefault,
			"create_at":   r.CreateAt,
			"description": r.Description,
		})
	}
	c.JSON(http.StatusOK, result.Success(out))
}

func (h *HandlerAuth) add(c *gin.Context) {
	result := &common.Result{}
	title := c.PostForm("title")
	organization := c.PostForm("organization")
	description := c.PostForm("description")
	if title == "" {
		c.JSON(http.StatusOK, result.Fail(400, "标题不能为空"))
		return
	}
	db := gorms.GetDB()
	row := &authRow{
		Title:       title,
		Organization: 0,
		Status:      1,
		IsDefault:   0,
		CreateAt:    0,
		Description: description,
	}
	if organization != "" {
		// 可选：解析 organization
	}
	if err := db.Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"id": row.Id}))
}

func (h *HandlerAuth) edit(c *gin.Context) {
	result := &common.Result{}
	id := c.PostForm("id")
	if id == "" {
		c.JSON(http.StatusOK, result.Fail(400, "id不能为空"))
		return
	}
	updates := map[string]any{}
	if v := c.PostForm("title"); v != "" {
		updates["title"] = v
	}
	if v := c.PostForm("description"); v != "" {
		updates["description"] = v
	}
	db := gorms.GetDB()
	_ = db.Model(&authRow{}).Where("id=?", id).Updates(updates).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAuth) apply(c *gin.Context) {
	result := &common.Result{}
	id := c.PostForm("id")
	action := c.PostForm("action")
	nodes := c.PostForm("nodes")
	if id == "" {
		c.JSON(http.StatusOK, result.Fail(400, "id不能为空"))
		return
	}
	db := gorms.GetDB()
	// 先删除旧的节点关联
	_ = db.Where("auth_id=?", id).Delete(&authNodeRow{}).Error
	// 保存新的节点关联
	if nodes != "" {
		// nodes 是逗号分隔的节点字符串
		// 这里简化处理，实际可能需要解析JSON
		node := &authNodeRow{AuthId: 0, Node: nodes}
		_ = db.Create(node).Error
	}
	_ = action // 避免未使用变量警告
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAuth) forbid(c *gin.Context) {
	result := &common.Result{}
	id := c.PostForm("id")
	status := c.PostForm("status")
	if id == "" {
		c.JSON(http.StatusOK, result.Fail(400, "id不能为空"))
		return
	}
	db := gorms.GetDB()
	statusInt := 0
	if status == "1" {
		statusInt = 1
	}
	_ = db.Model(&authRow{}).Where("id=?", id).Update("status", statusInt).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAuth) resume(c *gin.Context) {
	result := &common.Result{}
	id := c.PostForm("id")
	status := c.PostForm("status")
	if id == "" {
		c.JSON(http.StatusOK, result.Fail(400, "id不能为空"))
		return
	}
	db := gorms.GetDB()
	statusInt := 1
	if status == "0" {
		statusInt = 0
	}
	_ = db.Model(&authRow{}).Where("id=?", id).Update("status", statusInt).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAuth) setDefault(c *gin.Context) {
	result := &common.Result{}
	id := c.PostForm("id")
	isDefault := c.PostForm("is_default")
	if id == "" {
		c.JSON(http.StatusOK, result.Fail(400, "id不能为空"))
		return
	}
	db := gorms.GetDB()
	isDefaultInt := 0
	if isDefault == "1" {
		isDefaultInt = 1
	}
	_ = db.Model(&authRow{}).Where("id=?", id).Update("is_default", isDefaultInt).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAuth) del(c *gin.Context) {
	result := &common.Result{}
	id := c.PostForm("id")
	if id == "" {
		c.JSON(http.StatusOK, result.Fail(400, "id不能为空"))
		return
	}
	db := gorms.GetDB()
	_ = db.Where("id=?", id).Delete(&authRow{}).Error
	_ = db.Where("auth_id=?", id).Delete(&authNodeRow{}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}
