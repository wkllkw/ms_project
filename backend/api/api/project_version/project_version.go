package project_version

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/codecs"
	common "test.com/project-common"
)

type HandlerProjectVersion struct {
}

func New() *HandlerProjectVersion {
	return &HandlerProjectVersion{}
}

type projectVersionRow struct {
	Id              int64  `gorm:"primaryKey;autoIncrement"`
	FeaturesCode    int64  `gorm:"column:features_code"`
	ProjectCode     int64  `gorm:"column:project_code"`
	Name            string `gorm:"column:name"`
	Description     string `gorm:"column:description"`
	StartTime       int64  `gorm:"column:start_time"`
	PlanPublishTime int64  `gorm:"column:plan_publish_time"`
	PublishTime     int64  `gorm:"column:publish_time"`
	Status          int8   `gorm:"column:status"`
	Sort            int    `gorm:"column:sort"`
	CreateTime      int64  `gorm:"column:create_time"`
	UpdateTime      int64  `gorm:"column:update_time"`
}

func (*projectVersionRow) TableName() string { return "ms_project_version" }

type projectFeaturesRow struct {
	Id          int64  `gorm:"primaryKey"`
	ProjectCode int64  `gorm:"column:project_code"`
	Name        string `gorm:"column:name"`
}

func (*projectFeaturesRow) TableName() string { return "ms_project_features" }

// versionTaskRow 版本关联的任务
type versionTaskRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64  `gorm:"column:project_code"`
	Name        string `gorm:"column:name"`
	AssignTo    int64  `gorm:"column:assign_to"`
	StageCode   int64  `gorm:"column:stage_code"`
	VersionCode int64  `gorm:"column:version_code"`
	Done        int8   `gorm:"column:done"`
	Deleted     int8   `gorm:"column:deleted"`
}

func (*versionTaskRow) TableName() string { return "ms_task" }

// versionLogRow 版本日志
type versionLogRow struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	MemberCode  int64  `gorm:"column:member_code"`
	SourceCode  int64  `gorm:"column:source_code"`
	Content     string `gorm:"column:content"`
	Remark      string `gorm:"column:remark"`
	Type        string `gorm:"column:type"`
	CreateTime  int64  `gorm:"column:create_time"`
	Icon        string `gorm:"column:icon"`
}

func (*versionLogRow) TableName() string { return "ms_project_version_log" }

// versionMemberRow 版本日志关联的成员
type versionMemberRow struct {
	Id     int64  `gorm:"primaryKey;autoIncrement"`
	Name   string `gorm:"column:name"`
	Avatar string `gorm:"column:avatar"`
}

func (*versionMemberRow) TableName() string { return "ms_member" }

// list 获取版本列表
func (h *HandlerProjectVersion) list(c *gin.Context) {
	result := &common.Result{}
	featuresCode := c.PostForm("projectFeaturesCode")

	if featuresCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "projectFeaturesCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(featuresCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectFeaturesCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var rows []projectVersionRow
	db.Where("features_code=?", fid).Order("sort asc, id desc").Find(&rows)

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		// 计算状态文本
		statusText := "未开始"
		switch r.Status {
		case 1:
			statusText = "进行中"
		case 2:
			statusText = "延期发布"
		case 3:
			statusText = "已发布"
		}

		// 计算完成进度
		var totalTasks int64
		var doneTasks int64
		db.Model(&versionTaskRow{}).Where("version_code=? AND deleted=0", r.Id).Count(&totalTasks)
		db.Model(&versionTaskRow{}).Where("version_code=? AND deleted=0 AND done=1", r.Id).Count(&doneTasks)
		var schedule int
		if totalTasks > 0 {
			schedule = int(float64(doneTasks) / float64(totalTasks) * 100)
		}

		out = append(out, gin.H{
			"code":            codecs.EncryptInt64(r.Id),
			"featuresCode":    featuresCode,
			"name":            r.Name,
			"description":     r.Description,
			"startTime":       r.StartTime,
			"planPublishTime": r.PlanPublishTime,
			"publishTime":     r.PublishTime,
			"status":          r.Status,
			"statusText":      statusText,
			"schedule":        schedule,
			"sort":            r.Sort,
			"createTime":      r.CreateTime,
		})
	}

	c.JSON(http.StatusOK, result.Success(out))
}

