package project_template

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerProjectTemplate struct{}

func New() *HandlerProjectTemplate {
	return &HandlerProjectTemplate{}
}

// projectTemplateRow 项目模板表结构
type projectTemplateRow struct {
	Id               int64  `gorm:"primaryKey;autoIncrement"`
	Code             string `gorm:"column:code"`
	Name             string `gorm:"column:name"`
	Description      string `gorm:"column:description"`
	Cover            string `gorm:"column:cover"`
	MemberCode       int64  `gorm:"column:member_code"`
	OrganizationCode int64  `gorm:"column:organization_code"`
	IsSystem         int    `gorm:"column:is_system"`
	CreateTime       int64  `gorm:"column:create_time"`
}

func (*projectTemplateRow) TableName() string {
	return "ms_project_template"
}

// taskStagesTemplateRow 任务阶段模板表结构
type taskStagesTemplateRow struct {
	Id                  int64  `gorm:"primaryKey;autoIncrement"`
	Name                string `gorm:"column:name"`
	ProjectTemplateCode int64  `gorm:"column:project_template_code"`
	Sort                int    `gorm:"column:sort"`
	CreateTime          int64  `gorm:"column:create_time"`
}

func (*taskStagesTemplateRow) TableName() string {
	return "ms_task_stages_template"
}

// generateTemplateCode 生成模板编码
func generateTemplateCode() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "tpl_" + hex.EncodeToString(b)
}

// getTemplateList 获取模板列表
func (h *HandlerProjectTemplate) getTemplateList(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	orgCode := orgCodeFromContext(c)

	page := &model.Page{}
	page.Bind(c)

	db := gorms.GetDB().WithContext(c.Request.Context())

	viewType := c.PostForm("viewType") // -1:全部, 0:组织模板, 1:系统模板

	// 查找当前用户所属部门的成员ID列表（用于自定义模板过滤）
	var deptMemberIds []int64
	var deptIds []int64
	_ = db.Table("ms_department_member").Select("DISTINCT department_id").Where("member_id=?", memberId).Scan(&deptIds).Error
	if len(deptIds) > 0 {
		_ = db.Table("ms_department_member").Select("DISTINCT member_id").Where("department_id IN ?", deptIds).Scan(&deptMemberIds).Error
	}

	var templates []projectTemplateRow
	var total int64

	query := db.Table("ms_project_template")

	if viewType == "1" {
		// 只看系统模板
		query = query.Where("is_system = 1")
	} else if viewType == "0" {
		// 只看组织模板 - 添加部门过滤：只显示用户所属部门成员创建的模板
		query = query.Where("organization_code = ? AND is_system = 0", orgCode)
		if len(deptMemberIds) > 0 {
			query = query.Where("member_code IN ?", deptMemberIds)
		}
	} else {
		// 全部：系统模板 + 组织模板（组织模板添加部门过滤）
		if len(deptMemberIds) > 0 {
			query = query.Where("organization_code = ? AND is_system = 0 AND member_code IN ? OR is_system = 1", orgCode, deptMemberIds)
		} else {
			query = query.Where("organization_code = ? OR is_system = 1", orgCode)
		}
	}

	// 获取总数
	query.Count(&total)

	// 分页查询
	offset := (page.Page - 1) * page.PageSize
	if err := query.Order("create_time desc").
		Offset(int(offset)).
		Limit(int(page.PageSize)).
		Find(&templates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "查询失败"))
		return
	}

	// 为每个模板添加任务阶段信息
	templateList := make([]gin.H, 0)
	for _, template := range templates {
		var stages []taskStagesTemplateRow
		db.Table("ms_task_stages_template").
			Where("project_template_code = ?", template.Id).
			Order("sort desc, id asc").
			Find(&stages)

		stageNames := []string{}
		for _, stage := range stages {
			stageNames = append(stageNames, stage.Name)
		}

		templateList = append(templateList, gin.H{
			"id":               template.Id,
			"code":             template.Code,
			"name":             template.Name,
			"description":      template.Description,
			"cover":            template.Cover,
			"memberCode":       template.MemberCode,
			"organizationCode": template.OrganizationCode,
			"isSystem":         template.IsSystem,
			"createTime":       template.CreateTime,
			"taskStages":       stageNames,
		})
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  templateList,
		"total": total,
	}))
}

