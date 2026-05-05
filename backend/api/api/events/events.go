package events

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

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

type HandlerEvents struct {
}

func New() *HandlerEvents {
	return &HandlerEvents{}
}

type eventsRow struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64
	Title       string
	Description string
	BeginTime   string
	EndTime     string
	AllDay      int8
	Position    string
	CreateBy    int64
	CreateTime  int64
	Deleted     int8
}

func (*eventsRow) TableName() string { return "ms_project_events" }

type eventsMemberRow struct {
	Id       int64 `gorm:"primaryKey;autoIncrement"`
	EventsId int64
	MemberId int64
	Status   int8
}

func (*eventsMemberRow) TableName() string { return "ms_project_events_member" }

func (h *HandlerEvents) list(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	projectCode := c.PostForm("projectCode")
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	deleted := c.PostForm("deleted")
	db := gorms.GetDB()
	query := db.Model(&eventsRow{}).Where("project_code=?", pid)
	if deleted == "1" {
		query = query.Where("deleted=1")
	} else {
		query = query.Where("deleted=0")
	}
	var total int64
	_ = query.Count(&total).Error
	var rows []eventsRow
	_ = query.Order("id desc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, ev := range rows {
		out = append(out, h.toResponse(db, ev))
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (h *HandlerEvents) myList(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	memberId := c.GetInt64("memberId")
	orgCode := orgCodeFromContext(c)
	deleted := c.PostForm("deleted")
	db := gorms.GetDB()
	query := db.Model(&eventsRow{}).Joins("join ms_project_events_member m on m.events_id=ms_project_events.id").Where("m.member_id=?", memberId)
	// 组织过滤：只显示当前组织下的日程
	if orgCode != 0 {
		query = query.Where("ms_project_events.project_code IN (SELECT id FROM ms_project WHERE organization_code=? AND deleted=0)", orgCode)
	}
	if deleted == "1" {
		query = query.Where("ms_project_events.deleted=1")
	} else {
		query = query.Where("ms_project_events.deleted=0")
	}
	var total int64
	_ = query.Count(&total).Error
	var rows []eventsRow
	_ = query.Select("ms_project_events.*").Order("ms_project_events.id desc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, ev := range rows {
		out = append(out, h.toResponse(db, ev))
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (h *HandlerEvents) toResponse(db *gorm.DB, ev eventsRow) gin.H {
	var members []eventsMemberRow
	_ = db.Where("events_id=?", ev.Id).Find(&members).Error
	memberList := make([]gin.H, 0, len(members))
	for _, m := range members {
		// 查询成员信息
		var memberInfo struct {
			Name   string `json:"name"`
			Avatar string `json:"avatar"`
		}
		_ = db.Table("ms_member").Where("id = ?", m.MemberId).Select("name, avatar").First(&memberInfo).Error
		memberList = append(memberList, gin.H{
			"member_code": codecs.EncryptInt64(m.MemberId),
			"status":      m.Status,
			"is_owner":    m.MemberId == ev.CreateBy,
			"memberInfo": gin.H{
				"name":   memberInfo.Name,
				"avatar": memberInfo.Avatar,
			},
		})
	}
	return gin.H{
		"id":           ev.Id,
		"code":         codecs.EncryptInt64(ev.Id),
		"title":        ev.Title,
		"description":  ev.Description,
		"begin_time":   ev.BeginTime,
		"end_time":     ev.EndTime,
		"all_day":      ev.AllDay == 1,
		"position":     ev.Position,
		"project_code": codecs.EncryptInt64(ev.ProjectCode),
		"created_by":   codecs.EncryptInt64(ev.CreateBy),
		"memberList":   memberList,
		"waitConfirm":  0,
	}
}

func (h *HandlerEvents) save(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("project_code")
	if projectCode == "" {
		projectCode = c.PostForm("projectCode")
	}
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	title := c.PostForm("title")
	beginTime := c.PostForm("begin_time")
	endTime := c.PostForm("end_time")
	allDayStr := c.PostForm("all_day")
	allDay := int8(0)
	if allDayStr == "1" || allDayStr == "true" {
		allDay = 1
	}
	row := &eventsRow{
		ProjectCode: pid,
		Title:       title,
		Description: c.PostForm("description"),
		BeginTime:   beginTime,
		EndTime:     endTime,
		AllDay:      allDay,
		Position:    c.PostForm("position"),
		CreateBy:    c.GetInt64("memberId"),
		CreateTime:  time.Now().UnixMilli(),
		Deleted:     0,
	}
	db := gorms.GetDB()
	memberListRaw := c.PostForm("member_list")
	var memberCodes []string
	_ = json.Unmarshal([]byte(memberListRaw), &memberCodes)
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		for _, mc := range memberCodes {
			mid, err := codecs.DecryptInt64(mc)
			if err != nil {
				continue
			}
			if err := tx.Create(&eventsMemberRow{EventsId: row.Id, MemberId: mid, Status: 1}).Error; err != nil {
				return err
			}
		}
		if len(memberCodes) == 0 {
			tx.Create(&eventsMemberRow{EventsId: row.Id, MemberId: c.GetInt64("memberId"), Status: 1})
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(h.toResponse(db, *row)))
}

func (h *HandlerEvents) edit(c *gin.Context) {
	result := &common.Result{}
	idStr := c.PostForm("id")
	var id int64
	_ = json.Unmarshal([]byte(idStr), &id)
	if id == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "id无效"))
		return
	}
	updates := map[string]any{}
	if v := c.PostForm("title"); v != "" {
		updates["title"] = v
	}
	if v := c.PostForm("description"); v != "" {
		updates["description"] = v
	}
	if v := c.PostForm("begin_time"); v != "" {
		updates["begin_time"] = v
	}
	if v := c.PostForm("end_time"); v != "" {
		updates["end_time"] = v
	}
	if v := c.PostForm("position"); v != "" {
		updates["position"] = v
	}
	if v := c.PostForm("all_day"); v != "" {
		allDay := int8(0)
		if v == "1" || v == "true" {
			allDay = 1
		}
		updates["all_day"] = allDay
	}
	db := gorms.GetDB()
	_ = db.Model(&eventsRow{}).Where("id=?", id).Updates(updates).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerEvents) del(c *gin.Context) {
	result := &common.Result{}
	eventsCode := c.PostForm("eventsCode")
	id, err := codecs.DecryptInt64(eventsCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "eventsCode无效"))
		return
	}
	db := gorms.GetDB()
	_ = db.Model(&eventsRow{}).Where("id=?", id).Updates(map[string]any{"deleted": 1}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerEvents) read(c *gin.Context) {
	result := &common.Result{}
	eventsCode := c.PostForm("eventsCode")
	id, err := codecs.DecryptInt64(eventsCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "eventsCode无效"))
		return
	}
	db := gorms.GetDB()
	var row eventsRow
	if err := db.Where("id=?", id).First(&row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "日程不存在"))
		return
	}
	c.JSON(http.StatusOK, result.Success(h.toResponse(db, row)))
}