// save 创建版本
func (h *HandlerProjectVersion) save(c *gin.Context) {
	result := &common.Result{}
	featuresCode := c.PostForm("featuresCode")
	name := c.PostForm("name")
	description := c.PostForm("description")
	startTimeStr := c.PostForm("startTime")
	planPublishTimeStr := c.PostForm("planPublishTime")

	if featuresCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "featuresCode不能为空"))
		return
	}

	fid, err := codecs.DecryptInt64(featuresCode)
	if err != nil || fid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "featuresCode无效"))
		return
	}

	if name == "" {
		c.JSON(http.StatusOK, result.Fail(400, "名称不能为空"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())

	// 获取版本库的项目ID
	var features projectFeaturesRow
	if err := db.Where("id=?", fid).First(&features).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "版本库不存在"))
		return
	}

	// 解析时间
	var startTime, planPublishTime int64
	if startTimeStr != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", startTimeStr, time.Local)
		if err == nil {
			startTime = t.UnixMilli()
		}
	}
	if planPublishTimeStr != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", planPublishTimeStr, time.Local)
		if err == nil {
			planPublishTime = t.UnixMilli()
		}
	}

	// 获取当前最大排序
	var maxSort int
	db.Model(&projectVersionRow{}).Where("features_code=?", fid).Select("COALESCE(MAX(sort), 0)").Scan(&maxSort)

	now := time.Now().UnixMilli()
	row := &projectVersionRow{
		FeaturesCode:    fid,
		ProjectCode:     features.ProjectCode,
		Name:            name,
		Description:     description,
		StartTime:       startTime,
		PlanPublishTime: planPublishTime,
		Status:          0,
		Sort:            maxSort + 1,
		CreateTime:      now,
		UpdateTime:      now,
	}

	if err := db.Create(row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":            codecs.EncryptInt64(row.Id),
		"featuresCode":    featuresCode,
		"name":            row.Name,
		"description":     row.Description,
		"startTime":       row.StartTime,
		"planPublishTime": row.PlanPublishTime,
		"status":          row.Status,
		"createTime":      row.CreateTime,
	}))
}

// edit 编辑版本
func (h *HandlerProjectVersion) edit(c *gin.Context) {
	result := &common.Result{}
	versionCode := c.PostForm("versionCode")
	name := c.PostForm("name")
	// 兼容两种字段命名：startTime/start_time, planPublishTime/plan_publish_time
	startTimeStr := c.PostForm("startTime")
	if startTimeStr == "" {
		startTimeStr = c.PostForm("start_time")
	}
	planPublishTimeStr := c.PostForm("planPublishTime")
	if planPublishTimeStr == "" {
		planPublishTimeStr = c.PostForm("plan_publish_time")
	}

	if versionCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode不能为空"))
		return
	}

	vid, err := codecs.DecryptInt64(versionCode)
	if err != nil || vid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	updates := map[string]any{
		"update_time": time.Now().UnixMilli(),
	}
	if name != "" {
		updates["name"] = name
	}
	// description 可以设置为空
	if desc := c.PostForm("description"); desc != "" {
		updates["description"] = desc
	}
	if startTimeStr != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", startTimeStr, time.Local)
		if err == nil {
			updates["start_time"] = t.UnixMilli()
		}
	} else if c.PostForm("start_time") == "" && c.PostForm("startTime") == "" {
		// 如果明确传了空值，则清除
		if _, ok := c.GetPostForm("start_time"); ok {
			updates["start_time"] = int64(0)
		}
		if _, ok := c.GetPostForm("startTime"); ok {
			updates["start_time"] = int64(0)
		}
	}
	if planPublishTimeStr != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04", planPublishTimeStr, time.Local)
		if err == nil {
			updates["plan_publish_time"] = t.UnixMilli()
		}
	}

	if err := db.Model(&projectVersionRow{}).Where("id=?", vid).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "编辑失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": versionCode}))
}

// del 删除版本
func (h *HandlerProjectVersion) del(c *gin.Context) {
	result := &common.Result{}
	versionCode := c.PostForm("versionCode")

	if versionCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode不能为空"))
		return
	}

	vid, err := codecs.DecryptInt64(versionCode)
	if err != nil || vid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	db.Delete(&projectVersionRow{}, vid)

	c.JSON(http.StatusOK, result.Success(gin.H{"code": versionCode}))
}

