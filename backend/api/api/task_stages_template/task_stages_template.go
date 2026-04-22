package task_stages_template

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerTaskStagesTemplate struct {
}

func New() *HandlerTaskStagesTemplate {
	return &HandlerTaskStagesTemplate{}
}

type row struct {
	Id                  int64 `gorm:"primaryKey;autoIncrement"`
	Name                string
	ProjectTemplateCode int64
	CreateTime          int64
	Sort                int
}

func (*row) TableName() string { return "ms_task_stages_template" }

func (h *HandlerTaskStagesTemplate) list(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	templateCode := c.PostForm("project_template_code")
	if templateCode == "" {
		templateCode = c.PostForm("projectTemplateCode")
	}
	db := gorms.GetDB()
	query := db.Model(&row{})
	if templateCode != "" {
		tid, err := codecs.DecryptInt64(templateCode)
		if err == nil {
			query = query.Where("project_template_code=?", tid)
		}
	}
	var total int64
	_ = query.Count(&total).Error
	var list []row
	_ = query.Order("sort desc, id asc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&list).Error
	out := make([]gin.H, 0, len(list))
	for _, r := range list {
		out = append(out, gin.H{
			"id":                    r.Id,
			"code":                  codecs.EncryptInt64(r.Id),
			"name":                  r.Name,
			"project_template_code": codecs.EncryptInt64(r.ProjectTemplateCode),
			"sort":                  r.Sort,
			"create_time":           r.CreateTime,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (h *HandlerTaskStagesTemplate) save(c *gin.Context) {
	result := &common.Result{}
	name := c.PostForm("name")
	templateCode := c.PostForm("project_template_code")
	if templateCode == "" {
		templateCode = c.PostForm("projectTemplateCode")
	}
	tid, err := codecs.DecryptInt64(templateCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectTemplateCode无效"))
		return
	}
	db := gorms.GetDB()
	var maxSort int
	_ = db.Model(&row{}).Where("project_template_code=?", tid).Select("coalesce(max(sort),0)").Scan(&maxSort).Error
	sort := maxSort + 1
	if sortStr := c.PostForm("sort"); sortStr != "" {
		if parsedSort, parseErr := strconv.Atoi(sortStr); parseErr == nil {
			sort = parsedSort
		}
	}
	r := &row{Name: name, ProjectTemplateCode: tid, CreateTime: time.Now().UnixMilli(), Sort: sort}
	if err := db.Create(r).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"code": codecs.EncryptInt64(r.Id)}))
}

func (h *HandlerTaskStagesTemplate) edit(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("code")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	updates := map[string]any{}
	if v := c.PostForm("name"); v != "" {
		updates["name"] = v
	}
	if sortStr := c.PostForm("sort"); sortStr != "" {
		if sort, parseErr := strconv.Atoi(sortStr); parseErr == nil {
			updates["sort"] = sort
		}
	}
	_ = gorms.GetDB().Model(&row{}).Where("id=?", id).Updates(updates).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTaskStagesTemplate) del(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("code")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	_ = gorms.GetDB().Where("id=?", id).Delete(&row{}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}
