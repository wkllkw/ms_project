package task_tag

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	common "test.com/project-common"
)

type HandlerTaskTag struct{}

func New() *HandlerTaskTag {
	return &HandlerTaskTag{}
}

type taskTagRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64  `gorm:"column:project_code"`
	Name        string `gorm:"column:name"`
	Color       string `gorm:"column:color"`
	CreateTime  int64  `gorm:"column:create_time"`
	Deleted     int8   `gorm:"column:deleted"`
}

func (*taskTagRow) TableName() string { return "ms_task_tag" }

// list 获取任务标签列表
func (h *HandlerTaskTag) list(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	if projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "请选择一个项目"))
		return
	}

	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}

	db := gorms.GetDB()
	var tags []taskTagRow
	err = db.Where("project_code = ? AND deleted = 0", pid).Order("name ASC").Find(&tags).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "查询失败"))
		return
	}

	list := make([]gin.H, 0, len(tags))
	for _, tag := range tags {
		list = append(list, gin.H{
			"code":        codecs.EncryptInt64(tag.Id),
			"projectCode": codecs.EncryptInt64(tag.ProjectCode),
			"name":        tag.Name,
			"color":       tag.Color,
			"createTime":  tag.CreateTime,
		})
	}

	c.JSON(http.StatusOK, result.Success(list))
}

// save 创建任务标签
func (h *HandlerTaskTag) save(c *gin.Context) {
	result := &common.Result{}
	name := c.PostForm("name")
	projectCode := c.PostForm("projectCode")
	color := c.PostForm("color")

	if name == "" {
		c.JSON(http.StatusOK, result.Fail(400, "请填写标签名称"))
		return
	}

	if projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "请选择项目"))
		return
	}

	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}

	db := gorms.GetDB()

	// 检查标签是否已存在
	var existTag taskTagRow
	err = db.Where("name = ? AND project_code = ? AND deleted = 0", name, pid).First(&existTag).Error
	if err == nil {
		c.JSON(http.StatusOK, result.Fail(400, "该标签已存在"))
		return
	}

	tag := &taskTagRow{
		ProjectCode: pid,
		Name:        name,
		Color:       color,
		CreateTime:  time.Now().UnixMilli(),
		Deleted:     0,
	}

	if err := db.Create(tag).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":        codecs.EncryptInt64(tag.Id),
		"projectCode": codecs.EncryptInt64(tag.ProjectCode),
		"name":        tag.Name,
		"color":       tag.Color,
		"createTime":  tag.CreateTime,
	}))
}

// edit 编辑任务标签
func (h *HandlerTaskTag) edit(c *gin.Context) {
	result := &common.Result{}
	tagCode := c.PostForm("tagCode")
	name := c.PostForm("name")
	color := c.PostForm("color")

	if tagCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "请选择一个标签"))
		return
	}

	if name == "" {
		c.JSON(http.StatusOK, result.Fail(400, "请填写标签名称"))
		return
	}

	id, err := codecs.DecryptInt64(tagCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "tagCode无效"))
		return
	}

	db := gorms.GetDB()

	// 检查标签是否存在
	var tag taskTagRow
	err = db.Where("id = ? AND deleted = 0", id).First(&tag).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "该标签已失效"))
		return
	}

	updates := map[string]interface{}{
		"name":  name,
		"color": color,
	}

	if err := db.Model(&taskTagRow{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(nil))
}

// delete 删除任务标签
func (h *HandlerTaskTag) delete(c *gin.Context) {
	result := &common.Result{}
	tagCode := c.PostForm("tagCode")

	if tagCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "请选择一个标签"))
		return
	}

	id, err := codecs.DecryptInt64(tagCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "tagCode无效"))
		return
	}

	db := gorms.GetDB()

	// 删除标签
	if err := db.Where("id = ?", id).Delete(&taskTagRow{}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "删除失败"))
		return
	}

	// 删除任务标签关联关系
	if err := db.Where("tag_id = ?", id).Delete(&taskTagRelRow{}).Error; err != nil {
		// 忽略关联关系删除错误
	}

	c.JSON(http.StatusOK, result.Success(nil))
}

type taskTagRelRow struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	TaskId int64 `gorm:"column:task_id"`
	TagId  int64 `gorm:"column:tag_id"`
}

func (*taskTagRelRow) TableName() string { return "ms_task_tag_rel" }
