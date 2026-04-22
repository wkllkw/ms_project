package email

import (
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"
)

// Config 邮件配置
type Config struct {
	Host     string // SMTP 主机地址
	Port     int    // SMTP 端口
	User     string // 发件人账号
	Password string // 发件人密码/授权码
	From     string // 发件人名称
}

var cfg *Config

// Init 初始化邮件模块
func Init(c *Config) {
	cfg = c
}

// IsAvailable 检查邮件服务是否可用
func IsAvailable() bool {
	return cfg != nil && cfg.Host != ""
}

// Send 发送邮件
func Send(to, subject, htmlBody string) error {
	if !IsAvailable() {
		return fmt.Errorf("email service not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(cfg.User, cfg.From))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.User, cfg.Password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send email failed: %w", err)
	}
	return nil
}

// SendTaskAssigned 发送任务分配通知邮件
func SendTaskAssigned(toEmail, toName, taskName, projectName, assignerName string) error {
	subject := fmt.Sprintf("[%s] 您被分配了新任务: %s", projectName, taskName)
	body := renderTaskEmail(taskName, projectName, assignerName, toName, "assigned")
	return Send(toEmail, subject, body)
}

// SendTaskCommented 发送任务评论通知邮件
func SendTaskCommented(toEmail, toName, taskName, projectName, commenterName, commentContent string) error {
	subject := fmt.Sprintf("[%s] 任务有新评论: %s", projectName, taskName)
	body := renderCommentEmail(taskName, projectName, commenterName, toName, commentContent)
	return Send(toEmail, subject, body)
}

// SendTaskReminder 发送任务到期提醒邮件
func SendTaskReminder(toEmail, toName, taskName, projectName string, deadline int64) error {
	subject := fmt.Sprintf("[%s] 任务即将到期: %s", projectName, taskName)
	body := renderReminderEmail(taskName, projectName, toName, deadline)
	return Send(toEmail, subject, body)
}

// ==================== 邮件模板（HTML） ====================

const emailBaseStyle = `
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; margin: 0; padding: 20px; }
  .container { max-width: 600px; margin: 0 auto; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 12px rgba(0,0,0,0.1); }
  .header { background: linear-gradient(135deg, #1890ff, #096dd9); padding: 24px; color: #fff; }
  .header h1 { margin: 0; font-size: 20px; }
  .content { padding: 24px; color: #333; line-height: 1.6; }
  .field-label { color: #888; font-size: 13px; margin-bottom: 4px; }
  .field-value { font-weight: 500; margin-bottom: 16px; }
  .btn { display: inline-block; background: #1890ff; color: #fff; text-decoration: none; padding: 10px 24px; border-radius: 4px; margin-top: 8px; }
  .footer { background: #fafafa; padding: 16px 24px; text-align: center; color: #999; font-size: 12px; border-top: 1px solid #eee; }
</style>
`

func renderBaseHTML(bodyContent string) string {
	return fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="utf-8">%s</head><body><div class="container"><div class="header"><h1>MS Project 协作平台</h1></div><div class="content">%s</div><div class="footer"><p>此邮件由系统自动发送，请勿直接回复。</p><p>Powered by MS Project Task Collaboration System</p></div></div></body></html>`, emailBaseStyle, bodyContent)
}

func renderTaskEmail(taskName, project, assigner, recipient, action string) string {
	actionText := map[string]string{
		"assigned": "将您分配到了该任务",
	}[action]
	if actionText == "" {
		actionText = "通知您关于该任务"
	}

	content := fmt.Sprintf(`
<p>Hi %s，</p>
<p>%s 在项目 <strong>%s</strong> 中%s：</p>
<div class="field-label">任务名称</div>
<div class="field-value">%s</div>
<a class="btn" href="#">查看任务详情</a>
<p style="color:#999;font-size:13px;margin-top:20px;">如有疑问，请联系任务分配人或项目管理员。</p>
`, recipient, assigner, project, actionText, taskName)

	return renderBaseHTML(content)
}

func renderCommentEmail(taskName, project, commenter, recipient, comment string) string {
	// 截断过长评论
	if len(comment) > 200 {
		comment = comment[:200] + "..."
	}
	comment = strings.ReplaceAll(comment, "\n", "<br>")

	content := fmt.Sprintf(`
<p>Hi %s，</p>
<p>%s 在项目 <strong>%s</strong> 的任务 <strong>%s</strong> 中发表了新评论：</p>
<div style="background:#f5f5f5;padding:12px;border-radius:4px;margin:12px 0;color:#555;">%s</div>
<a class="btn" href="#">查看评论并回复</a>
`, recipient, commenter, project, taskName, comment)

	return renderBaseHTML(content)
}

func renderReminderEmail(taskName, project, recipient string, deadline int64) string {
	deadlineStr := formatTimestamp(deadline)

	content := fmt.Sprintf(`
<p>Hi %s，</p>
<p>您在项目 <strong>%s</strong> 中的任务即将到期，请注意及时处理：</p>
<div class="field-label">任务名称</div>
<div class="field-value">%s</div>
<div class="field-label">截止时间</div>
<div class="field-value" style="color:#ff4d4f;">%s</div>
<a class="btn" href="#">立即处理任务</a>
<p style="color:#999;font-size:13px;margin-top:20px;">此为系统自动提醒，请根据实际情况安排工作进度。</p>
`, recipient, project, taskName, deadlineStr)

	return renderBaseHTML(content)
}

func formatTimestamp(ts int64) string {
	t := (ts / 1000)
	// 简单格式化
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d",
		(t/31536000)+1970,
		(t%31536000)/2592000+1,
		(t%2592000)/86400+1,
		(t%86400)/3600,
		(t%3600)/60,
	)
}
