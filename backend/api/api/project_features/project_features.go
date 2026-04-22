package project_features

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	common "test.com/project-common"
)

type HandlerProjectFeatures struct {
}

func New() *HandlerProjectFeatures {
	return &HandlerProjectFeatures{}
}

type projectFeaturesRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64  `gorm:"column:project_code"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	Sort        int    `gorm:"column:sort"`
	CreateTime  int64  `gorm:"column:create_time"`
	UpdateTime  int64  `gorm:"column:update_time"`
}

func (*projectFeaturesRow) TableName() string { return "ms_project_features" }

// list 获取版本库列表
func (h *HandlerProjectFeatures) list(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")

	if projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode不能为空"))
		return
	}

	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil || pid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var rows []projectFeaturesRow
	db.Where("project_code=?", pid).Order("sort asc, id asc").Find(&rows)

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"code":        codecs.EncryptInt64(r.Id),
			"name":        r.Name,
			"description": r.Description,
			"sort":        r.Sort,
			"createTime":  r.CreateTime,
			"projectCode": projectCode,
		})
	}

	c.JSON(http.StatusOK, result.Success(out))
}

// save 创建版本库
func (h *HandlerProjectFeatures) save(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	name := c.PostForm("name")
	description := c.PostForm("description")

	if projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode不能为空"))
		return
	}

	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil || pid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}

	if name == "" {
		c.JSON(http.StatusOK, result.Fail(400, "名称不能为空"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	now := time.Now().UnixMilli()

	// 获取当前最大排序
	var maxSort int
	db.Model(&projectFeaturesRow{}).Where("project_code=?", pid).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort)

	row := &projectFeaturesRow{
		ProjectCode: pid,
		Name:        name,
		Description: description,
		Sort:        maxSort + 1,
		CreateTime:  now,
		UpdateTime:  now,
	}

	if err := db.Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":        codecs.EncryptInt64(row.Id),
		"name":        row.Name,
		"description": row.Description,
		"sort":        row.Sort,
		"createTime":  row.CreateTime,
		"projectCode": projectCode,
	}))
}

// edit 编辑版本库
func (h *HandlerProjectFeatures) edit(c *gin.Context) {
	result := &common.Result{}
	featuresCode := c.PostForm("featuresCode")
	name := c.PostForm("name")
	description := c.PostForm("description")

	if featuresCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "featuresCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(featuresCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "featuresCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	updates := map[string]any{
		"update_time": time.Now().UnixMilli(),
	}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}

	if err := db.Model(&projectFeaturesRow{}).Where("id=?", fid).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "编辑失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": featuresCode}))
}

// del 删除版本库
func (h *HandlerProjectFeatures) del(c *gin.Context) {
	result := &common.Result{}
	featuresCode := c.PostForm("featuresCode")

	if featuresCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "featuresCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(featuresCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "featuresCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 删除版本库下的所有版本
	db.Where("features_code=?", fid).Delete(&projectVersionRow{})

	// 删除版本库
	db.Delete(&projectFeaturesRow{}, fid)

	c.JSON(http.StatusOK, result.Success(gin.H{"code": featuresCode}))
}

// projectVersionRow 版本表结构
type projectVersionRow struct {
	Id              int64  `gorm:"primaryKey;autoIncrement"`
	FeaturesCode    int64  `gorm:"column:features_code"`
	ProjectCode     int64  `gorm:"column:project_code"`
	Name            string `gorm:"column:name"`
	Description     string `gorm:"column:description"`
	StartTime       int64  `gorm:"column:start_time"`
	PlanPublishTime int64  `gorm:"column:plan_publish_time"`
	PublishTime     int64  `gorm:"column:publish_time"`
	Status          int8   `gorm:"column:status"`
	Sort            int    `gorm:"column:sort"`
	CreateTime      int64  `gorm:"column:create_time"`
	UpdateTime      int64  `gorm:"column:update_time"`
}

func (*projectVersionRow) TableName() string { return "ms_project_version" }
