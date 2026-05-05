package department

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerDepartment struct {
}

func New() *HandlerDepartment {
	return &HandlerDepartment{}
}

type departmentRow struct {
	Id               int64 `gorm:"primaryKey;autoIncrement"`
	Name             string
	ParentId         int64
	OrganizationCode int64
	Sort             int
	CreateTime       int64
	Deleted          int8
}

func (*departmentRow) TableName() string { return "ms_department" }

func (h *HandlerDepartment) list(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	orgCode := c.GetString("organizationCode")
	orgId, _ := codecs.DecryptInt64(orgCode)
	// 支持 pcode 参数，按父部门过滤（树形懒加载）
	pcode := c.PostForm("pcode")
	var parentId int64
	if pcode != "" {
		parentId, _ = codecs.DecryptInt64(pcode)
	}
	db := gorms.GetDB()
	query := db.Model(&departmentRow{}).Where("deleted=0")
	if orgId != 0 {
		query = query.Where("organization_code=?", orgId)
	}
	// 如果指定了父部门，则只查询该父部门下的子部门
	if pcode != "" {
		query = query.Where("parent_id=?", parentId)
	} else {
		// 未指定父部门时，只返回顶级部门（parent_id=0）
		query = query.Where("parent_id=0")
	}
	var total int64
	_ = query.Count(&total).Error
	var rows []departmentRow
	_ = query.Order("sort asc, id asc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&rows).Error
	// 查询每个部门是否有子部门，用于前端树形展示
	out := make([]gin.H, 0, len(rows))
	for _, d := range rows {
		var childCount int64
		_ = db.Model(&departmentRow{}).Where("deleted=0 and parent_id=?", d.Id).Count(&childCount).Error
		out = append(out, gin.H{
			"id":               d.Id,
			"code":             codecs.EncryptInt64(d.Id),
			"name":             d.Name,
			"parent_id":        codecs.EncryptInt64(d.ParentId),
			"organization_code": codecs.EncryptInt64(d.OrganizationCode),
			"sort":             d.Sort,
			"create_time":      d.CreateTime,
			"hasNext":          childCount > 0,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (h *HandlerDepartment) read(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("departmentCode")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "departmentCode无效"))
		return
	}
	db := gorms.GetDB()
	var d departmentRow
	if err := db.Where("id=?", id).First(&d).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "部门不存在"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"id":          d.Id,
		"code":        codecs.EncryptInt64(d.Id),
		"name":        d.Name,
		"parent_id":   codecs.EncryptInt64(d.ParentId),
		"sort":        d.Sort,
		"create_time": d.CreateTime,
	}))
}

func (h *HandlerDepartment) save(c *gin.Context) {
	result := &common.Result{}
	name := c.PostForm("name")
	parentCode := c.PostForm("parent_id")
	orgCode := c.GetString("organizationCode")
	orgId, _ := codecs.DecryptInt64(orgCode)
	var parentId int64
	if parentCode != "" {
		parentId, _ = codecs.DecryptInt64(parentCode)
	}
	row := &departmentRow{
		Name:             name,
		ParentId:         parentId,
		OrganizationCode: orgId,
		Sort:             0,
		CreateTime:       time.Now().UnixMilli(),
		Deleted:          0,
	}
	if err := gorms.GetDB().Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"code": codecs.EncryptInt64(row.Id)}))
}

func (h *HandlerDepartment) edit(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("departmentCode")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "departmentCode无效"))
		return
	}
	updates := map[string]any{}
	if v := c.PostForm("name"); v != "" {
		updates["name"] = v
	}
	_ = gorms.GetDB().Model(&departmentRow{}).Where("id=?", id).Updates(updates).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerDepartment) del(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("departmentCode")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "departmentCode无效"))
		return
	}
	_ = gorms.GetDB().Model(&departmentRow{}).Where("id=?", id).Updates(map[string]any{"deleted": 1}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

