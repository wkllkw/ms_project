package authz

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
)

func IsProjectMember(db *gorm.DB, memberId, projectId int64) bool {
	var count int64
	_ = db.Table("ms_project_member").Where("member_code=? AND project_code=?", memberId, projectId).Count(&count).Error
	return count > 0
}

func IsProjectOwner(db *gorm.DB, memberId, projectId int64) bool {
	var count int64
	_ = db.Table("ms_project_member").Where("member_code=? AND project_code=? AND is_owner=1", memberId, projectId).Count(&count).Error
	return count > 0
}

func CanOperateProject(c *gin.Context, projectCode string) (int64, bool) {
	memberId := c.GetInt64("memberId")
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		return 0, false
	}
	db := gorms.GetDB()
	if !IsProjectMember(db, memberId, projectId) {
		return projectId, false
	}
	return projectId, true
}

func CanManageProject(c *gin.Context, projectCode string) (int64, bool) {
	memberId := c.GetInt64("memberId")
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		return 0, false
	}
	db := gorms.GetDB()
	if !IsProjectOwner(db, memberId, projectId) {
		return projectId, false
	}
	return projectId, true
}

func CanOperateTask(c *gin.Context, taskCode string) (int64, int64, bool) {
	memberId := c.GetInt64("memberId")
	taskId, err := codecs.DecryptInt64(taskCode)
	if err != nil || taskId == 0 {
		return 0, 0, false
	}
	db := gorms.GetDB()
	var row struct {
		ProjectCode int64 `gorm:"column:project_code"`
	}
	_ = db.Table("ms_task").Select("project_code").Where("id=?", taskId).Scan(&row).Error
	if row.ProjectCode == 0 {
		return taskId, 0, false
	}
	if !IsProjectMember(db, memberId, row.ProjectCode) {
		return taskId, row.ProjectCode, false
	}
	return taskId, row.ProjectCode, true
}
