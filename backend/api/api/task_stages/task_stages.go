package task_stages

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	common "test.com/project-common"
)

type HandlerTaskStages struct {
}

func New() *HandlerTaskStages {
	return &HandlerTaskStages{}
}

type taskStage struct {
	Id          int64 `gorm:"primaryKey;autoIncrement" json:"id"`
	ProjectCode int64 `json:"project_code"`
	Name        string `json:"name"`
	Sort        int    `json:"sort"`
	CreateTime  int64  `json:"create_time"`
	Deleted     int8   `json:"deleted"`
}

func (*taskStage) TableName() string { return "ms_task_stages" }

type taskRow struct {
	Id           int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode  int64
	Name         string
	Description  string
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

func (*taskRow) TableName() string { return "ms_task" }

type memberRow struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	Name   string
	Avatar string
	Status int
}

func (*memberRow) TableName() string { return "ms_member" }

func (h *HandlerTaskStages) list(c *gin.Context) {
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
	err = db.Model(&taskStage{}).Where("project_code=? and deleted=0", pid).Count(&total).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "查询失败"))
		return
	}
	var list []taskStage
	err = db.
		Where("project_code=? and deleted=0", pid).
		Order("sort asc, id asc").
		Limit(int(page.PageSize)).
		Offset(int((page.Page - 1) * page.PageSize)).
		Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "查询失败"))
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, s := range list {
		out = append(out, gin.H{
			"id":           s.Id,
			"code":         codecs.EncryptInt64(s.Id),
			"name":         s.Name,
			"project_code": projectCode,
			"sort":         s.Sort,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (h *HandlerTaskStages) getAll(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	db := gorms.GetDB()
	var list []taskStage
	err = db.Where("project_code=? and deleted=0", pid).Order("sort asc, id asc").Find(&list).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "查询失败"))
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, s := range list {
		out = append(out, gin.H{
			"id":           s.Id,
			"code":         codecs.EncryptInt64(s.Id),
			"name":         s.Name,
			"project_code": projectCode,
			"sort":         s.Sort,
		})
	}
	c.JSON(http.StatusOK, result.Success(out))
}

func (h *HandlerTaskStages) taskTree(c *gin.Context) {
	h.getAll(c)
}

func (h *HandlerTaskStages) save(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusOK, result.Fail(400, "name必填"))
		return
	}
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	db := gorms.GetDB()
	var maxSort int
	_ = db.Model(&taskStage{}).Where("project_code=? and deleted=0", pid).Select("coalesce(max(sort),0)").Scan(&maxSort).Error
	stage := &taskStage{
		ProjectCode: pid,
		Name:        name,
		Sort:        maxSort + 1,
		CreateTime:  time.Now().UnixMilli(),
		Deleted:     0,
	}
	if err := db.Create(stage).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"id":           stage.Id,
		"code":         codecs.EncryptInt64(stage.Id),
		"name":         stage.Name,
		"project_code": projectCode,
		"sort":         stage.Sort,
	}))
}

