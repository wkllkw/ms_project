package testutil

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 测试用数据库配置（从环境变量读取，默认连本地）
var testDSN = func() string {
	user := envOr("TEST_DB_USER", "root")
	pass := envOr("TEST_DB_PASS", "123456")
	host := envOr("TEST_DB_HOST", "127.0.0.1")
	port := envOr("TEST_DB_PORT", "3306")
	dbName := envOr("TEST_DB_NAME", "ms_project_test")
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbName)
}()

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// NewTestDB 创建一个独立的测试数据库连接
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(mysql.Open(testDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Skipf("跳过测试：无法连接测试数据库 (%v)", err)
		return nil
	}

	if err := autoMigrate(db); err != nil {
		t.Fatalf("迁移测试库失败: %v", err)
	}

	return db
}

// Truncate 清理测试数据
func Truncate(db *gorm.DB, tables ...string) {
	for _, t := range tables {
		db.Exec("DELETE FROM " + t)
	}
}

// MustExecSQL 执行原生 SQL 插入测试种子数据
func MustExecSQL(db *gorm.DB, sql string, args ...interface{}) {
	if err := db.Exec(sql, args...).Error; err != nil {
		panic(fmt.Sprintf("seed data failed: %v", err))
	}
}

// ========================
// 测试用 GORM 模型（严格对齐生产代码中的结构体）
// ========================

type testProject struct {
	Id           int64  `gorm:"primaryKey;autoIncrement"`
	Name         string `gorm:"size:255;default:''"`
	Description  string `gorm:"type:text"`
	Cover        string `gorm:"size:500;default:''"`
	TemplateCode int    `gorm:"column:template_code;default:0"`
	CreateTime   int64  `gorm:"column:create_time;default:0"`
	Deleted      int8   `gorm:"default:0"`
}

func (*testProject) TableName() string { return "ms_project" }

type testProjectMember struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"column:project_code;index"`
	MemberCode  int64 `gorm:"column:member_code;index"`
	JoinTime    int64 `gorm:"column:join_time"`
	IsOwner     int8  `gorm:"column:is_owner"`
}

func (*testProjectMember) TableName() string { return "ms_project_member" }

type testTaskStages struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"column:project_code"`
	Name        string
	Sort        int
	CreateTime  int64 `gorm:"column:create_time"`
	Deleted     int8
}

func (*testTaskStages) TableName() string { return "ms_task_stages" }

type testTask struct {
	Id           int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode  int64  `gorm:"column:project_code"`
	Name         string `gorm:"size:500"`
	Description  string `gorm:"type:text"`
	Status       int8
	Priority     int8
	BeginTime    int64 `gorm:"column:begin_time"`
	EndTime      int64 `gorm:"column:end_time"`
	CreateTime   int64 `gorm:"column:create_time"`
	MemberCode   int64 `gorm:"column:member_code"`
	OwnerCode    int64 `gorm:"column:owner_code"`
	AssignTo     int64 `gorm:"column:assign_to"`
	StageCode    int64 `gorm:"column:stage_code"`
	ParentTaskId int64 `gorm:"column:parent_task_id"`
	VersionCode  int64 `gorm:"column:version_code;default:0"`
	Sort         int
	Deleted      int8
	Private      int8
	Done         int8
	DoneTime     int64 `gorm:"column:done_time;default:0"`
	LikeCount    int   `gorm:"column:like;default:0"`
	Star         int   `gorm:"default:0"`
	WorkTime     int64 `gorm:"column:work_time;default:0"`
}

func (*testTask) TableName() string { return "ms_task" }

type testTaskComment struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	TaskId     int64 `gorm:"column:task_id;index"`
	MemberId   int64 `gorm:"column:member_id"`
	Comment    string `gorm:"type:text"`
	CreateTime int64  `gorm:"column:create_time"`
}

func (*testTaskComment) TableName() string { return "ms_task_comment" }

type testTaskWorkTime struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	TaskId     int64 `gorm:"column:task_id;index"`
	MemberId   int64 `gorm:"column:member_id"`
	WorkTime   int64 `gorm:"column:work_time"`
	Remark     string
	CreateTime int64 `gorm:"column:create_time"`
}

func (*testTaskWorkTime) TableName() string { return "ms_task_work_time" }

// testProjectEvent 对齐 task.go 中的 taskLogRow:
// EventType string, MemberCode 无 column 覆盖 → 默认 member_code
type testProjectEvent struct {
	Id           int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode  int64  `gorm:"column:project_code;index"`
	MemberCode   int64  `gorm:"column:member_code;index"`
	TaskId       int64  `gorm:"column:task_id;index;default:0"`
	EventType    string `gorm:"column:event_type"`
	EventContent string `gorm:"column:event_content;type:text"`
	CreateTime   int64  `gorm:"column:create_time"`
}

func (*testProjectEvent) TableName() string { return "ms_project_event" }

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&testProject{},
		&testProjectMember{},
		&testTaskStages{},
		&testTask{},
		&testTaskComment{},
		&testTaskWorkTime{},
		&testProjectEvent{},
	)
}