// changeStatus 更改版本状态（发布）
func (h *HandlerProjectVersion) changeStatus(c *gin.Context) {
	result := &common.Result{}
	versionCode := c.PostForm("versionCode")
	statusStr := c.PostForm("status")
	publishTimeStr := c.PostForm("publishTime")

	if versionCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode不能为空"))
		return
	}

	vid, err := codecs.DecryptInt64(versionCode)
	if err != nil || vid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode无效"))
		return
	}

	status := int8(0)
	if statusStr == "1" {
		status = 1
	} else if statusStr == "2" {
		status = 2
	} else if statusStr == "3" {
		status = 3
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	updates := map[string]any{
		"status":      status,
		"update_time": time.Now().UnixMilli(),
	}
	if status == 3 {
		// 发布版本
		if publishTimeStr != "" {
			t, err := time.ParseInLocation("2006-01-02 15:04", publishTimeStr, time.Local)
			if err == nil {
				updates["publish_time"] = t.UnixMilli()
			}
		} else {
			updates["publish_time"] = time.Now().UnixMilli()
		}
	}

	if err := db.Model(&projectVersionRow{}).Where("id=?", vid).Updates(updates).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "操作失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{"code": versionCode, "status": status}))
}

// read 获取版本详情
func (h *HandlerProjectVersion) read(c *gin.Context) {
	result := &common.Result{}
	versionCode := c.PostForm("versionCode")

	if versionCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode不能为空"))
		return
	}

	vid, err := codecs.DecryptInt64(versionCode)
	if err != nil || vid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var row projectVersionRow
	if err := db.Where("id=?", vid).First(&row).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "版本不存在"))
		return
	}

	// 获取版本库名称
	featuresCode := codecs.EncryptInt64(row.FeaturesCode)
	var features projectFeaturesRow
	featureName := ""
	if err := db.Where("id=?", row.FeaturesCode).First(&features).Error; err == nil {
		featureName = features.Name
	}

	// 计算状态文本
	statusText := "未开始"
	switch row.Status {
	case 1:
		statusText = "进行中"
	case 2:
		statusText = "延期发布"
	case 3:
		statusText = "已发布"
	}

	// 计算完成进度
	var totalTasks int64
	var doneTasks int64
	db.Model(&versionTaskRow{}).Where("version_code=? AND deleted=0", vid).Count(&totalTasks)
	db.Model(&versionTaskRow{}).Where("version_code=? AND deleted=0 AND done=1", vid).Count(&doneTasks)
	var schedule int
	if totalTasks > 0 {
		schedule = int(float64(doneTasks) / float64(totalTasks) * 100)
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":              versionCode,
		"featuresCode":      featuresCode,
		"featureName":       featureName,
		"projectCode":       codecs.EncryptInt64(row.ProjectCode),
		"name":              row.Name,
		"description":       row.Description,
		"startTime":         row.StartTime,
		"planPublishTime":   row.PlanPublishTime,
		"publishTime":       row.PublishTime,
		"status":            row.Status,
		"statusText":        statusText,
		"schedule":          schedule,
		"sort":              row.Sort,
		"createTime":        row.CreateTime,
	}))
}

// getVersionTask 获取版本关联的任务列表
func (h *HandlerProjectVersion) getVersionTask(c *gin.Context) {
	result := &common.Result{}
	versionCode := c.PostForm("versionCode")

	if versionCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode不能为空"))
		return
	}

	vid, err := codecs.DecryptInt64(versionCode)
	if err != nil || vid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	var tasks []versionTaskRow
	db.Where("version_code=? AND deleted=0", vid).Order("id desc").Find(&tasks)

	out := make([]gin.H, 0, len(tasks))
	for _, t := range tasks {
		item := gin.H{
			"code":        codecs.EncryptInt64(t.Id),
			"name":        t.Name,
			"done":        t.Done,
			"versionCode": versionCode,
		}
		// 获取执行者信息
		if t.AssignTo > 0 {
			var member versionMemberRow
			if err := db.Where("id=?", t.AssignTo).First(&member).Error; err == nil {
				item["executor"] = gin.H{
					"code":   codecs.EncryptInt64(member.Id),
					"name":   member.Name,
					"avatar": member.Avatar,
				}
			}
		}
		out = append(out, item)
	}

	c.JSON(http.StatusOK, result.Success(out))
}

