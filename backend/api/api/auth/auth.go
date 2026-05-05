package auth

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	Id        int64  `gorm:"primaryKey;autoIncrement"`
	Title     string `gorm:"column:title"`
	Desc      string `gorm:"column:desc"`
	Status    int    `gorm:"column:status"`
	IsDefault int    `gorm:"column:is_default"`
	CreateAt  int64  `gorm:"column:create_at;index"`
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
	var rows []authRow
	_ = db.Model(&authRow{}).Order("id desc").Find(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":         r.Id,
			"title":      r.Title,
			"desc":       r.Desc,
			"status":     r.Status,
			"is_default": r.IsDefault,
			"create_at":  r.CreateAt,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": len(out)}))
}

func (h *HandlerAuth) add(c *gin.Context) {
	result := &common.Result{}
	title := c.PostForm("title")
	desc := c.PostForm("desc")
	if title == "" {
		c.JSON(http.StatusOK, result.Fail(400, "标题不能为空"))
		return
	}
	db := gorms.GetDB()
	row := &authRow{
		Title:     title,
		Desc:      desc,
		Status:    1,
		IsDefault: 0,
		CreateAt:  time.Now().UnixMilli(),
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
	if v := c.PostForm("desc"); v != "" {
		updates["desc"] = v
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
	authId, _ := strconv.ParseInt(id, 10, 64)
	if authId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "id无效"))
		return
	}
	db := gorms.GetDB()
	// 查询操作：只读取，不修改
	if action == "getnode" || action == "get" {
		var rows []authNodeRow
		db.Where("auth_id=?", authId).Find(&rows)
		checkedNodes := make([]string, 0, len(rows))
		for _, r := range rows {
			if r.Node != "" {
				checkedNodes = append(checkedNodes, r.Node)
			}
		}
		c.JSON(http.StatusOK, result.Success(gin.H{
			"list":        checkedNodes,
			"checkedList": checkedNodes,
		}))
		return
	}
	// 保存操作：先删除旧的节点关联，再保存新的
	_ = db.Where("auth_id=?", authId).Delete(&authNodeRow{}).Error
	if nodes != "" {
		// nodes 可能是逗号分隔的节点字符串，也可能是 JSON 数组字符串
		var nodeList []string
		if strings.HasPrefix(nodes, "[") {
			_ = json.Unmarshal([]byte(nodes), &nodeList)
		} else {
			nodeList = strings.Split(nodes, ",")
		}
		for _, node := range nodeList {
			node = strings.TrimSpace(node)
			if node != "" {
				_ = db.Create(&authNodeRow{AuthId: authId, Node: node}).Error
			}
		}
	}
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
	// 设置为默认时，先取消其他默认角色，确保只有一个默认角色
	if isDefaultInt == 1 {
		_ = db.Model(&authRow{}).Where("is_default=1").Update("is_default", 0).Error
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
