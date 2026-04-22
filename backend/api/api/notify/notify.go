package notify

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerNotify struct {
}

func New() *HandlerNotify {
	return &HandlerNotify{}
}

type notifyRow struct {
	Id         int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	MemberCode int64 `json:"member_code"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Type       int8   `json:"type"`
	IsRead     int8   `json:"is_read"`
	CreateTime int64  `json:"create_time"`
	Action     string `json:"action"`
	SendData   string `json:"send_data" gorm:"column:send_data"`
}

func (*notifyRow) TableName() string { return "ms_notify" }

func typeToInt(t string) int8 {
	switch t {
	case "notice":
		return 1
	case "message":
		return 0
	default:
		return 1
	}
}

func formatTime(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).Format("2006-01-02 15:04")
}

func (h *HandlerNotify) list(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	memberId := c.GetInt64("memberId")
	t := c.PostForm("type")
	db := gorms.GetDB()
	query := db.Model(&notifyRow{}).Where("member_code=?", memberId)
	if t != "" {
		query = query.Where("type=?", typeToInt(t))
	}
	if title := c.PostForm("title"); title != "" {
		query = query.Where("title like ?", "%"+title+"%")
	}
	var total int64
	_ = query.Count(&total).Error
	var list []notifyRow
	_ = query.Order("id desc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&list).Error
	// 格式化输出
	out := make([]gin.H, 0, len(list))
	for _, n := range list {
		out = append(out, gin.H{
			"id":          n.Id,
			"member_code": n.MemberCode,
			"title":       n.Title,
			"content":     n.Content,
			"type":        n.Type,
			"is_read":     n.IsRead,
			"create_time": formatTime(n.CreateTime),
			"action":      n.Action,
			"send_data":   n.SendData,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (h *HandlerNotify) noReads(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB()
	var notices []notifyRow
	_ = db.Where("member_code=? and type=? and is_read=0", memberId, 1).Order("id desc").Limit(10).Find(&notices).Error
	var messages []notifyRow
	_ = db.Where("member_code=? and type=? and is_read=0", memberId, 0).Order("id desc").Limit(10).Find(&messages).Error
	var noticeTotal int64
	_ = db.Model(&notifyRow{}).Where("member_code=? and type=? and is_read=0", memberId, 1).Count(&noticeTotal).Error
	var messageTotal int64
	_ = db.Model(&notifyRow{}).Where("member_code=? and type=? and is_read=0", memberId, 0).Count(&messageTotal).Error
	// 格式化输出
	formatList := func(list []notifyRow) []gin.H {
		out := make([]gin.H, 0, len(list))
		for _, n := range list {
			out = append(out, gin.H{
				"id":          n.Id,
				"member_code": n.MemberCode,
				"title":       n.Title,
				"content":     n.Content,
				"type":        n.Type,
				"is_read":     n.IsRead,
				"create_time": formatTime(n.CreateTime),
				"action":      n.Action,
				"send_data":   n.SendData,
			})
		}
		return out
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list": gin.H{
			"message": formatList(messages),
			"notice":  formatList(notices),
		},
		"total": int(noticeTotal + messageTotal),
		"totalSum": gin.H{
			"message": messageTotal,
			"notice":  noticeTotal,
		},
	}))
}

func (h *HandlerNotify) clearAll(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB()
	_ = db.Model(&notifyRow{}).Where("member_code=?", memberId).Updates(map[string]any{"is_read": 1}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerNotify) save(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	row := &notifyRow{
		MemberCode: memberId,
		Title:      c.PostForm("title"),
		Content:    c.PostForm("content"),
		Type:       typeToInt(c.PostForm("type")),
		IsRead:     0,
		CreateTime: time.Now().UnixMilli(),
		Action:     c.PostForm("action"),
		SendData:   c.PostForm("send_data"),
	}
	if row.SendData == "" {
		row.SendData = "{}"
	}
	if err := gorms.GetDB().Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(row))
}

func (h *HandlerNotify) edit(c *gin.Context) {
	result := &common.Result{}
	id := c.PostForm("id")
	var pid int64
	_ = json.Unmarshal([]byte(id), &pid)
	updates := map[string]any{}
	if v := c.PostForm("title"); v != "" {
		updates["title"] = v
	}
	if v := c.PostForm("content"); v != "" {
		updates["content"] = v
	}
	_ = gorms.GetDB().Model(&notifyRow{}).Where("id=?", pid).Updates(updates).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerNotify) del(c *gin.Context) {
	result := &common.Result{}
	idStr := c.PostForm("id")
	var id int64
	_ = json.Unmarshal([]byte(idStr), &id)
	_ = gorms.GetDB().Where("id=?", id).Delete(&notifyRow{}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerNotify) batchDel(c *gin.Context) {
	result := &common.Result{}
	raw := c.PostForm("ids")
	var ids []int64
	_ = json.Unmarshal([]byte(raw), &ids)
	if len(ids) > 0 {
		_ = gorms.GetDB().Where("id in ?", ids).Delete(&notifyRow{}).Error
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerNotify) setReadied(c *gin.Context) {
	result := &common.Result{}
	raw := c.PostForm("ids")
	var ids []int64
	_ = json.Unmarshal([]byte(raw), &ids)
	if len(ids) > 0 {
		_ = gorms.GetDB().Model(&notifyRow{}).Where("id in ?", ids).Updates(map[string]any{"is_read": 1}).Error
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

