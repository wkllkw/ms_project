package source_link

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerSourceLink struct {
}

func New() *HandlerSourceLink {
	return &HandlerSourceLink{}
}

type sourceLinkRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	TaskCode    int64  `gorm:"column:task_code"`
	MemberCode  int64  `gorm:"column:member_code"`
	Title       string `gorm:"column:title"`
	Url         string `gorm:"column:url"`
	Description string `gorm:"column:description"`
	Sort        int    `gorm:"column:sort"`
	CreateTime  int64  `gorm:"column:create_time"`
}

func (*sourceLinkRow) TableName() string { return "ms_source_link" }

// list 获取资源链接列表
func (h *HandlerSourceLink) list(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")

	if taskCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode不能为空"))
		return
	}

	tid, err := codecs.DecryptInt64(taskCode)
	if err != nil || tid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var rows []sourceLinkRow
	db.Where("task_code=?", tid).Order("sort asc, id desc").Find(&rows)

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"code":        codecs.EncryptInt64(r.Id),
			"taskCode":    taskCode,
			"title":       r.Title,
			"url":         r.Url,
			"description": r.Description,
			"sort":        r.Sort,
			"createTime":  r.CreateTime,
		})
	}

	c.JSON(http.StatusOK, result.Success(out))
}

// save 创建资源链接
func (h *HandlerSourceLink) save(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")
	title := c.PostForm("title")
	url := c.PostForm("url")
	description := c.PostForm("description")
	memberId := c.GetInt64("memberId")

	if taskCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode不能为空"))
		return
	}

	tid, err := codecs.DecryptInt64(taskCode)
	if err != nil || tid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}

	if url == "" {
		c.JSON(http.StatusOK, result.Fail(400, "链接地址不能为空"))
		return
	}

	if title == "" {
		title = url
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 获取当前最大排序
	var maxSort int
	db.Model(&sourceLinkRow{}).Where("task_code=?", tid).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort)

	row := &sourceLinkRow{
		TaskCode:    tid,
		MemberCode:  memberId,
		Title:       title,
		Url:         url,
		Description: description,
		Sort:        maxSort + 1,
		CreateTime:  time.Now().UnixMilli(),
	}

	if err := db.Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":        codecs.EncryptInt64(row.Id),
		"taskCode":    taskCode,
		"title":       row.Title,
		"url":         row.Url,
		"description": row.Description,
		"sort":        row.Sort,
		"createTime":  row.CreateTime,
	}))
}

// edit 编辑资源链接
func (h *HandlerSourceLink) edit(c *gin.Context) {
	result := &common.Result{}
	sourceCode := c.PostForm("sourceCode")
	title := c.PostForm("title")
	url := c.PostForm("url")
	description := c.PostForm("description")

	if sourceCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "sourceCode不能为空"))
		return
	}

	sid, err := codecs.DecryptInt64(sourceCode)
	if err != nil || sid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "sourceCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	updates := map[string]any{}
	if title != "" {
		updates["title"] = title
	}
	if url != "" {
		updates["url"] = url
	}
	if description != "" {
		updates["description"] = description
	}

	if len(updates) == 0 {
		c.JSON(http.StatusOK, result.Success(gin.H{"code": sourceCode}))
		return
	}

	if err := db.Model(&sourceLinkRow{}).Where("id=?", sid).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "编辑失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": sourceCode}))
}

// del 删除资源链接
func (h *HandlerSourceLink) del(c *gin.Context) {
	result := &common.Result{}
	sourceCode := c.PostForm("sourceCode")

	if sourceCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "sourceCode不能为空"))
		return
	}

	sid, err := codecs.DecryptInt64(sourceCode)
	if err != nil || sid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "sourceCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	db.Delete(&sourceLinkRow{}, sid)

	c.JSON(http.StatusOK, result.Success(gin.H{"code": sourceCode}))
}

// 确保model被使用
var _ = model.Page{}