func (h *HandlerEvents) confirmJoin(c *gin.Context) {
	result := &common.Result{}
	eventsCode := c.PostForm("eventsCode")
	statusStr := c.PostForm("status")
	eventsId, err := codecs.DecryptInt64(eventsCode)
	if err != nil || eventsId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "eventsCode无效"))
		return
	}
	status, err := strconv.ParseInt(statusStr, 10, 8)
	if err != nil || (status != 1 && status != 2) {
		c.JSON(http.StatusOK, result.Fail(400, "status无效"))
		return
	}
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB()
	err = db.Model(&eventsMemberRow{}).Where("events_id=? and member_id=?", eventsId, memberId).Update("status", int8(status)).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerEvents) getEventsListByCalendar(c *gin.Context) {
	result := &common.Result{}
	dateStr := c.PostForm("date")
	memberCodesRaw := c.PostForm("memberCodes")
	memberId := c.GetInt64("memberId") // 当前登录用户ID
	orgCode := orgCodeFromContext(c)

	// 解析日期，获取月份的第一天和最后一天
	parsedDate, err := time.Parse("2006-01-02 15:04:05", dateStr)
	if err != nil {
		// 尝试其他格式
		parsedDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusOK, result.Success(gin.H{"list": gin.H{}}))
			return
		}
	}
	// 获取月份的第一天和最后一天
	firstOfMonth := time.Date(parsedDate.Year(), parsedDate.Month(), 1, 0, 0, 0, 0, parsedDate.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// 解析成员列表
	var memberCodes []string
	if memberCodesRaw != "" && memberCodesRaw != "[]" {
		_ = json.Unmarshal([]byte(memberCodesRaw), &memberCodes)
	}

	db := gorms.GetDB()
	// 查询日程
	query := db.Model(&eventsRow{}).Where("deleted=0")
	// 日期范围筛选
	query = query.Where("begin_time >= ? AND begin_time <= ?", firstOfMonth.Format("2006-01-02 15:04:05"), lastOfMonth.Format("2006-01-02 15:04:05"))

	// 成员筛选
	if len(memberCodes) > 0 {
		var memberIds []int64
		for _, mc := range memberCodes {
			if mid, err := codecs.DecryptInt64(mc); err == nil {
				memberIds = append(memberIds, mid)
			}
		}
		if len(memberIds) > 0 {
			query = query.Joins("JOIN ms_project_events_member mem ON mem.events_id = ms_project_events.id").
				Where("mem.member_id IN ?", memberIds)
		}
	} else if memberId > 0 {
		// 如果没有指定成员筛选，默认查询当前用户的日程
		query = query.Joins("JOIN ms_project_events_member mem ON mem.events_id = ms_project_events.id").
			Where("mem.member_id = ?", memberId)
	}

	// 组织过滤：只显示当前组织下的日程
	if orgCode != 0 {
		query = query.Where("ms_project_events.project_code IN (SELECT id FROM ms_project WHERE organization_code=? AND deleted=0)", orgCode)
	}

	var rows []eventsRow
	if err := query.Order("begin_time asc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusOK, result.Success(gin.H{"list": gin.H{}}))
		return
	}

	// 按日期分组
	grouped := make(map[string][]gin.H)
	for _, ev := range rows {
		// 解析开始时间，获取日期部分（兼容有/无秒的格式）
		beginTime, err := time.Parse("2006-01-02 15:04:05", ev.BeginTime)
		if err != nil {
			beginTime, err = time.Parse("2006-01-02 15:04", ev.BeginTime)
			if err != nil {
				continue
			}
		}
		dateKey := beginTime.Format("2006-01-02")
		if _, ok := grouped[dateKey]; !ok {
			grouped[dateKey] = []gin.H{}
		}
		grouped[dateKey] = append(grouped[dateKey], h.toResponse(db, ev))
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": grouped}))
}

func (h *HandlerEvents) getEventsLog(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	projectCode := c.PostForm("projectCode")
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Success(gin.H{"list": []any{}, "total": 0}))
		return
	}
	db := gorms.GetDB()
	var rows []eventsRow
	var total int64
	query := db.Model(&eventsRow{}).Where("project_code=? and deleted=0", projectId)
	query.Count(&total)
	_ = query.Order("id desc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Find(&rows).Error
	list := make([]gin.H, 0, len(rows))
	for _, ev := range rows {
		item := h.toResponse(db, ev)
		// 添加日志相关信息
		item["log_type"] = "created"
		item["log_time"] = ev.CreateTime
		list = append(list, item)
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": list, "total": total}))
}

