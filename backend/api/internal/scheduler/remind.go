package scheduler

import (
	"fmt"
	"log"
	"time"

	ws "test.com/project-api/api/websocket"
	"test.com/project-api/internal/cache"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/internal/email"
)

// StartTaskReminder 启动任务到期提醒定时任务
// 每 5 分钟扫描一次即将到期和已逾期的任务
func StartTaskReminder() {
	ticker := time.NewTicker(5 * time.Minute)
	log.Println("[Scheduler] Task reminder started, interval: 5min")

	// 首次启动立即执行一次
	checkTaskDeadlines()

	for range ticker.C {
		checkTaskDeadlines()
	}
}

type taskInfo struct {
	Id       int64  `gorm:"column:id"`
	Name     string `gorm:"column:name"`
	EndTime  int64  `gorm:"column:end_time"`
	Done     int8   `gorm:"column:done"`
	AssignTo string `gorm:"column:assign_to"`
	Code     string `gorm:"column:code"` // 加密后的 ID
}

// checkTaskDeadlines 检查任务截止时间
func checkTaskDeadlines() {
	now := time.Now().UnixMilli()
	dayLater := now + (24 * time.Hour).Milliseconds()

	db := gorms.GetDB()
	if db == nil {
		return
	}

	var tasks []taskInfo

	// 查询：未完成 + 未删除 + 截止时间在未来24小时内
	err := db.Table("ms_task").
		Select("id, name, end_time, done, assign_to").
		Where("end_time > ? AND end_time <= ? AND done = 0 AND deleted = 0", now, dayLater).
		Find(&tasks).Error

	if err != nil {
		fmt.Printf("[Scheduler] query upcoming tasks error: %v\n", err)
		return
	}

	// 处理即将到期的任务
	for _, t := range tasks {
		remindKey := fmt.Sprintf("task:reminded:%d", t.Id)

		// 用 Redis 记录已提醒的任务，避免重复通知
		if cache.IsAvailable() {
			exists, _ := cache.Exists(remindKey)
			if exists {
				continue // 已提醒过，跳过
			}
			_ = cache.Set(remindKey, "1", 23*time.Hour) // 23小时后过期（下次循环可能再提醒）
		}

		// 创建通知记录
		createNotifyRecord(t.Id, t.AssignTo, "deadline_reminder",
			fmt.Sprintf("任务「%s」将在 %s 后截止，请及时处理", t.Name, formatDuration(t.EndTime-now)))

		// 如果邮件服务可用，发送提醒邮件
		sendReminderEmail(t)
	}

	// 同时检查逾期任务（超过截止时间仍未完成的）
	var overdueTasks []taskInfo
	err = db.Table("ms_task").
		Select("id, name, end_time, done, assign_to").
		Where("end_time < ? AND end_time > 0 AND done = 0 AND deleted = 0", now).
		Find(&overdueTasks).Error

	if err != nil {
		return
	}

	for _, t := range overdueTasks {
		overdueKey := fmt.Sprintf("task:overdue:%d", t.Id)
		if cache.IsAvailable() {
			exists, _ := cache.Exists(overdueKey)
			if exists {
				continue
			}
			_ = cache.Set(overdueKey, "1", 6*time.Hour) // 6小时内不再重复逾期告警
		}

		createNotifyRecord(t.Id, t.AssignTo, "overdue",
			fmt.Sprintf("任务「%s」已逾期，请尽快处理！", t.Name))

		sendOverdueEmail(t)
	}
}

type notifyRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	Title       string `gorm:"column:title"`
	Content     string `gorm:"column:content"`
	Type        string `gorm:"column:type"`
	FromId      int64  `gorm:"column:from_id"`
	TargetId    int64  `gorm:"column:target_id"`
	IsRead      int8   `gorm:"column:is_read"`
	CreateTime  int64  `gorm:"column:create_time"`
}

func (*notifyRow) TableName() string { return "ms_notify" }

// createNotifyRecord 创建通知记录
func createNotifyRecord(taskId int64, assignTo, notifyType, content string) {
	db := gorms.GetDB()
	now := time.Now().UnixMilli()
	row := &notifyRow{
		Title:      map[string]string{"deadline_reminder": "任务到期提醒", "overdue": "任务逾期告警"}[notifyType],
		Content:    content,
		Type:       notifyType,
		TargetId:   taskId,
		IsRead:     0,
		CreateTime: now,
	}
	if err := db.Create(row).Error; err != nil {
		fmt.Printf("[Scheduler] create notify record error: %v\n", err)
	}
}

// sendReminderEmail 发送到期提醒邮件
func sendReminderEmail(task taskInfo) {
	if !email.IsAvailable() || task.AssignTo == "" {
		return
	}

	// 查询用户邮箱
	type memberInfo struct {
		Email   string `gorm:"column:email"`
		Name    string `gorm:"column:name"`
		Realname string `gorm:"column:realname"`
	}
	var member memberInfo
	db := gorms.GetDB()
	if err := db.Table("ms_member").Select("email, name, realname").Where("account = ?", task.AssignTo).First(&member).Error; err != nil {
		return
	}

	displayName := member.Realname
	if displayName == "" {
		displayName = member.Name
	}
	if displayName == "" {
		displayName = task.AssignTo
	}

	if err := email.SendTaskReminder(member.Email, displayName, task.Name, "", task.EndTime); err != nil {
		fmt.Printf("[Scheduler] send reminder email error: %v\n", err)
	}
}

// sendOverdueEmail 发送逾期告警邮件
func sendOverdueEmail(task taskInfo) {
	if !email.IsAvailable() || task.AssignTo == "" {
		return
	}

	type memberInfo struct {
		Email   string `gorm:"column:email"`
		Name    string `gorm:"column:name"`
		Realname string `gorm:"column:realname"`
	}
	var member memberInfo
	db := gorms.GetDB()
	if err := db.Table("ms_member").Select("email, name, realname").Where("account = ?", task.AssignTo).First(&member).Error; err != nil {
		return
	}

	subject := fmt.Sprintf("[MS Project] 任务「%s」已逾期，请立即处理！", task.Name)
	body := fmt.Sprintf(
		"<p>Hi %s，</p><p>您负责的任务 <strong>%s</strong> 已超过截止时间，请尽快处理！</p>",
		member.Realname+member.Name, task.Name,
	)
	if err := email.Send(member.Email, subject, body); err != nil {
		fmt.Printf("[Scheduler] send overdue email error: %v\n", err)
	}
}

// formatDuration 将毫秒数格式化为可读的时间差
func formatDuration(ms int64) string {
	minutes := ms / 60000
	if minutes < 60 {
		return fmt.Sprintf("%d分钟", minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes%60)
	}
	days := hours / 24
	return fmt.Sprintf("%d天%d小时", days, hours%24)
}

// BroadcastNotify 通过 WebSocket 广播通知消息
func BroadcastNotify(action string, data interface{}) {
	msg := ws.Message{Action: action, Data: data}
	ws.Manager.Broadcast(msg)
}