// createTemplate 创建项目模板
func (h *HandlerProjectTemplate) createTemplate(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	orgCode := orgCodeFromContext(c)

	name := c.PostForm("name")
	description := c.PostForm("description")
	cover := c.PostForm("cover")

	if name == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "请填写模板名称"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 创建模板
	template := &projectTemplateRow{
		Code:             generateTemplateCode(),
		Name:             name,
		Description:      description,
		Cover:            cover,
		MemberCode:       memberId,
		OrganizationCode: orgCode,
		IsSystem:         0,
		CreateTime:       time.Now().UnixMilli(),
	}

	if err := db.Table("ms_project_template").Create(template).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "创建模板失败"))
		return
	}

	// 创建默认任务阶段
	defaultStages := []string{"待办", "进行中", "已完成"}
	for i, stageName := range defaultStages {
		stage := &taskStagesTemplateRow{
			Name:                stageName,
			ProjectTemplateCode: template.Id,
			Sort:                len(defaultStages) - i,
			CreateTime:          time.Now().UnixMilli(),
		}
		db.Table("ms_task_stages_template").Create(stage)
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"message": "制作模板成功",
		"code":    template.Code,
		"data":    template,
	}))
}

// editTemplate 编辑项目模板
func (h *HandlerProjectTemplate) editTemplate(c *gin.Context) {
	result := &common.Result{}

	code := c.PostForm("code")
	name := c.PostForm("name")
	description := c.PostForm("description")
	cover := c.PostForm("cover")

	if code == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "请选择模板"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 检查模板是否存在
	var template projectTemplateRow
	if err := db.Table("ms_project_template").
		Where("code = ?", code).
		First(&template).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "该模板已失效"))
		return
	}

	// 检查是否为系统模板
	if template.IsSystem == 1 {
		c.JSON(http.StatusOK, result.Fail(http.StatusForbidden, "无法编辑系统模板"))
		return
	}

	// 更新模板信息
	updates := map[string]interface{}{}
	if name != "" {
		updates["name"] = name
	}
	if description != "" {
		updates["description"] = description
	}
	if cover != "" {
		updates["cover"] = cover
	}

	if len(updates) > 0 {
		if err := db.Table("ms_project_template").
			Where("code = ?", code).
			Updates(updates).Error; err != nil {
			c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "编辑失败"))
			return
		}
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"message": "编辑模板成功",
	}))
}

// deleteTemplate 删除项目模板
func (h *HandlerProjectTemplate) deleteTemplate(c *gin.Context) {
	result := &common.Result{}

	code := c.PostForm("code")
	if code == "" {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "请选择模板"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 检查模板是否存在
	var template projectTemplateRow
	if err := db.Table("ms_project_template").
		Where("code = ?", code).
		First(&template).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "该模板不存在"))
		return
	}

	// 检查是否为系统模板
	if template.IsSystem == 1 {
		c.JSON(http.StatusOK, result.Fail(http.StatusForbidden, "无法删除系统模板"))
		return
	}

	// 删除模板的任务阶段
	db.Table("ms_task_stages_template").
		Where("project_template_code = ?", template.Id).
		Delete(nil)

	// 删除模板
	if err := db.Table("ms_project_template").
		Where("code = ?", code).
		Delete(nil).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "删除失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(""))
}

// uploadCover 上传模板封面
func (h *HandlerProjectTemplate) uploadCover(c *gin.Context) {
	result := &common.Result{}

	// 获取上传的文件
	file, err := c.FormFile("cover")
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "请上传封面图片"))
		return
	}

	// 保存文件（这里简化处理，实际应该上传到文件服务器）
	filePath := "/uploads/template_covers/" + file.Filename
	if err := c.SaveUploadedFile(file, "."+filePath); err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusInternalServerError, "上传失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"path": filePath,
		"url":  filePath,
	}))
}

// orgCodeFromContext 从 gin 上下文中获取解密后的 organizationCode
func orgCodeFromContext(c *gin.Context) int64 {
	orgVal, _ := c.Get("organizationCode")
	switch t := orgVal.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case string:
		decrypted, err := codecs.DecryptInt64(t)
		if err == nil && decrypted > 0 {
			return decrypted
		}
		i, _ := strconv.ParseInt(t, 10, 64)
		return i
	default:
		return 0
	}
}
