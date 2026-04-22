package account

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerAccount struct {
}

func New() *HandlerAccount {
	return &HandlerAccount{}
}

type memberRow struct {
	Id            int64 `gorm:"primaryKey;autoIncrement"`
	Account       string
	Password      string
	Name          string
	Mobile        string
	Email         string
	Status        int
	Avatar        string
	CreateTime    int64
	LastLoginTime int64
}

func (*memberRow) TableName() string { return "ms_member" }

func (h *HandlerAccount) list(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	db := gorms.GetDB()
	query := db.Model(&memberRow{})
	if keyword := c.PostForm("keyword"); keyword != "" {
		query = query.Where("name like ? or account like ? or email like ? or mobile like ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	_ = query.Count(&total).Error
	var list []memberRow
	_ = query.Order("id desc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&list).Error
	out := make([]gin.H, 0, len(list))
	for _, m := range list {
		out = append(out, gin.H{
			"id":          m.Id,
			"code":        codecs.EncryptInt64(m.Id),
			"name":        m.Name,
			"account":     m.Account,
			"email":       m.Email,
			"mobile":      m.Mobile,
			"avatar":      m.Avatar,
			"status":      m.Status,
			"create_time": m.CreateTime,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total, "authList": []any{}}))
}

func (h *HandlerAccount) allList(c *gin.Context) {
	result := &common.Result{}
	db := gorms.GetDB()
	var list []memberRow
	_ = db.Order("id desc").Limit(300).Find(&list).Error
	out := make([]gin.H, 0, len(list))
	for _, m := range list {
		out = append(out, gin.H{
			"id":     m.Id,
			"code":   codecs.EncryptInt64(m.Id),
			"name":   m.Name,
			"avatar": m.Avatar,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": len(out)}))
}

func (h *HandlerAccount) forbid(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("accountCode")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "accountCode无效"))
		return
	}
	_ = gorms.GetDB().Model(&memberRow{}).Where("id=?", id).Update("status", 0).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAccount) resume(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("accountCode")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "accountCode无效"))
		return
	}
	_ = gorms.GetDB().Model(&memberRow{}).Where("id=?", id).Update("status", 1).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAccount) add(c *gin.Context) {
	result := &common.Result{}
	row := &memberRow{
		Account:    c.PostForm("account"),
		Password:   c.PostForm("password"),
		Name:       c.PostForm("name"),
		Mobile:     c.PostForm("mobile"),
		Email:      c.PostForm("email"),
		Status:     1,
		Avatar:     c.PostForm("avatar"),
		CreateTime: time.Now().UnixMilli(),
	}
	if err := gorms.GetDB().Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"code": codecs.EncryptInt64(row.Id)}))
}

func (h *HandlerAccount) edit(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("code")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	updates := map[string]any{}
	for _, k := range []string{"name", "mobile", "email", "avatar"} {
		if v := c.PostForm(k); v != "" {
			updates[k] = v
		}
	}
	_ = gorms.GetDB().Model(&memberRow{}).Where("id=?", id).Updates(updates).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAccount) auth(c *gin.Context) {
	result := &common.Result{}
	idStr := c.PostForm("id")
	authStr := c.PostForm("auth")
	if idStr == "" {
		c.JSON(http.StatusOK, result.Fail(400, "id必填"))
		return
	}
	id, err := codecs.DecryptInt64(idStr)
	if err != nil || id == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "id无效"))
		return
	}
	// 将权限角色ID存储到 ms_project_member 的 authorize 字段（全局授权）
	db := gorms.GetDB().WithContext(c.Request.Context())
	_ = db.Table("ms_project_member").Where("member_code=?", id).Update("authorize", authStr).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAccount) del(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("accountCode")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "accountCode无效"))
		return
	}
	_ = gorms.GetDB().Where("id=?", id).Delete(&memberRow{}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerAccount) read(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("code")
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	db := gorms.GetDB()
	var m memberRow
	if err := db.Where("id=?", id).First(&m).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "账号不存在"))
		return
	}
	memberCode := codecs.EncryptInt64(m.Id)
	c.JSON(http.StatusOK, result.Success(gin.H{
		"id":          m.Id,
		"code":        memberCode,
		"member_code": memberCode,
		"name":        m.Name,
		"account":     m.Account,
		"email":       m.Email,
		"mobile":      m.Mobile,
		"avatar":      m.Avatar,
		"status":      m.Status,
		"create_time": m.CreateTime,
		"position":    "",
		"departments": "",
		"description": "",
	}))
}

func (h *HandlerAccount) syncDetail(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("code")
	if code == "" {
		c.JSON(http.StatusOK, result.Fail(400, "code必填"))
		return
	}
	id, err := codecs.DecryptInt64(code)
	if err != nil || id == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	var m memberRow
	if err := db.Where("id=?", id).First(&m).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "账号不存在"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"id":          m.Id,
		"code":        codecs.EncryptInt64(m.Id),
		"name":        m.Name,
		"account":     m.Account,
		"email":       m.Email,
		"mobile":      m.Mobile,
		"avatar":      m.Avatar,
		"status":      m.Status,
		"create_time": m.CreateTime,
	}))
}

func (h *HandlerAccount) joinByInviteLink(c *gin.Context) {
	result := &common.Result{}
	inviteCode := c.PostForm("inviteCode")
	if inviteCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "inviteCode必填"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}
