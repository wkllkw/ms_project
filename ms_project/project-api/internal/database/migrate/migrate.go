package migrate

import (
	"test.com/project-api/internal/database/gorms"
)

type TaskStages struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64
	Name       string
	Sort       int
	CreateTime int64
	Deleted    int8
}

func (*TaskStages) TableName() string { return "ms_task_stages" }

type Task struct {
	Id           int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode  int64
	Name         string
	Description  string `gorm:"type:text"`
	Status       int8
	Priority     int8
	BeginTime    int64
	EndTime      int64
	CreateTime   int64
	MemberCode   int64
	OwnerCode    int64
	AssignTo     int64
	StageCode    int64
	ParentTaskId int64
	Sort         int
	Deleted      int8
	Private      int8
	Done         int8
	LikeCount    int `gorm:"column:like"`
	Star         int
}

func (*Task) TableName() string { return "ms_task" }

type TaskTag struct {
	Id              int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode     int64
	Name            string
	Color           string
	CreateTime      int64
	Deleted         int8
}

func (*TaskTag) TableName() string { return "ms_task_tag" }

type TaskTagRel struct {
	Id      int64 `gorm:"primaryKey;autoIncrement"`
	TaskId  int64 `gorm:"index"`
	TagId   int64 `gorm:"index"`
}

func (*TaskTagRel) TableName() string { return "ms_task_tag_rel" }

type TaskMember struct {
	Id      int64 `gorm:"primaryKey;autoIncrement"`
	TaskId  int64 `gorm:"index"`
	MemberId int64 `gorm:"index"`
}

func (*TaskMember) TableName() string { return "ms_task_member" }

type TaskComment struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	TaskId     int64 `gorm:"index"`
	MemberId   int64
	Comment    string `gorm:"type:text"`
	CreateTime int64
}

func (*TaskComment) TableName() string { return "ms_task_comment" }

type TaskWorkTime struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	TaskId     int64 `gorm:"index"`
	MemberId   int64
	WorkTime   int64
	Remark     string
	CreateTime int64
}

func (*TaskWorkTime) TableName() string { return "ms_task_work_time" }

type ProjectEvents struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"index"`
	Title       string
	Description string `gorm:"type:text"`
	BeginTime   string
	EndTime     string
	AllDay      int8
	Position    string
	CreateBy    int64
	CreateTime  int64
	Deleted     int8 `gorm:"index"`
}

func (*ProjectEvents) TableName() string { return "ms_project_events" }

type ProjectEventsMember struct {
	Id       int64 `gorm:"primaryKey;autoIncrement"`
	EventsId int64 `gorm:"index"`
	MemberId int64 `gorm:"index"`
	Status   int8
}

func (*ProjectEventsMember) TableName() string { return "ms_project_events_member" }

type DepartmentMember struct {
	Id           int64 `gorm:"primaryKey;autoIncrement"`
	DepartmentId int64 `gorm:"index"`
	MemberId     int64 `gorm:"index"`
}

func (*DepartmentMember) TableName() string { return "ms_department_member" }

type Notify struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	MemberCode int64 `gorm:"index"`
	Title      string
	Content    string `gorm:"type:text"`
	Type       int8   `gorm:"index"`
	IsRead     int8   `gorm:"index"`
	CreateTime int64  `gorm:"index"`
	Action     string
	SendData   string `gorm:"type:text;column:send_data"`
}

func (*Notify) TableName() string { return "ms_notify" }

type ProjectMenu struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	Pid        int64 `gorm:"index"`
	Title      string
	Icon       string
	Url        string
	FilePath   string
	Params     string
	Node       string
	Sort       int
	Status     int
	CreateBy   int64
	IsInner    int
	Values     string
	ShowSlider int
}

func (*ProjectMenu) TableName() string { return "ms_project_menu" }

type ProjectAuth struct {
	Id        int64 `gorm:"primaryKey;autoIncrement"`
	Title     string
	Desc      string `gorm:"column:desc"`
	Status    int
	IsDefault int   `gorm:"column:is_default"`
	CreateAt  int64 `gorm:"index"`
}

func (*ProjectAuth) TableName() string { return "ms_project_auth" }

type ProjectAuthNode struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	AuthId int64 `gorm:"index"`
	Node   string
}

func (*ProjectAuthNode) TableName() string { return "ms_project_auth_node" }

type InviteLink struct {
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	ProjectId  int64 `gorm:"index"`
	InviteCode string `gorm:"uniqueIndex;size:64"`
	ExpiredAt  int64
	CreateBy   int64
	CreateTime int64
}

func (*InviteLink) TableName() string { return "ms_invite_link" }

func AutoMigrate() error {
	db := gorms.GetDB()
	return db.AutoMigrate(
		&TaskStages{},
		&Task{},
		&TaskTag{},
		&TaskTagRel{},
		&TaskMember{},
		&TaskComment{},
		&TaskWorkTime{},
		&ProjectEvents{},
		&ProjectEventsMember{},
		&DepartmentMember{},
		&Notify{},
		&ProjectMenu{},
		&ProjectAuth{},
		&ProjectAuthNode{},
		&InviteLink{},
	)
}