func (h *HandlerTaskStages) edit(c *gin.Context) {
	result := &common.Result{}
	stageCode := c.PostForm("stageCode")
	name := c.PostForm("name")
	if stageCode == "" || name == "" {
		c.JSON(http.StatusOK, result.Fail(400, "参数不完整"))
		return
	}
	id, err := codecs.DecryptInt64(stageCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "stageCode无效"))
		return
	}
	db := gorms.GetDB()
	if err := db.Model(&taskStage{}).Where("id=?", id).Updates(map[string]any{"name": name}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTaskStages) del(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("code")
	if code == "" {
		c.JSON(http.StatusOK, result.Fail(400, "code必填"))
		return
	}
	id, err := codecs.DecryptInt64(code)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	db := gorms.GetDB()
	if err := db.Model(&taskStage{}).Where("id=?", id).Updates(map[string]any{"deleted": 1}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "删除失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTaskStages) sort(c *gin.Context) {
	result := &common.Result{}
	preCode := c.PostForm("preCode")
	nextCode := c.PostForm("nextCode")
	projectCode := c.PostForm("projectCode")
	if preCode == "" || projectCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "参数不完整"))
		return
	}
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	preId, err := codecs.DecryptInt64(preCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "preCode无效"))
		return
	}
	var nextId int64
	if nextCode != "" {
		nextId, _ = codecs.DecryptInt64(nextCode)
	}
	db := gorms.GetDB()
	var stages []taskStage
	if err := db.Where("project_code=? and deleted=0", pid).Order("sort asc, id asc").Find(&stages).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "排序失败"))
		return
	}
	var moved *taskStage
	newStages := make([]taskStage, 0, len(stages))
	for _, s := range stages {
		if s.Id == preId {
			tmp := s
			moved = &tmp
			continue
		}
		newStages = append(newStages, s)
	}
	if moved == nil {
		c.JSON(http.StatusOK, result.Success([]int{}))
		return
	}
	inserted := false
	if nextId != 0 {
		for i, s := range newStages {
			if s.Id == nextId {
				left := append([]taskStage{}, newStages[:i]...)
				right := append([]taskStage{}, newStages[i:]...)
				newStages = append(left, *moved)
				newStages = append(newStages, right...)
				inserted = true
				break
			}
		}
	}
	if !inserted {
		newStages = append(newStages, *moved)
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		for i, s := range newStages {
			if err := tx.Model(&taskStage{}).Where("id=?", s.Id).Update("sort", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "排序失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (h *HandlerTaskStages) tasks(c *gin.Context) {
	result := &common.Result{}
	stageCode := c.PostForm("stageCode")
	if stageCode == "" {
		c.JSON(http.StatusOK, result.Success([]any{}))
		return
	}
	stageId, err := codecs.DecryptInt64(stageCode)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "stageCode无效"))
		return
	}
	db := gorms.GetDB()
	var rows []taskRow
	if err := db.Where("stage_code=? and deleted=0", stageId).Order("sort asc, id asc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "查询失败"))
		return
	}
	memberIds := make([]int64, 0, len(rows))
	seen := map[int64]struct{}{}
	for _, t := range rows {
		execId := t.AssignTo
		if execId == 0 {
			execId = t.OwnerCode
		}
		if execId != 0 {
			if _, ok := seen[execId]; !ok {
				seen[execId] = struct{}{}
				memberIds = append(memberIds, execId)
			}
		}
	}
	executorMap := map[int64]memberRow{}
	if len(memberIds) > 0 {
		var members []memberRow
		_ = db.Where("id in ?", memberIds).Find(&members).Error
		for _, m := range members {
			executorMap[m.Id] = m
		}
	}
	out := make([]gin.H, 0, len(rows))
	for _, t := range rows {
		execId := t.AssignTo
		if execId == 0 {
			execId = t.OwnerCode
		}
		var executor any
		if m, ok := executorMap[execId]; ok {
			executor = gin.H{
				"code":   codecs.EncryptInt64(m.Id),
				"name":   m.Name,
				"avatar": m.Avatar,
			}
		}
		out = append(out, gin.H{
			"id":           t.Id,
			"id_num":       t.Id,
			"code":         codecs.EncryptInt64(t.Id),
			"name":         t.Name,
			"description":  t.Description,
			"done":         t.Done == 1,
			"canRead":      true,
			"status":       t.Status,
			"statusText":   "",
			"pri":          t.Priority,
			"begin_time":   t.BeginTime,
			"end_time":     t.EndTime,
			"executor":     executor,
			"childCount":   []int{0, 0},
			"tags":         []any{},
			"hasUnDone":    false,
			"hasSource":    false,
			"hasComment":   false,
			"like":         t.LikeCount,
			"star":         t.Star,
			"stage_code":   stageCode,
			"project_code": codecs.EncryptInt64(t.ProjectCode),
		})
	}
	c.JSON(http.StatusOK, result.Success(out))
}

