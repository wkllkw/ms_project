package project_collect

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	common "test.com/project-common"
)

type HandlerProjectCollect struct{}

func New() *HandlerProjectCollect {
	return &HandlerProjectCollect{}
}

// projectCollectionRow 项目收藏表结构
type projectCollectionRow struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"column:project_code;index"`
	MemberCode  int64 `gorm:"column:member_code;index"`
	CreateTime  int64 `gorm:"column:create_time"`
}

func (*projectCollectionRow) TableName() string {
	return "ms_project_collection"
}

// projectRow 项目表结构（用于查询）
type projectRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	Code        string `gorm:"column:code"`
	Name        string `gorm:"column:name"`
	Description string `gorm:"column:description"`
	Cover       string `gorm:"column:cover"`
	Deleted     int    `gorm:"column:deleted"`
}

func (*projectRow) TableName() string {
	return "ms_project"
}

// collectProject 收藏/取消收藏项目
func (h *HandlerProjectCollect) collectProject(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	projectCode := c.PostForm("projectCode")
	collectType := c.PostForm("type")

	if projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "请选择项目"))
		return
	}

	if collectType == "" {
		collectType = "collect" // 默认为收藏操作
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 检查项目是否存在
	var project projectRow
	if err := db.Table("ms_project").
		Where("id = ? AND deleted = 0", projectCode).
		First(&project).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "项目不存在或已删除"))
		return
	}

	// 查询是否已收藏
	var existingCollect projectCollectionRow
	err := db.Table("ms_project_collection").
		Where("member_code = ? AND project_code = ?", memberId, projectCode).
		First(&existingCollect).Error

	if collectType == "collect" {
		// 收藏操作
		if err == nil {
			c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "项目已收藏"))
			return
		}

		// 创建收藏记录
		collection := &projectCollectionRow{
			MemberCode:  memberId,
			ProjectCode: parseStringToInt64(projectCode),
			CreateTime:  time.Now().UnixMilli(),
		}

		if err := db.Table("ms_project_collection").Create(collection).Error; err != nil {
			c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "收藏失败"))
			return
		}

		c.JSON(http.StatusOK, result.Success(gin.H{
			"message": "加入收藏成功",
		}))
	} else {
		// 取消收藏操作
		if err != nil {
			c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "尚未收藏该项目"))
			return
		}

		// 删除收藏记录
		if err := db.Table("ms_project_collection").
			Where("member_code = ? AND project_code = ?", memberId, projectCode).
			Delete(nil).Error; err != nil {
			c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "取消收藏失败"))
			return
		}

		c.JSON(http.StatusOK, result.Success(gin.H{
			"message": "取消收藏成功",
		}))
	}
}

// getCollectionList 获取收藏的项目列表
func (h *HandlerProjectCollect) getCollectionList(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 查询用户收藏的项目
	var collections []projectCollectionRow
	if err := db.Table("ms_project_collection").
		Where("member_code = ?", memberId).
		Order("create_time desc").
		Find(&collections).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "查询失败"))
		return
	}

	// 获取项目详情
	var projectList []gin.H
	for _, collection := range collections {
		var project projectRow
		if err := db.Table("ms_project").
			Where("id = ? AND deleted = 0", collection.ProjectCode).
			First(&project).Error; err == nil {
			projectList = append(projectList, gin.H{
				"id":          project.Id,
				"code":        project.Code,
				"name":        project.Name,
				"description": project.Description,
				"cover":       project.Cover,
				"collectTime": collection.CreateTime,
			})
		}
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  projectList,
		"total": len(projectList),
	}))
}

// parseStringToInt64 辅助函数：字符串转int64
func parseStringToInt64(s string) int64 {
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		}
	}
	return result
}
