package project_member

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerProjectMember struct {
}

func New() *HandlerProjectMember {
	return &HandlerProjectMember{}
}

type projectMemberRow struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64
	MemberCode  int64
	JoinTime    int64
	IsOwner     int64
	Authorize   string
}

func (*projectMemberRow) TableName() string { return "ms_project_member" }

type memberRow struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	Name   string
	Avatar string
	Email  string
	Mobile string
	Account string
	Status int
}

func (*memberRow) TableName() string { return "ms_member" }

type inviteLinkRow struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	ProjectId  int64
	InviteCode string
	ExpiredAt  int64
	CreateBy   int64
	CreateTime int64
}

func (*inviteLinkRow) TableName() string { return "ms_invite_link" }

func (h *HandlerProjectMember) index(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	projectCode := c.PostForm("projectCode")
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	db := gorms.GetDB()
	var total int64
	_ = db.Model(&projectMemberRow{}).Where("project_code=?", pid).Count(&total).Error
	var list []projectMemberRow
	_ = db.Where("project_code=?", pid).Order("id asc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&list).Error
	memberIds := make([]int64, 0, len(list))
	for _, pm := range list {
		memberIds = append(memberIds, pm.MemberCode)
	}
	memberMap := map[int64]memberRow{}
	if len(memberIds) > 0 {
		var members []memberRow
		_ = db.Where("id in ?", memberIds).Find(&members).Error
		for _, m := range members {
			memberMap[m.Id] = m
		}
	}
	out := make([]gin.H, 0, len(list))
	for _, pm := range list {
		m := memberMap[pm.MemberCode]
		out = append(out, gin.H{
			"id":          pm.Id,
			"code":        codecs.EncryptInt64(pm.MemberCode),
			"memberCode":  codecs.EncryptInt64(pm.MemberCode),
			"name":        m.Name,
			"avatar":      m.Avatar,
			"join_time":   pm.JoinTime,
			"is_owner":    pm.IsOwner,
			"authorize":   pm.Authorize,
			"status":      m.Status,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (h *HandlerProjectMember) searchInviteMember(c *gin.Context) {
	result := &common.Result{}
	keyword := c.PostForm("keyword")
	projectCode := c.PostForm("projectCode")
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	db := gorms.GetDB()
	var joined []projectMemberRow
	_ = db.Where("project_code=?", pid).Find(&joined).Error
	joinedMap := map[int64]struct{}{}
	for _, pm := range joined {
		joinedMap[pm.MemberCode] = struct{}{}
	}
	var members []memberRow
	q := db.Model(&memberRow{})
	if keyword != "" {
		q = q.Where("name like ? or account like ? or email like ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	_ = q.Limit(50).Find(&members).Error
	out := make([]gin.H, 0, len(members))
	for _, m := range members {
		_, joined := joinedMap[m.Id]
		out = append(out, gin.H{
			"code":      codecs.EncryptInt64(m.Id),
			"memberCode": codecs.EncryptInt64(m.Id),
			"name":      m.Name,
			"avatar":    m.Avatar,
			"joined":    joined,
		})
	}
	c.JSON(http.StatusOK, result.Success(out))
}

func (h *HandlerProjectMember) inviteMember(c *gin.Context) {
	result := &common.Result{}
	memberCode := c.PostForm("memberCode")
	projectCode := c.PostForm("projectCode")
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	mid, err := codecs.DecryptInt64(memberCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "memberCode无效"))
		return
	}
	db := gorms.GetDB()
	var count int64
	_ = db.Model(&projectMemberRow{}).Where("project_code=? and member_code=?", pid, mid).Count(&count).Error
	if count > 0 {
		c.JSON(http.StatusOK, result.Success([]int{}))
		return
	}
	pm := &projectMemberRow{
		ProjectCode: pid,
		MemberCode:  mid,
		JoinTime:    time.Now().UnixMilli(),
		IsOwner:     0,
		Authorize:   "",
	}
	if err := db.Create(pm).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "邀请失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerProjectMember) removeMember(c *gin.Context) {
	result := &common.Result{}
	memberCode := c.PostForm("memberCode")
	projectCode := c.PostForm("projectCode")
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	mid, err := codecs.DecryptInt64(memberCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "memberCode无效"))
		return
	}
	db := gorms.GetDB()
	_ = db.Where("project_code=? and member_code=?", pid, mid).Delete(&projectMemberRow{}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerProjectMember) listForInvite(c *gin.Context) {
	c.JSON(http.StatusOK, (&common.Result{}).Success(gin.H{"list": []any{}, "total": 0}))
}

func (h *HandlerProjectMember) joinByInviteLink(c *gin.Context) {
	result := &common.Result{}
	inviteCode := c.PostForm("inviteCode")
	if inviteCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "inviteCode必填"))
		return
	}
	db := gorms.GetDB()
	var link inviteLinkRow
	if err := db.Where("invite_code=?", inviteCode).First(&link).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "邀请链接不存在"))
		return
	}
	if link.ExpiredAt != 0 && link.ExpiredAt < time.Now().UnixMilli() {
		c.JSON(http.StatusOK, result.Fail(400, "邀请链接已过期"))
		return
	}
	memberId := c.GetInt64("memberId")
	var count int64
	_ = db.Model(&projectMemberRow{}).Where("project_code=? and member_code=?", link.ProjectId, memberId).Count(&count).Error
	if count == 0 {
		_ = db.Create(&projectMemberRow{
			ProjectCode: link.ProjectId,
			MemberCode:  memberId,
			JoinTime:    time.Now().UnixMilli(),
			IsOwner:     0,
		}).Error
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

