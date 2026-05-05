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
	Id           int64  `gorm:"primaryKey;autoIncrement"`
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
	DoneTime     int64  `gorm:"column:done_time;default:0"`
	LikeCount    int    `gorm:"column:like"`
	Star         int
	VersionCode  int64  `gorm:"column:version_code;default:0"`
	FeaturesCode int64  `gorm:"column:features_code;default:0"`
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
	Id         int64 `gorm:"primaryKey;autoIncrement"`
	TaskId     int64 `gorm:"column:task_id;index"`
	MemberId   int64 `gorm:"column:member_id;index"`
	IsExecutor int8  `gorm:"column:is_executor;default:0"`
	IsOwner    int8  `gorm:"column:is_owner;default:0"`
	JoinTime   int64 `gorm:"column:join_time"`
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

type File struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64  `gorm:"column:project_code;index"`
	MemberCode  int64  `gorm:"column:member_code;index"`
	Title       string `gorm:"column:title"`
	FileName    string `gorm:"column:file_name"`
	FileType    string `gorm:"column:file_type"`
	FileSize    int64  `gorm:"column:file_size"`
	FileUrl     string `gorm:"column:file_url"`
	FilePath    string `gorm:"column:file_path"`
	Description string `gorm:"column:description;type:text"`
	Deleted     int8   `gorm:"column:deleted;index"`
	CreateTime  int64  `gorm:"column:create_time"`
	UpdateTime  int64  `gorm:"column:update_time"`
}

func (*File) TableName() string { return "ms_file" }

type SourceLink struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	TaskCode    int64  `gorm:"column:task_code;index"`
	MemberCode  int64  `gorm:"column:member_code"`
	Title       string `gorm:"column:title"`
	Url         string `gorm:"column:url"`
	Description string `gorm:"column:description"`
	Sort        int    `gorm:"column:sort"`
	CreateTime  int64  `gorm:"column:create_time"`
}

func (*SourceLink) TableName() string { return "ms_source_link" }

// ========== 缺失的表模型 ==========

type Department struct {
	Id              int64  `gorm:"primaryKey;autoIncrement"`
	Name            string
	ParentId        int64  `gorm:"column:parent_id"`
	OrganizationCode string `gorm:"column:organization_code;index"`
	Sort            int
	CreateTime      int64  `gorm:"column:create_time"`
	Deleted         int8   `gorm:"index"`
}

func (*Department) TableName() string { return "ms_department" }

type ProjectMember struct {
	Id         int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"column:project_code;index"`
	MemberCode  int64 `gorm:"column:member_code;index"`
	JoinTime   int64  `gorm:"column:join_time"`
	IsOwner    int8   `gorm:"column:is_owner"`
	Authorize  string `gorm:"column:authorize;type:text"`
}

func (*ProjectMember) TableName() string { return "ms_project_member" }

type ProjectCollection struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"column:project_code;index"`
	MemberCode  int64 `gorm:"column:member_code;index"`
	CreateTime  int64 `gorm:"column:create_time"`
}

func (*ProjectCollection) TableName() string { return "ms_project_collection" }

type ProjectEvent struct {
	Id           int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode  int64  `gorm:"column:project_code;index"`
	MemberCode   int64  `gorm:"column:member_id;index"`
	TaskId       int64  `gorm:"column:task_id;index;default:0"`
	EventType    int8   `gorm:"column:event_type"`
	EventContent string `gorm:"column:event_content;type:text"`
	CreateTime   int64  `gorm:"column:create_time"`
}

func (*ProjectEvent) TableName() string { return "ms_project_event" }

type ProjectFeatures struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"column:project_code;index"`
	Name        string
	Description string `gorm:"type:text"`
	Sort        int
	CreateTime  int64 `gorm:"column:create_time"`
	UpdateTime  int64 `gorm:"column:update_time"`
}

func (*ProjectFeatures) TableName() string { return "ms_project_features" }

type ProjectTemplate struct {
	Id               int64  `gorm:"primaryKey;autoIncrement"`
	Code             string `gorm:"size:64;uniqueIndex"`
	Name             string
	Description      string `gorm:"type:text"`
	Sort             int
	CreateTime       int64  `gorm:"column:create_time"`
	OrganizationCode int64  `gorm:"column:organization_code;index"`
	Cover            string
	MemberCode       int64  `gorm:"column:member_code"`
	IsSystem         int8   `gorm:"column:is_system;default:0"`
}

func (*ProjectTemplate) TableName() string { return "ms_project_template" }

type ProjectVersion struct {
	Id               int64 `gorm:"primaryKey;autoIncrement"`
	FeaturesCode     int64 `gorm:"column:features_code"`
	ProjectCode      int64 `gorm:"column:project_code;index"`
	Name             string
	Description      string `gorm:"type:text"`
	StartTime        int64 `gorm:"column:start_time"`
	PlanPublishTime  int64 `gorm:"column:plan_publish_time"`
	PublishTime      int64 `gorm:"column:publish_time"`
	Status           int8
	Sort             int
	CreateTime       int64 `gorm:"column:create_time"`
	UpdateTime       int64 `gorm:"column:update_time"`
}

func (*ProjectVersion) TableName() string { return "ms_project_version" }

type ProjectVersionLog struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	MemberCode  int64  `gorm:"column:member_code;index"`
	SourceCode  int64  `gorm:"column:source_code"`
	Content     string `gorm:"type:text"`
	Remark      string
	Type        int8
	CreateTime  int64  `gorm:"column:create_time"`
	Icon        string
	FeaturesCode int64 `gorm:"column:features_code"`
}

func (*ProjectVersionLog) TableName() string { return "ms_project_version_log" }

type TaskStagesTemplate struct {
	Id                 int64 `gorm:"primaryKey;autoIncrement"`
	Name               string
	ProjectTemplateCode int64 `gorm:"column:project_template_code"`
	CreateTime         int64 `gorm:"column:create_time"`
	Sort               int
}

func (*TaskStagesTemplate) TableName() string { return "ms_task_stages_template" }

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
		&File{},
		&SourceLink{},
		// 缺失的表
		&Department{},
		&ProjectMember{},
		&ProjectCollection{},
		&ProjectEvent{},
		&ProjectFeatures{},
		&ProjectTemplate{},
		&ProjectVersion{},
		&ProjectVersionLog{},
		&TaskStagesTemplate{},
	)
}
