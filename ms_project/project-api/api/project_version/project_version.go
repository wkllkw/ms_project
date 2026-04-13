package project_version

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	common "test.com/project-common"
)

type HandlerProjectVersion struct {
}

func New() *HandlerProjectVersion {
	return &HandlerProjectVersion{}
}

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

type projectFeaturesRow struct {
	Id          int64 `gorm:"primaryKey"`
	ProjectCode int64 `gorm:"column:project_code"`
}

func (*projectFeaturesRow) TableName() string { return "ms_project_features" }

// list 获取版本列表
func (h *HandlerProjectVersion) list(c *gin.Context) {
	result := &common.Result{}
	featuresCode := c.PostForm("projectFeaturesCode")

	if featuresCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "projectFeaturesCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(featuresCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectFeaturesCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var rows []projectVersionRow
	db.Where("features_code=?", fid).Order("sort asc, id desc").Find(&rows)

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"code":            codecs.EncryptInt64(r.Id),
			"featuresCode":    featuresCode,
			"name":            r.Name,
			"description":     r.Description,
			"startTime":       r.StartTime,
			"planPublishTime": r.PlanPublishTime,
			"publishTime":     r.PublishTime,
			"status":          r.Status,
			"sort":            r.Sort,
			"createTime":      r.CreateTime,
		})
	}

	c.JSON(http.StatusOK, result.Success(out))
}

// save 创建版本
func (h *HandlerProjectVersion) save(c *gin.Context) {
	result := &common.Result{}
	featuresCode := c.PostForm("featuresCode")
	name := c.PostForm("name")
	description := c.PostForm("description")
	startTimeStr := c.PostForm("startTime")
	planPublishTimeStr := c.PostForm("planPublishTime")

	if featuresCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "featuresCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(featuresCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "featuresCode无效"))
		return
	}

	if name == "" {
		c.JSON(http.StatusOK, result.Fail(400, "名称不能为空"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 获取版本库的项目ID
	var features projectFeaturesRow
	if err := db.Where("id=?", fid).First(&features).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "版本库不存在"))
		return
	}

	// 解析时间
	var startTime, planPublishTime int64
	if startTimeStr != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", startTimeStr, time.Local)
		if err == nil {
			startTime = t.UnixMilli()
		}
	}
	if planPublishTimeStr != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", planPublishTimeStr, time.Local)
		if err == nil {
			planPublishTime = t.UnixMilli()
		}
	}

	// 获取当前最大排序
	var maxSort int
	db.Model(&projectVersionRow{}).Where("features_code=?", fid).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort)

	now := time.Now().UnixMilli()
	row := &projectVersionRow{
		FeaturesCode:    fid,
		ProjectCode:     features.ProjectCode,
		Name:            name,
		Description:     description,
		StartTime:       startTime,
		PlanPublishTime: planPublishTime,
		Status:          0,
		Sort:            maxSort + 1,
		CreateTime:      now,
		UpdateTime:      now,
	}

	if err := db.Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":            codecs.EncryptInt64(row.Id),
		"featuresCode":    featuresCode,
		"name":            row.Name,
		"description":     row.Description,
		"startTime":       row.StartTime,
		"planPublishTime": row.PlanPublishTime,
		"status":          row.Status,
		"createTime":      row.CreateTime,
	}))
}

// edit 编辑版本
func (h *HandlerProjectVersion) edit(c *gin.Context) {
	result := &common.Result{}
	versionCode := c.PostForm("versionCode")
	name := c.PostForm("name")
	description := c.PostForm("description")
	startTimeStr := c.PostForm("startTime")
	planPublishTimeStr := c.PostForm("planPublishTime")

	if versionCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode不能为空"))
		return
	}

	vid, err := codecs.DecryptInt64(versionCode)
	if err != nil || vid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode无效"))
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
	if startTimeStr != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", startTimeStr, time.Local)
		if err == nil {
			updates["start_time"] = t.UnixMilli()
		}
	}
	if planPublishTimeStr != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", planPublishTimeStr, time.Local)
		if err == nil {
			updates["plan_publish_time"] = t.UnixMilli()
		}
	}

	if err := db.Model(&projectVersionRow{}).Where("id=?", vid).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "编辑失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": versionCode}))
}

// del 删除版本
func (h *HandlerProjectVersion) del(c *gin.Context) {
	result := &common.Result{}
	versionCode := c.PostForm("versionCode")

	if versionCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode不能为空"))
		return
	}

	vid, err := codecs.DecryptInt64(versionCode)
	if err != nil || vid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	db.Delete(&projectVersionRow{}, vid)

	c.JSON(http.StatusOK, result.Success(gin.H{"code": versionCode}))
}

// changeStatus 更改版本状态（发布）
func (h *HandlerProjectVersion) changeStatus(c *gin.Context) {
	result := &common.Result{}
	versionCode := c.PostForm("versionCode")
	statusStr := c.PostForm("status")

	if versionCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode不能为空"))
		return
	}

	vid, err := codecs.DecryptInt64(versionCode)
	if err != nil || vid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode无效"))
		return
	}

	status := int8(0)
	if statusStr == "1" {
		status = 1
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	updates := map[string]any{
		"status":      status,
		"update_time": time.Now().UnixMilli(),
	}
	if status == 1 {
		updates["publish_time"] = time.Now().UnixMilli()
	}

	if err := db.Model(&projectVersionRow{}).Where("id=?", vid).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "操作失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": versionCode, "status": status}))
}