// getVersionLog 获取版本日志
func (h *HandlerProjectVersion) getVersionLog(c *gin.Context) {
	result := &common.Result{}
	versionCode := c.PostForm("versionCode")
	showAll := c.PostForm("all")

	if versionCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode不能为空"))
		return
	}

	vid, err := codecs.DecryptInt64(versionCode)
	if err != nil || vid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	query := db.Where("source_code=?", vid)

	var total int64
	query.Model(&versionLogRow{}).Count(&total)

	var logs []versionLogRow
	if showAll != "" {
		query.Order("id asc").Find(&logs)
	} else {
		query.Order("id desc").Limit(5).Find(&logs)
		// 反转为时间正序
		for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
			logs[i], logs[j] = logs[j], logs[i]
		}
	}

	out := make([]gin.H, 0, len(logs))
	for _, l := range logs {
		item := gin.H{
			"code":       codecs.EncryptInt64(l.Id),
			"sourceCode": versionCode,
			"content":    l.Content,
			"remark":     l.Remark,
			"type":       l.Type,
			"createTime": l.CreateTime,
			"icon":       l.Icon,
		}
		// 获取操作人信息
		if l.MemberCode > 0 {
			var member versionMemberRow
			if err := db.Where("id=?", l.MemberCode).First(&member).Error; err == nil {
				item["member"] = gin.H{
					"code":   codecs.EncryptInt64(member.Id),
					"name":   member.Name,
					"avatar": member.Avatar,
				}
			} else {
				item["member"] = gin.H{"name": "", "avatar": ""}
			}
		} else {
			item["member"] = gin.H{"name": "", "avatar": ""}
		}
		out = append(out, item)
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  out,
		"total": total,
	}))
}

// addVersionTask 添加发布内容到版本
func (h *HandlerProjectVersion) addVersionTask(c *gin.Context) {
	result := &common.Result{}
	versionCode := c.PostForm("versionCode")
	taskCodeListStr := c.PostForm("taskCodeList")

	if versionCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode不能为空"))
		return
	}

	vid, err := codecs.DecryptInt64(versionCode)
	if err != nil || vid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "versionCode无效"))
		return
	}

	// 解析任务code列表 (JSON数组字符串)
	var taskCodes []string
	if taskCodeListStr != "" {
		// 简单处理：如果以 [ 开头则是 JSON 数组
		if len(taskCodeListStr) > 0 && taskCodeListStr[0] == '[' {
			// 去掉方括号和引号，按逗号分割
			inner := taskCodeListStr[1 : len(taskCodeListStr)-1]
			if inner != "" {
				parts := splitByComma(inner)
				for _, p := range parts {
					p = trimQuotes(p)
					if p != "" {
						taskCodes = append(taskCodes, p)
					}
				}
			}
		} else {
			taskCodes = append(taskCodes, taskCodeListStr)
		}
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	successTotal := 0

	for _, tc := range taskCodes {
		tid, err := codecs.DecryptInt64(tc)
		if err != nil || tid == 0 {
			continue
		}
		// 更新任务的 version_code
		if err := db.Model(&versionTaskRow{}).Where("id=? AND deleted=0", tid).
			Updates(map[string]any{"version_code": vid}).Error; err == nil {
			successTotal++
		}
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"successTotal": successTotal,
	}))
}

// removeVersionTask 移除发布内容
func (h *HandlerProjectVersion) removeVersionTask(c *gin.Context) {
	result := &common.Result{}
	taskCode := c.PostForm("taskCode")

	if taskCode == "" {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode不能为空"))
		return
	}

	tid, err := codecs.DecryptInt64(taskCode)
	if err != nil || tid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "taskCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	if err := db.Model(&versionTaskRow{}).Where("id=?", tid).
		Updates(map[string]any{"version_code": 0}).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "移除失败"))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{}))
}

// splitByComma 按逗号分割字符串
func splitByComma(s string) []string {
	var result []string
	current := ""
	inQuote := false
	for _, ch := range s {
		if ch == '"' {
			inQuote = !inQuote
			continue
		}
		if ch == ',' && !inQuote {
			result = append(result, current)
			current = ""
			continue
		}
		current += string(ch)
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// trimQuotes 去除字符串两端的引号和空格
func trimQuotes(s string) string {
	for len(s) > 0 && (s[0] == '"' || s[0] == ' ' || s[0] == '\'') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == '"' || s[len(s)-1] == ' ' || s[len(s)-1] == '\'') {
		s = s[:len(s)-1]
	}
	return s
}
