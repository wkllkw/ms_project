package project_info

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	common "test.com/project-common"
)

type HandlerProjectInfo struct {
}

func New() *HandlerProjectInfo {
	return &HandlerProjectInfo{}
}

type projectInfoRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64  `gorm:"column:project_code"`
	Name        string `gorm:"column:name"`
	Value       string `gorm:"column:value"`
	Description string `gorm:"column:description"`
	Sort        int    `gorm:"column:sort"`
	CreateTime  int64  `gorm:"column:create_time"`
	UpdateTime  int64  `gorm:"column:update_time"`
}

func (*projectInfoRow) TableName() string { return "ms_project_info" }

// list 获取项目信息列表
func (h *HandlerProjectInfo) list(c *gin.Context) {
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
	var rows []projectInfoRow
	db.Where("project_code=?", pid).Order("sort asc, id asc").Find(&rows)

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"code":        codecs.EncryptInt64(r.Id),
			"name":        r.Name,
			"value":       r.Value,
			"description": r.Description,
			"sort":        r.Sort,
			"createTime":  r.CreateTime,
			"projectCode": projectCode,
		})
	}

	c.JSON(http.StatusOK, result.Success(out))
}

// save 创建项目信息
func (h *HandlerProjectInfo) save(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	name := c.PostForm("name")
	value := c.PostForm("value")
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
	db.Model(&projectInfoRow{}).Where("project_code=?", pid).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort)

	row := &projectInfoRow{
		ProjectCode: pid,
		Name:        name,
		Value:       value,
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
		"value":       row.Value,
		"description": row.Description,
		"sort":        row.Sort,
		"createTime":  row.CreateTime,
		"projectCode": projectCode,
	}))
}

// edit 编辑项目信息
func (h *HandlerProjectInfo) edit(c *gin.Context) {
	result := &common.Result{}
	infoCode := c.PostForm("infoCode")
	name := c.PostForm("name")
	value := c.PostForm("value")
	description := c.PostForm("description")

	if infoCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "infoCode不能为空"))
		return
	}

	iid, err := codecs.DecryptInt64(infoCode)
	if err != nil || iid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "infoCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	updates := map[string]any{
		"update_time": time.Now().UnixMilli(),
	}
	if name != "" {
		updates["name"] = name
	}
	if value != "" {
		updates["value"] = value
	}
	if description != "" {
		updates["description"] = description
	}

	if err := db.Model(&projectInfoRow{}).Where("id=?", iid).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "编辑失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": infoCode}))
}

// del 删除项目信息
func (h *HandlerProjectInfo) del(c *gin.Context) {
	result := &common.Result{}
	infoCode := c.PostForm("infoCode")

	if infoCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "infoCode不能为空"))
		return
	}

	iid, err := codecs.DecryptInt64(infoCode)
	if err != nil || iid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "infoCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	db.Delete(&projectInfoRow{}, iid)

	c.JSON(http.StatusOK, result.Success(gin.H{"code": infoCode}))
}
