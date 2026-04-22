package department_member

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerDepartmentMember struct {
}

func New() *HandlerDepartmentMember {
	return &HandlerDepartmentMember{}
}

type departmentMemberRow struct {
	Id           int64 `gorm:"primaryKey;autoIncrement"`
	DepartmentId int64
	MemberId     int64
}

func (*departmentMemberRow) TableName() string { return "ms_department_member" }

type memberRow struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	Name   string
	Avatar string
	Account string
	Email  string
}

func (*memberRow) TableName() string { return "ms_member" }

func (h *HandlerDepartmentMember) index(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	departmentCode := c.PostForm("departmentCode")
	depId, err := codecs.DecryptInt64(departmentCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "departmentCode无效"))
		return
	}
	db := gorms.GetDB()
	var total int64
	_ = db.Model(&departmentMemberRow{}).Where("department_id=?", depId).Count(&total).Error
	var rows []departmentMemberRow
	_ = db.Where("department_id=?", depId).Limit(int(page.PageSize)).Offset(int((page.Page-1)*page.PageSize)).Find(&rows).Error
	memberIds := make([]int64, 0, len(rows))
	for _, r := range rows {
		memberIds = append(memberIds, r.MemberId)
	}
	memberMap := map[int64]memberRow{}
	if len(memberIds) > 0 {
		var ms []memberRow
		_ = db.Where("id in ?", memberIds).Find(&ms).Error
		for _, m := range ms {
			memberMap[m.Id] = m
		}
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		m := memberMap[r.MemberId]
		out = append(out, gin.H{
			"id":   r.Id,
			"code": codecs.EncryptInt64(m.Id),
			"name": m.Name,
			"avatar": m.Avatar,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (h *HandlerDepartmentMember) detail(c *gin.Context) {
	h.index(c)
}

func (h *HandlerDepartmentMember) searchInviteMember(c *gin.Context) {
	result := &common.Result{}
	keyword := c.PostForm("keyword")
	db := gorms.GetDB()
	var members []memberRow
	q := db.Model(&memberRow{})
	if keyword != "" {
		q = q.Where("name like ? or account like ? or email like ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	_ = q.Limit(50).Find(&members).Error
	out := make([]gin.H, 0, len(members))
	for _, m := range members {
		out = append(out, gin.H{
			"code":   codecs.EncryptInt64(m.Id),
			"name":   m.Name,
			"avatar": m.Avatar,
		})
	}
	c.JSON(http.StatusOK, result.Success(out))
}

func (h *HandlerDepartmentMember) inviteMember(c *gin.Context) {
	result := &common.Result{}
	accountCode := c.PostForm("accountCode")
	departmentCode := c.PostForm("departmentCode")
	memberId, err := codecs.DecryptInt64(accountCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "accountCode无效"))
		return
	}
	depId, err := codecs.DecryptInt64(departmentCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "departmentCode无效"))
		return
	}
	db := gorms.GetDB()
	var count int64
	_ = db.Model(&departmentMemberRow{}).Where("department_id=? and member_id=?", depId, memberId).Count(&count).Error
	if count == 0 {
		_ = db.Create(&departmentMemberRow{DepartmentId: depId, MemberId: memberId}).Error
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerDepartmentMember) removeMember(c *gin.Context) {
	result := &common.Result{}
	accountCode := c.PostForm("accountCode")
	departmentCode := c.PostForm("departmentCode")
	memberId, _ := codecs.DecryptInt64(accountCode)
	depId, _ := codecs.DecryptInt64(departmentCode)
	_ = gorms.GetDB().Where("department_id=? and member_id=?", depId, memberId).Delete(&departmentMemberRow{}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

