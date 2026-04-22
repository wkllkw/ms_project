package project

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"test.com/project-api/internal/authz"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/internal/menus"
	"test.com/project-api/pkg/codecs"
	"test.com/project-api/pkg/model"
	"test.com/project-api/pkg/model/menu"
	"test.com/project-api/pkg/model/pro"
	common "test.com/project-common"
	"test.com/project-grpc/project"
)

type HandlerProject struct {
}

type projectRow struct {
	Id                 int64 `gorm:"primaryKey;autoIncrement"`
	Cover              string
	Name               string
	Description        string
	Private            int `gorm:"column:private"`
	Deleted            int
	Archive            int
	Schedule           float64
	CreateTime         int64 `gorm:"column:create_time"`
	OrganizationCode   int64 `gorm:"column:organization_code"`
	DeletedTime        string
	ArchiveTime        int64 `gorm:"column:archive_time"`
	Prefix             string
	OpenPrefix         int    `gorm:"column:open_prefix"`
	OpenBeginTime      int    `gorm:"column:open_begin_time"`
	OpenTaskPrivate    int    `gorm:"column:open_task_private"`
	TaskBoardTheme     string `gorm:"column:task_board_theme"`
	AutoUpdateSchedule int    `gorm:"column:auto_update_schedule"`
}

func (*projectRow) TableName() string { return "ms_project" }

type projectMemberRow struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"column:project_code;index"`
	MemberCode  int64 `gorm:"column:member_code;index"`
	JoinTime    int64 `gorm:"column:join_time"`
	IsOwner     int   `gorm:"column:is_owner"`
}

func (*projectMemberRow) TableName() string { return "ms_project_member" }

type projectCollectionRow struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode int64 `gorm:"column:project_code;index"`
	MemberCode  int64 `gorm:"column:member_code;index"`
	CreateTime  int64 `gorm:"column:create_time"`
}

func (*projectCollectionRow) TableName() string { return "ms_project_collection" }

type memberBaseRow struct {
	Id     int64 `gorm:"primaryKey;autoIncrement"`
	Name   string
	Avatar string
}

func (*memberBaseRow) TableName() string { return "ms_member" }

func orgCodeToInt64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case string:
		i, _ := strconv.ParseInt(t, 10, 64)
		return i
	default:
		return 0
	}
}

func nowDeletedTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (p *HandlerProject) listFromDB(c *gin.Context, selectBy string) (gin.H, error) {
	page := &model.Page{}
	page.Bind(c)
	memberId := c.GetInt64("memberId")
	orgVal, _ := c.Get("organizationCode")
	orgCode := orgCodeToInt64(orgVal)

	db := gorms.GetDB().WithContext(c.Request.Context())
	base := db.Table("ms_project p").
		Joins("join ms_project_member pm on pm.project_code=p.id").
		Joins("left join ms_project_collection pc on pc.project_code=p.id and pc.member_code=?", memberId).
		Joins("left join ms_project_member opm on opm.project_code=p.id and opm.is_owner=1").
		Joins("left join ms_member owner on owner.id=opm.member_code").
		Where("pm.member_code=?", memberId)
	if orgCode != 0 {
		base = base.Where("p.organization_code=?", orgCode)
	}
	switch selectBy {
	case "collect":
		base = base.Where("pc.id is not null").Where("p.deleted=0").Where("p.archive=0")
	case "archive":
		base = base.Where("p.deleted=0").Where("p.archive=1")
	case "deleted":
		base = base.Where("p.deleted=1")
	default:
		base = base.Where("p.deleted=0").Where("p.archive=0")
	}

	var total int64
	_ = base.Distinct("p.id").Count(&total).Error

	rows := make([]struct {
		Id          int64
		Cover       string
		Name        string
		Description string
		Private     int `gorm:"column:private"`
		Schedule    float64
		CreateTime  int64 `gorm:"column:create_time"`
		DeletedTime string
		OwnerName   string `gorm:"column:owner_name"`
		Collected   int    `gorm:"column:collected"`
	}, 0)
	err := base.Select("p.id, p.cover, p.name, p.description, p.private, p.schedule, p.create_time, p.deleted_time, coalesce(owner.name,'') as owner_name, case when pc.id is null then 0 else 1 end as collected").
		Order("p.id desc").
		Limit(int(page.PageSize)).
		Offset(int((page.Page - 1) * page.PageSize)).
		Scan(&rows).Error
	if err != nil {
		return gin.H{"list": []any{}, "total": 0}, err
	}
	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		list = append(list, gin.H{
			"code":         codecs.EncryptInt64(r.Id),
			"cover":        r.Cover,
			"name":         r.Name,
			"description":  r.Description,
			"private":      r.Private,
			"schedule":     r.Schedule,
			"create_time":  r.CreateTime,
			"deleted_time": r.DeletedTime,
			"owner_name":   r.OwnerName,
			"collected":    r.Collected == 1,
		})
	}
	return gin.H{"list": list, "total": total}, nil
}

// @Summary 项目列表
// @Description 查询当前用户可访问的项目列表（支持分页和关键词搜索）
// @Tags project
// @Accept x-www-form-urlencoded
// @Produce json
// @Param keyword formData string false "项目名称搜索"
// @Param page formData int false "页码" default(1)
// @Param pageSize formData int false "每页条数" default(10)
// @Success 200 {object} common.Result "返回项目列表和分页信息"
// @Security ApiKeyAuth
// @Router /project/index [post]
func (p *HandlerProject) index(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 侧边栏菜单以当前 API 的菜单配置为准，避免被旧的 gRPC 菜单数据覆盖
	if err := menus.SeedDefaultIfEmpty(ctx); err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "初始化菜单失败"))
		return
	}
	pms, dbErr := menus.FindMenus(ctx)
	if dbErr != nil {
		c.JSON(http.StatusOK, result.Success([]*menu.Menu{}))
		return
	}
	tree := menus.BuildTree(pms)
	c.JSON(http.StatusOK, result.Success(tree))
}

func (p *HandlerProject) myProjectList(c *gin.Context) {
	result := &common.Result{}
	//1. 获取参数
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if memberCode := c.PostForm("memberCode"); memberCode != "" {
		memberId, err := codecs.DecryptInt64(memberCode)
		if err != nil || memberId == 0 {
			c.JSON(http.StatusOK, result.Fail(400, "memberCode无效"))
			return
		}
		page := &model.Page{}
		page.Bind(c)
		db := gorms.GetDB().WithContext(ctx)
		base := db.Table("ms_project_member pm").Joins("join ms_project p on p.id=pm.project_code").Where("pm.member_code=?", memberId).Where("p.deleted=0")
		var total int64
		_ = base.Count(&total).Error
		var rows []struct {
			Id          int64
			Cover       string
			Name        string
			Description string
			Private     int `gorm:"column:private"`
		}
		_ = base.Select("p.id, p.cover, p.name, p.description, p.private").Order("pm.id desc").Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).Scan(&rows).Error
		list := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			list = append(list, gin.H{
				"code":        codecs.EncryptInt64(r.Id),
				"cover":       r.Cover,
				"name":        r.Name,
				"description": r.Description,
				"private":     r.Private,
			})
		}
		c.JSON(http.StatusOK, result.Success(gin.H{"list": list, "total": total}))
		return
	}
	memberId := c.GetInt64("memberId")
	memberName := c.GetString("memberName")
	page := &model.Page{}
	page.Bind(c)
	selectBy := c.PostForm("selectBy")
	msg := &project.ProjectRpcMessage{
		MemberId:   memberId,
		MemberName: memberName,
		SelectBy:   selectBy,
		Page:       page.Page,
		PageSize:   page.PageSize}
	myProjectResponse, err := ProjectServiceClient.FindProjectByMemId(ctx, msg)
	if err == nil && myProjectResponse != nil {
		var pms []*pro.ProjectAndMember
		copier.Copy(&pms, myProjectResponse.Pm)
		if pms == nil {
			pms = []*pro.ProjectAndMember{}
		}
		c.JSON(http.StatusOK, result.Success(gin.H{
			"list":  pms,
			"total": myProjectResponse.Total,
		}))
		return
	}

	data, _ := p.listFromDB(c, selectBy)
	c.JSON(http.StatusOK, result.Success(data))
}

func (p *HandlerProject) projectTemplate(c *gin.Context) {
	result := &common.Result{}
	//1. 获取参数
	page := &model.Page{}
	page.Bind(c)
	viewTypeStr := c.PostForm("viewType")
	viewType, _ := strconv.ParseInt(viewTypeStr, 10, 64)
	memberId := c.GetInt64("memberId")
	orgVal, _ := c.Get("organizationCode")
	orgCode := orgCodeToInt64(orgVal)
	// 直接数据库查询
	db := gorms.GetDB().WithContext(c.Request.Context())
	var templates []*pro.ProjectTemplate
	query := db.Model(&pro.ProjectTemplate{})
	// 根据viewType筛选
	if viewType == 0 {
		// 自定义模板：非系统模板，且属于当前组织或当前成员
		query = query.Where("is_system = 0 AND (organization_code = ? OR member_code = ?)", orgCode, memberId)
	} else if viewType == 1 {
		// 公共模板：系统模板
		query = query.Where("is_system = 1")
	} else if viewType == 2 {
		// 收藏的模板，需要关联收藏表，暂不实现，返回空
		query = query.Where("1 = 0")
	}
	// 其他viewType（如-1）表示全部模板，不添加筛选
	var total int64
	query.Count(&total)
	err := query.Offset(int((page.Page - 1) * page.PageSize)).Limit(int(page.PageSize)).Find(&templates).Error
	if err != nil {
		c.JSON(http.StatusOK, result.Success(gin.H{"list": []*pro.ProjectTemplate{}, "total": 0}))
		return
	}
	if templates == nil {
		templates = []*pro.ProjectTemplate{}
	}
	// 为每个模板填充加密的 Code 字段，并查询任务阶段
	for _, tmpl := range templates {
		tmpl.Code = codecs.EncryptInt64(int64(tmpl.Id))
		var stages []*pro.TaskStagesOnlyName
		_ = db.Table("ms_task_stages_template").Where("project_template_code = ?", tmpl.Id).Order("sort desc, id asc").Find(&stages).Error
		if stages == nil {
			stages = []*pro.TaskStagesOnlyName{}
		}
		tmpl.TaskStages = stages
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  templates,
		"total": total,
	}))
}

// @Summary 创建项目
// @Description 创建新的协作项目
// @Tags project
// @Accept x-www-form-urlencoded
// @Produce json
// @Param name formData string true "项目名称"
// @Param description formData string false "项目描述"
// @Param cover formData string false "项目封面URL"
// @Param prefix formData string false "项目代号前缀"
// @Success 200 {object} common.Result "返回新项目加密ID"
// @Failure 400 {object} common.Result "参数错误"
// @Security ApiKeyAuth
// @Router /project/save [post]
func (p *HandlerProject) projectSave(c *gin.Context) {
	result := &common.Result{}
	//1. 获取参数
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	memberId := c.GetInt64("memberId")
	organizationCode := c.GetString("organizationCode")
	req := &pro.SaveProjectRequest{}
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "参数错误"))
		return
	}
	msg := &project.ProjectRpcMessage{
		MemberId:         memberId,
		OrganizationCode: organizationCode,
		TemplateCode:     req.TemplateCode,
		Name:             req.Name,
		Id:               int64(req.Id),
		Description:      req.Description,
	}
	saveProject, err := ProjectServiceClient.SaveProject(ctx, msg)
	var rsp *pro.SaveProject
	if err == nil && saveProject != nil {
		rsp = &pro.SaveProject{}
		copier.Copy(rsp, saveProject)
	} else {
		org := orgCodeToInt64(organizationCode)
		db := gorms.GetDB().WithContext(c.Request.Context())
		now := time.Now().UnixMilli()
		if req.Id == 0 {
			row := &projectRow{
				Name:             req.Name,
				Description:      req.Description,
				Cover:            "",
				Private:          0,
				Deleted:          0,
				Archive:          0,
				Schedule:         0,
				CreateTime:       now,
				OrganizationCode: org,
			}
			if dbErr := db.Create(row).Error; dbErr != nil {
				c.JSON(http.StatusOK, result.Fail(500, "创建项目失败"))
				return
			}
			_ = db.Create(&projectMemberRow{
				ProjectCode: row.Id,
				MemberCode:  memberId,
				JoinTime:    now,
				IsOwner:     1,
			}).Error
			rsp = &pro.SaveProject{
				Id:               row.Id,
				Cover:            row.Cover,
				Name:             row.Name,
				Description:      row.Description,
				Code:             codecs.EncryptInt64(row.Id),
				CreateTime:       strconv.FormatInt(row.CreateTime, 10),
				OrganizationCode: organizationCode,
			}
		} else {
			id := int64(req.Id)
			updates := map[string]any{
				"name":        req.Name,
				"description": req.Description,
			}
			_ = db.Model(&projectRow{}).Where("id=?", id).Updates(updates).Error
			rsp = &pro.SaveProject{
				Id:               id,
				Name:             req.Name,
				Description:      req.Description,
				Code:             codecs.EncryptInt64(id),
				CreateTime:       strconv.FormatInt(now, 10),
				OrganizationCode: organizationCode,
			}
		}
	}
	if req.Id == 0 {
		projectId, err := codecs.DecryptInt64(rsp.Code)
		if err == nil {
			db := gorms.GetDB()
			var stageCount int64
			_ = db.Table("ms_task_stages").Where("project_code=? and deleted=0", projectId).Count(&stageCount).Error
			if stageCount == 0 {
				templates := make([]struct {
					Id                  int64 `gorm:"primaryKey;autoIncrement"`
					Name                string
					ProjectTemplateCode int64
					Sort                int
				}, 0)
				templateId, err := codecs.DecryptInt64(req.TemplateCode)
				if err == nil && templateId != 0 {
					_ = db.Table("ms_task_stages_template").Where("project_template_code=?", templateId).Order("sort desc, id asc").Find(&templates).Error
				}
				stageNames := make([]string, 0, len(templates))
				for _, t := range templates {
					if t.Name != "" {
						stageNames = append(stageNames, t.Name)
					}
				}
				if len(stageNames) == 0 {
					stageNames = []string{"待处理", "进行中", "已完成"}
				}
				now := time.Now().UnixMilli()
				for i, name := range stageNames {
					_ = db.Table("ms_task_stages").Create(map[string]any{
						"project_code": projectId,
						"name":         name,
						"sort":         i + 1,
						"create_time":  now,
						"deleted":      0,
					}).Error
				}
			}
		}
	}
	c.JSON(http.StatusOK, result.Success(rsp))
}

// @Summary 项目详情
// @Description 获取项目完整信息（含成员列表、统计等）
// @Tags project
// @Accept x-www-form-urlencoded
// @Produce json
// @Param code formData string true "项目加密ID"
// @Success 200 {object} common.Result "返回项目完整数据"
// @Failure 404 {object} common.Result "项目不存在"
// @Security ApiKeyAuth
// @Router /project/read [post]
func (p *HandlerProject) readProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	detail, err := ProjectServiceClient.FindProjectDetail(ctx, &project.ProjectRpcMessage{ProjectCode: projectCode, MemberId: memberId})
	if err == nil && detail != nil {
		pd := &pro.ProjectDetail{}
		copier.Copy(pd, detail)
		c.JSON(http.StatusOK, result.Success(pd))
		return
	}
	projectId, decErr := codecs.DecryptInt64(projectCode)
	if decErr != nil || projectId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	var pr projectRow
	if dbErr := db.Where("id=?", projectId).First(&pr).Error; dbErr != nil {
		c.JSON(http.StatusOK, result.Fail(404, "项目不存在"))
		return
	}
	var owner memberBaseRow
	_ = db.Table("ms_project_member pm").Joins("join ms_member m on m.id=pm.member_code").Where("pm.project_code=? and pm.is_owner=1", projectId).Select("m.id,m.name,m.avatar").First(&owner).Error
	var collected int64
	_ = db.Model(&projectCollectionRow{}).Where("project_code=? and member_code=?", projectId, memberId).Count(&collected).Error
	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":                 codecs.EncryptInt64(pr.Id),
		"cover":                pr.Cover,
		"name":                 pr.Name,
		"description":          pr.Description,
		"private":              pr.Private,
		"deleted":              pr.Deleted,
		"archive":              pr.Archive,
		"schedule":             pr.Schedule,
		"create_time":          pr.CreateTime,
		"owner_name":           owner.Name,
		"owner_avatar":         owner.Avatar,
		"collected":            collected > 0,
		"prefix":               pr.Prefix,
		"open_prefix":          pr.OpenPrefix,
		"open_begin_time":      pr.OpenBeginTime,
		"open_task_private":    pr.OpenTaskPrivate,
		"task_board_theme":     pr.TaskBoardTheme,
		"auto_update_schedule": pr.AutoUpdateSchedule,
	}))
}

func (p *HandlerProject) recycleProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")

	// 权限校验：只有项目负责人才能移入回收站
	projectId, ok := authz.CanManageProject(c, projectCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此项目"))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateDeletedProject(ctx, &project.ProjectRpcMessage{ProjectCode: projectCode, Deleted: true})
	if err != nil {
		_ = gorms.GetDB().WithContext(c.Request.Context()).Model(&projectRow{}).Where("id=?", projectId).Updates(map[string]any{
			"deleted":      1,
			"deleted_time": nowDeletedTime(),
		}).Error
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) recoveryProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")

	// 权限校验：只有项目负责人才能恢复项目
	projectId, ok := authz.CanManageProject(c, projectCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此项目"))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateDeletedProject(ctx, &project.ProjectRpcMessage{ProjectCode: projectCode, Deleted: false})
	if err != nil {
		_ = gorms.GetDB().WithContext(c.Request.Context()).Model(&projectRow{}).Where("id=?", projectId).Updates(map[string]any{
			"deleted":      0,
			"deleted_time": "",
		}).Error
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) collectProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	collectType := c.PostForm("type")
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := ProjectServiceClient.UpdateCollectProject(ctx, &project.ProjectRpcMessage{ProjectCode: projectCode, CollectType: collectType, MemberId: memberId})
	if err != nil {
		projectId, decErr := codecs.DecryptInt64(projectCode)
		if decErr != nil || projectId == 0 {
			c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
			return
		}
		db := gorms.GetDB().WithContext(c.Request.Context())
		if collectType == "cancel" {
			_ = db.Where("project_code=? and member_code=?", projectId, memberId).Delete(&projectCollectionRow{}).Error
		} else {
			var cnt int64
			_ = db.Model(&projectCollectionRow{}).Where("project_code=? and member_code=?", projectId, memberId).Count(&cnt).Error
			if cnt == 0 {
				_ = db.Create(&projectCollectionRow{ProjectCode: projectId, MemberCode: memberId, CreateTime: time.Now().UnixMilli()}).Error
			}
		}
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// @Summary 编辑项目
// @Description 修改项目名称、描述、封面等信息
// @Tags project
// @Accept x-www-form-urlencoded
// @Produce json
// @Param code formData string true "项目加密ID"
// @Param name formData string false "项目名称"
// @Param description formData string false "项目描述"
// @Param cover formData string false "项目封面URL"
// @Success 200 {object} common.Result "返回编辑后的项目数据"
// @Security ApiKeyAuth
// @Router /project/edit [post]
func (p *HandlerProject) editProject(c *gin.Context) {
	result := &common.Result{}
	var req *pro.ProjectReq
	_ = c.ShouldBind(&req)
	memberId := c.GetInt64("memberId")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &project.UpdateProjectMessage{}
	copier.Copy(msg, req)
	msg.MemberId = memberId
	_, err := ProjectServiceClient.UpdateProject(ctx, msg)
	if err != nil {
		projectId, decErr := codecs.DecryptInt64(req.ProjectCode)
		if decErr != nil || projectId == 0 {
			c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
			return
		}
		updates := map[string]any{}
		if req.Name != "" {
			updates["name"] = req.Name
		}
		updates["description"] = req.Description
		if req.Cover != "" {
			updates["cover"] = req.Cover
		}
		updates["private"] = req.Private
		if req.Prefix != "" {
			updates["prefix"] = req.Prefix
		}
		updates["open_prefix"] = req.OpenPrefix
		updates["open_begin_time"] = req.OpenBeginTime
		updates["open_task_private"] = req.OpenTaskPrivate
		updates["task_board_theme"] = req.TaskBoardTheme
		updates["schedule"] = req.Schedule
		updates["auto_update_schedule"] = req.AutoUpdateSchedule
		_ = gorms.GetDB().WithContext(c.Request.Context()).Model(&projectRow{}).Where("id=?", projectId).Updates(updates).Error
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) projectTemplateSave(c *gin.Context) {
	result := &common.Result{}
	name := c.PostForm("name")
	description := c.PostForm("description")
	cover := c.PostForm("cover")
	code := c.PostForm("code")
	if name == "" {
		c.JSON(http.StatusOK, result.Fail(400, "模板名称不能为空"))
		return
	}
	memberId := c.GetInt64("memberId")
	orgVal, _ := c.Get("organizationCode")
	orgCode := orgCodeToInt64(orgVal)
	db := gorms.GetDB().WithContext(c.Request.Context())
	now := time.Now().UnixMilli()
	var templateId int64
	if code != "" {
		// 更新
		decryptedId, err := codecs.DecryptInt64(code)
		if err != nil || decryptedId == 0 {
			c.JSON(http.StatusOK, result.Fail(400, "code无效"))
			return
		}
		templateId = decryptedId
		updates := map[string]any{
			"name":        name,
			"description": description,
			"cover":       cover,
		}
		if dbErr := db.Table("ms_project_template").Where("id = ?", templateId).Updates(updates).Error; dbErr != nil {
			c.JSON(http.StatusOK, result.Fail(500, "更新失败"))
			return
		}
	} else {
		// 创建
		type templateRow struct {
			Id               int64  `gorm:"primaryKey;autoIncrement"`
			Name             string
			Description      string
			Cover            string
			CreateTime       int64
			OrganizationCode int64
			MemberCode       int64
			IsSystem         int
		}
		newRow := &templateRow{
			Name:             name,
			Description:      description,
			Cover:            cover,
			CreateTime:       now,
			OrganizationCode: orgCode,
			MemberCode:       memberId,
			IsSystem:         0,
		}
		if dbErr := db.Table("ms_project_template").Create(newRow).Error; dbErr != nil {
			c.JSON(http.StatusOK, result.Fail(500, "创建失败"))
			return
		}
		templateId = newRow.Id
	}
	// 返回模板信息
	var template pro.ProjectTemplate
	_ = db.Where("id = ?", templateId).First(&template).Error
	template.Code = codecs.EncryptInt64(int64(template.Id))
	var stages []*pro.TaskStagesOnlyName
	_ = db.Table("ms_task_stages_template").Where("project_template_code = ?", template.Id).Order("sort desc, id asc").Find(&stages).Error
	if stages == nil {
		stages = []*pro.TaskStagesOnlyName{}
	}
	template.TaskStages = stages
	c.JSON(http.StatusOK, result.Success(template))
}

func (p *HandlerProject) projectTemplateDelete(c *gin.Context) {
	result := &common.Result{}
	code := c.PostForm("code")
	templateId, err := codecs.DecryptInt64(code)
	if err != nil || templateId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "code无效"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	// 删除关联的任务阶段模板
	_ = db.Table("ms_task_stages_template").Where("project_template_code = ?", templateId).Delete(nil).Error
	// 删除项目模板
	if dbErr := db.Table("ms_project_template").Where("id = ?", templateId).Delete(nil).Error; dbErr != nil {
		c.JSON(http.StatusOK, result.Fail(500, "删除失败"))
		return
	}
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) quitProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	memberId := c.GetInt64("memberId")
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	_ = gorms.GetDB().WithContext(c.Request.Context()).Where("project_code=? and member_code=? and is_owner=0", projectId, memberId).Delete(&projectMemberRow{}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) archiveProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")

	// 权限校验：只有项目负责人才能归档
	projectId, ok := authz.CanManageProject(c, projectCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此项目"))
		return
	}
	_ = gorms.GetDB().WithContext(c.Request.Context()).Model(&projectRow{}).Where("id=?", projectId).Updates(map[string]any{
		"archive":      1,
		"archive_time": time.Now().UnixMilli(),
	}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) recoveryArchiveProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")

	// 权限校验：只有项目负责人才能恢复归档
	projectId, ok := authz.CanManageProject(c, projectCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此项目"))
		return
	}
	_ = gorms.GetDB().WithContext(c.Request.Context()).Model(&projectRow{}).Where("id=?", projectId).Updates(map[string]any{
		"archive":      0,
		"archive_time": int64(0),
	}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

func (p *HandlerProject) deleteProject(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")

	// 权限校验：只有项目负责人才能删除项目
	projectId, ok := authz.CanManageProject(c, projectCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此项目"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	_ = db.Where("project_code=?", projectId).Delete(&projectCollectionRow{}).Error
	_ = db.Where("project_code=?", projectId).Delete(&projectMemberRow{}).Error
	_ = db.Where("id=?", projectId).Delete(&projectRow{}).Error
	c.JSON(http.StatusOK, result.Success([]int{}))
}

// batchDeleteProjects 批量彻底删除项目
func (p *HandlerProject) batchDeleteProjects(c *gin.Context) {
	result := &common.Result{}
	projectCodes := c.PostFormArray("projectCodes")
	if len(projectCodes) == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "请选择要删除的项目"))
		return
	}

	memberId := c.GetInt64("memberId")
	db := gorms.GetDB().WithContext(c.Request.Context())
	successCount := 0
	failCount := 0

	for _, projectCode := range projectCodes {
		projectId, err := codecs.DecryptInt64(projectCode)
		if err != nil || projectId == 0 {
			failCount++
			continue
		}
		// 权限校验
		if !authz.IsProjectOwner(db, memberId, projectId) {
			failCount++
			continue
		}
		// 只能删除在回收站中的项目
		var pr projectRow
		if err := db.Where("id=?", projectId).First(&pr).Error; err != nil || pr.Deleted != 1 {
			failCount++
			continue
		}
		// 彻底删除
		_ = db.Where("project_code=?", projectId).Delete(&projectCollectionRow{}).Error
		_ = db.Where("project_code=?", projectId).Delete(&projectMemberRow{}).Error
		_ = db.Where("id=?", projectId).Delete(&projectRow{}).Error
		successCount++
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"successCount": successCount,
		"failCount":    failCount,
		"message":      fmt.Sprintf("成功删除 %d 个项目，失败 %d 个", successCount, failCount),
	}))
}

// batchRecoveryProjects 批量恢复项目
func (p *HandlerProject) batchRecoveryProjects(c *gin.Context) {
	result := &common.Result{}
	projectCodes := c.PostFormArray("projectCodes")
	if len(projectCodes) == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "请选择要恢复的项目"))
		return
	}

	memberId := c.GetInt64("memberId")
	db := gorms.GetDB().WithContext(c.Request.Context())
	successCount := 0
	failCount := 0

	for _, projectCode := range projectCodes {
		projectId, err := codecs.DecryptInt64(projectCode)
		if err != nil || projectId == 0 {
			failCount++
			continue
		}
		// 权限校验
		if !authz.IsProjectOwner(db, memberId, projectId) {
			failCount++
			continue
		}
		_ = db.Model(&projectRow{}).Where("id=?", projectId).Updates(map[string]any{
			"deleted":      0,
			"deleted_time": "",
		}).Error
		successCount++
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"successCount": successCount,
		"failCount":    failCount,
		"message":      fmt.Sprintf("成功恢复 %d 个项目，失败 %d 个", successCount, failCount),
	}))
}

// @Summary 项目数据分析
// @Description 获取项目的任务统计、进度分析、成员工作量等数据
// @Tags project
// @Accept x-www-form-urlencoded
// @Produce json
// @Param code formData string true "项目加密ID"
// @Success 200 {object} common.Result "返回项目统计数据"
// @Security ApiKeyAuth
// @Router /project/analysis [post]
func (p *HandlerProject) analysis(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB().WithContext(c.Request.Context())

	// 如果没有传入 projectCode，返回全局数据分析
	if projectCode == "" {
		// 统计用户参与的项目数量
		var projectCount int64
		_ = db.Model(&projectMemberRow{}).Where("member_code=?", memberId).Count(&projectCount)

		// 统计本月新建项目数
		firstOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1).Truncate(24 * time.Hour)
		var monthProjectCount int64
		_ = db.Table("ms_project p").
			Joins("join ms_project_member pm on pm.project_code=p.id").
			Where("pm.member_code=? and p.create_time >= ? and p.deleted=0", memberId, firstOfMonth.UnixMilli()).
			Count(&monthProjectCount)

		// 统计用户的任务总数
		var taskCount int64
		_ = db.Table("ms_task t").
			Joins("join ms_project_member pm on pm.project_code=t.project_code").
			Where("pm.member_code=? and t.deleted=0", memberId).
			Count(&taskCount)

		// 统计已完成的任务数
		var taskDoneCount int64
		_ = db.Table("ms_task t").
			Joins("join ms_project_member pm on pm.project_code=t.project_code").
			Where("pm.member_code=? and t.deleted=0 and t.done=1", memberId).
			Count(&taskDoneCount)

		// 统计逾期任务数
		var taskOverdueCount int64
		now := time.Now().UnixMilli()
		_ = db.Table("ms_task t").
			Joins("join ms_project_member pm on pm.project_code=t.project_code").
			Where("pm.member_code=? and t.deleted=0 and t.done=0 and t.end_time > 0 and t.end_time < ?", memberId, now).
			Count(&taskOverdueCount)

		// 计算完成进度
		var projectSchedule float64
		if taskCount > 0 {
			projectSchedule = float64(taskDoneCount) / float64(taskCount) * 100
		}

		// 计算逾期率
		var taskOverduePercent float64
		if taskCount > 0 {
			taskOverduePercent = float64(taskOverdueCount) / float64(taskCount) * 100
		}

		// 生成项目列表（按月分组）
		projectList := make([]gin.H, 0)
		rows := make([]struct {
			Month string
			Count int64
		}, 0)
		_ = db.Table("ms_project p").
			Joins("join ms_project_member pm on pm.project_code=p.id").
			Select("DATE_FORMAT(FROM_UNIXTIME(p.create_time/1000), '%Y-%m') as month, count(*) as count").
			Where("pm.member_code=? and p.deleted=0", memberId).
			Group("month").
			Order("month desc").
			Limit(12).
			Scan(&rows).Error
		for _, r := range rows {
			projectList = append(projectList, gin.H{
				"日期": r.Month,
				"数量": r.Count,
			})
		}

		// 生成任务列表（按月分组）
		taskList := make([]gin.H, 0)
		taskRows := make([]struct {
			Month string
			Count int64
		}, 0)
		_ = db.Table("ms_task t").
			Joins("join ms_project_member pm on pm.project_code=t.project_code").
			Select("DATE_FORMAT(FROM_UNIXTIME(t.create_time/1000), '%Y-%m') as month, count(*) as count").
			Where("pm.member_code=? and t.deleted=0", memberId).
			Group("month").
			Order("month desc").
			Limit(12).
			Scan(&taskRows).Error
		for _, r := range taskRows {
			taskList = append(taskList, gin.H{
				"日期": r.Month,
				"任务": r.Count,
			})
		}

		// 执行者分布
		executorList := make([]gin.H, 0)
		execRows := make([]struct {
			Name  string
			Count int64
		}, 0)
		_ = db.Table("ms_task t").
			Joins("join ms_project_member pm on pm.project_code=t.project_code").
			Joins("left join ms_member m on m.id=t.assign_to").
			Select("coalesce(m.name, '未指派') as name, count(*) as count").
			Where("pm.member_code=? and t.deleted=0", memberId).
			Group("m.id, m.name").
			Order("count desc").
			Limit(10).
			Scan(&execRows).Error
		for _, r := range execRows {
			executorList = append(executorList, gin.H{
				"name":  r.Name,
				"count": r.Count,
			})
		}

		// 计算本周完成任务数
		nowTime := time.Now()
		weekStart := nowTime.AddDate(0, 0, -int(nowTime.Weekday())).Truncate(24 * time.Hour)
		var weekDoneCount int64
		_ = db.Table("ms_task t").
			Joins("join ms_project_member pm on pm.project_code=t.project_code").
			Where("pm.member_code=? and t.deleted=0 and t.done=1 and t.create_time >= ?", memberId, weekStart.UnixMilli()).
			Count(&weekDoneCount).Error
		weekRate := float64(0)
		if taskCount > 0 {
			weekRate = float64(weekDoneCount) / float64(taskCount) * 100
		}

		// 计算今日完成任务数
		dayStart := time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 0, 0, 0, 0, nowTime.Location())
		var dayDoneCount int64
		_ = db.Table("ms_task t").
			Joins("join ms_project_member pm on pm.project_code=t.project_code").
			Where("pm.member_code=? and t.deleted=0 and t.done=1 and t.create_time >= ?", memberId, dayStart.UnixMilli()).
			Count(&dayDoneCount).Error
		dayRate := float64(0)
		if taskCount > 0 {
			dayRate = float64(dayDoneCount) / float64(taskCount) * 100
		}

		// 按优先级统计任务（priority: 0=普通, 1=紧急, 2=非常紧急, 3=最高）
		var normalCount, urgentCount, veryUrgentCount int64
		_ = db.Table("ms_task t").
			Joins("join ms_project_member pm on pm.project_code=t.project_code").
			Where("pm.member_code=? and t.deleted=0 and (t.priority=0 OR t.priority IS NULL)", memberId).
			Count(&normalCount).Error
		_ = db.Table("ms_task t").
			Joins("join ms_project_member pm on pm.project_code=t.project_code").
			Where("pm.member_code=? and t.deleted=0 and t.priority=1", memberId).
			Count(&urgentCount).Error
		_ = db.Table("ms_task t").
			Joins("join ms_project_member pm on pm.project_code=t.project_code").
			Where("pm.member_code=? and t.deleted=0 and t.priority>=2", memberId).
			Count(&veryUrgentCount).Error

		c.JSON(http.StatusOK, result.Success(gin.H{
			"projectCount":        projectCount,
			"monthProjectCount":   monthProjectCount,
			"taskCount":           taskCount,
			"taskDoneCount":       taskDoneCount,
			"taskOverdueCount":    taskOverdueCount,
			"taskOverduePercent":  taskOverduePercent,
			"projectSchedule":     projectSchedule,
			"projectList":         projectList,
			"taskList":            taskList,
			"executorList":        executorList,
			"weekRate":            weekRate,
			"dayRate":             dayRate,
			"urgentCount":         urgentCount,
			"veryUrgentCount":     veryUrgentCount,
			"normalCount":         normalCount,
		}))
		return
	}

	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	// 查询项目基本信息
	var project projectRow
	if err := db.Where("id=?", projectId).First(&project).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "项目不存在"))
		return
	}
	// 统计任务总数
	var totalTasks int64
	_ = db.Table("ms_task").Where("project_code=? and deleted=0", projectId).Count(&totalTasks)
	// 统计已完成任务数
	var doneTasks int64
	_ = db.Table("ms_task").Where("project_code=? and deleted=0 and done=1", projectId).Count(&doneTasks)
	// 统计成员数量
	var memberCount int64
	_ = db.Model(&projectMemberRow{}).Where("project_code=?", projectId).Count(&memberCount)
	// 计算完成率
	completionRate := 0.0
	if totalTasks > 0 {
		completionRate = float64(doneTasks) / float64(totalTasks) * 100
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"project": gin.H{
			"name":        project.Name,
			"description": project.Description,
			"schedule":    project.Schedule,
			"create_time": project.CreateTime,
		},
		"stats": gin.H{
			"total_tasks":     totalTasks,
			"done_tasks":      doneTasks,
			"member_count":    memberCount,
			"completion_rate": completionRate,
		},
	}))
}

func (p *HandlerProject) projectStats(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}

	db := gorms.GetDB().WithContext(c.Request.Context())
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location()).UnixMilli()

	// 统计项目维度的任务数据
	var total, doneCount, unDone, overdue, toBeAssign, expireToday, doneOverdue int64

	// 总任务数（未删除）
	_ = db.Table("ms_task").Where("project_code=? AND deleted=0", projectId).Count(&total).Error

	// 已完成
	_ = db.Table("ms_task").Where("project_code=? AND deleted=0 AND done=1", projectId).Count(&doneCount).Error

	// 未完成
	unDone = total - doneCount

	// 已逾期（未完成且截止时间小于当前时间）
	_ = db.Table("ms_task").Where("project_code=? AND deleted=0 AND done=0 AND end_time>0 AND end_time<?", projectId, now.UnixMilli()).Count(&overdue).Error

	// 待认领（未指派人）
	_ = db.Table("ms_task").Where("project_code=? AND deleted=0 AND (assign_to=0 OR assign_to IS NULL)", projectId).Count(&toBeAssign).Error

	// 今日到期（未完成且截止时间在今天）
	_ = db.Table("ms_task").Where("project_code=? AND deleted=0 AND done=0 AND end_time>=? AND end_time<=?", projectId, todayStart, todayEnd).Count(&expireToday).Error

	// 逾期完成（已完成但截止时间小于当前时间，需要判断在完成前是否已逾期）
	// 由于缺少完成时间字段，这里用简化逻辑：已完成且截止时间小于现在（即完成前已逾期）
	_ = db.Table("ms_task").Where("project_code=? AND deleted=0 AND done=1 AND end_time>0 AND end_time<?", projectId, now.UnixMilli()).Count(&doneOverdue).Error

	c.JSON(http.StatusOK, result.Success(gin.H{
		"total":        total,
		"unDone":       unDone,
		"done":         doneCount,
		"overdue":      overdue,
		"toBeAssign":   toBeAssign,
		"expireToday":  expireToday,
		"doneOverdue":  doneOverdue,
	}))
}

func (p *HandlerProject) projectReport(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	projectId, err := codecs.DecryptInt64(projectCode)
	if err != nil || projectId == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())

	// 获取项目基本信息
	var project projectRow
	if err := db.Where("id=?", projectId).First(&project).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(404, "项目不存在"))
		return
	}

	// 计算燃尽图数据
	// 1. 确定时间范围：从项目创建日期到当前日期
	createTime := time.UnixMilli(project.CreateTime)
	now := time.Now()
	
	// 计算总天数
	startDate := time.Date(createTime.Year(), createTime.Month(), createTime.Day(), 0, 0, 0, 0, createTime.Location())
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1

	// 2. 获取项目总任务数
	var totalTasks int64
	_ = db.Table("ms_task").Where("project_code=? AND deleted=0", projectId).Count(&totalTasks).Error

	// 3. 查询每日已完成的任务数
	type DailyDone struct {
		Date string `gorm:"column:date"`
		Cnt  int64  `gorm:"column:cnt"`
	}
	var dailyDoneList []DailyDone
	
	// 查询任务完成日志，按日期分组
	// 由于ms_task可能没有done_time字段，我们简化处理：统计每日的累计完成数
	// 这里用ms_task_done_log或类似表，如果没有，用简化方案：查当前未完成数，反推
	_ = db.Table("ms_task").
		Select("DATE(FROM_UNIXTIME(create_time/1000)) as date, COUNT(*) as cnt").
		Where("project_code=? AND deleted=0 AND done=1", projectId).
		Where("create_time >= ?", startDate.UnixMilli()).
		Group("DATE(FROM_UNIXTIME(create_time/1000))").
		Order("date").
		Scan(&dailyDoneList).Error

	// 4. 构建燃尽图数据
	dates := make([]string, 0, totalDays)
	undoneTask := make([]int, 0, totalDays)
	baseLineList := make([]int, 0, totalDays)

	// 按天遍历
	doneMap := make(map[string]int64)
	for _, d := range dailyDoneList {
		doneMap[d.Date] = d.Cnt
	}

	cumulativeDone := int64(0)
	for i := 0; i < totalDays; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		dateStr := currentDate.Format("2006-01-02")
		dates = append(dates, dateStr)

		// 累计完成任务
		if cnt, ok := doneMap[dateStr]; ok {
			cumulativeDone += cnt
		}

		// 剩余未完成
		undone := int(totalTasks - cumulativeDone)
		if undone < 0 {
			undone = 0
		}
		undoneTask = append(undoneTask, undone)

		// 理想基线（线性递减）
		ideal := 0
		if totalDays <= 1 {
			if i == 0 {
				ideal = int(totalTasks)
			}
		} else {
			ideal = int(float64(totalTasks) * (1.0 - float64(i)/float64(totalDays-1)))
			if i == totalDays-1 {
				ideal = 0
			}
		}
		if ideal < 0 {
			ideal = 0
		}
		baseLineList = append(baseLineList, ideal)
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"date":         dates,
		"undoneTask":   undoneTask,
		"baseLineList": baseLineList,
	}))
}

// projectLogRow 项目操作日志
type projectLogRow struct {
	Id           int64 `gorm:"primaryKey;autoIncrement"`
	ProjectCode  int64
	MemberCode   int64
	EventType    string
	EventContent string
	CreateTime   int64
}

func (*projectLogRow) TableName() string { return "ms_project_event" }

type taskBaseRow struct {
	Id          int64 `gorm:"primaryKey;autoIncrement"`
	Name        string
	ProjectCode int64 `gorm:"column:project_code"`
}

func (*taskBaseRow) TableName() string { return "ms_task" }

func extractEventTaskID(eventContent string) int64 {
	lower := strings.ToLower(eventContent)
	for _, key := range []string{"taskid:", "task_id:"} {
		idx := strings.Index(lower, key)
		if idx < 0 {
			continue
		}
		raw := lower[idx+len(key):]
		end := 0
		for end < len(raw) && raw[end] >= '0' && raw[end] <= '9' {
			end++
		}
		if end == 0 {
			continue
		}
		id, _ := strconv.ParseInt(raw[:end], 10, 64)
		return id
	}
	return 0
}

func cleanEventContent(eventContent string) string {
	cleaned := eventContent
	for _, key := range []string{"taskid:", "task_id:"} {
		lower := strings.ToLower(cleaned)
		idx := strings.Index(lower, key)
		if idx < 0 {
			continue
		}
		end := idx + len(key)
		for end < len(cleaned) && cleaned[end] >= '0' && cleaned[end] <= '9' {
			end++
		}
		cleaned = cleaned[:idx] + cleaned[end:]
	}
	return strings.TrimSpace(strings.Trim(cleaned, " ,|;：:"))
}

func parseProjectEventType(eventType, eventContent string) (string, bool) {
	switch eventType {
	case "create", "task:create":
		return "创建了任务", false
	case "done", "task:done":
		return "完成了任务", false
	case "edit", "task:edit":
		return "编辑了任务", false
	case "comment", "task:comment":
		return "发表了评论", true
	case "assign", "task:assign":
		return "指派了任务", false
	case "delete", "task:delete":
		return "删除了任务", false
	case "recovery", "task:recovery":
		return "恢复了任务", false
	case "move", "task:move":
		return "移动了任务", false
	case "priority", "task:priority":
		return "修改了优先级", false
	case "tag", "task:tag":
		return "修改了标签", false
	case "file", "task:file", "upload_file":
		return "上传了附件", false
	default:
		if strings.Contains(strings.ToLower(eventType), "comment") {
			return "发表了评论", true
		}
		if strings.Contains(strings.ToLower(eventType), "task") || extractEventTaskID(eventContent) > 0 {
			return "更新了任务", false
		}
		return "更新了项目", false
	}
}

func (p *HandlerProject) taskSelfList(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB().WithContext(c.Request.Context())
	query := db.Table("ms_task t").
		Joins("join ms_project_member pm on pm.project_code=t.project_code").
		Where("pm.member_code=? and t.deleted=0", memberId)
	var total int64
	_ = query.Count(&total).Error
	var rows []struct {
		Id          int64  `gorm:"column:id"`
		Name        string `gorm:"column:name"`
		ProjectCode int64  `gorm:"column:project_code"`
		Done        int8   `gorm:"column:done"`
		EndTime     int64  `gorm:"column:end_time"`
	}
	_ = query.Select("t.id, t.name, t.project_code, t.done, t.end_time").
		Order("t.id desc").
		Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).
		Scan(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":           r.Id,
			"code":         codecs.EncryptInt64(r.Id),
			"name":         r.Name,
			"project_code": codecs.EncryptInt64(r.ProjectCode),
			"done":         r.Done == 1,
			"end_time":     r.EndTime,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (p *HandlerProject) getLogBySelfProject(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	memberId := c.GetInt64("memberId")
	projectCode := c.PostForm("projectCode")
	orgVal, _ := c.Get("organizationCode")
	orgCode := orgCodeToInt64(orgVal)
	db := gorms.GetDB().WithContext(c.Request.Context())

	projectQuery := db.Table("ms_project_member pm").
		Select("pm.project_code").
		Joins("join ms_project p on p.id=pm.project_code").
		Where("pm.member_code=?", memberId)
	if orgCode != 0 {
		projectQuery = projectQuery.Where("p.organization_code=?", orgCode)
	}
	var projectIds []int64
	_ = projectQuery.Scan(&projectIds).Error
	if len(projectIds) == 0 {
		c.JSON(http.StatusOK, result.Success(gin.H{"list": []any{}, "total": 0}))
		return
	}

	query := db.Model(&projectLogRow{}).Where("project_code IN ?", projectIds)
	if projectCode != "" {
		pid, err := codecs.DecryptInt64(projectCode)
		if err != nil || pid == 0 {
			c.JSON(http.StatusOK, result.Success(gin.H{"list": []any{}, "total": 0}))
			return
		}
		query = query.Where("project_code=?", pid)
	}

	var total int64
	_ = query.Count(&total).Error
	var rows []projectLogRow
	_ = query.Order("id desc").
		Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).
		Find(&rows).Error

	memberIds := make([]int64, 0, len(rows))
	projectIdSet := make(map[int64]struct{})
	taskIdSet := make(map[int64]struct{})
	for _, r := range rows {
		memberIds = append(memberIds, r.MemberCode)
		if r.ProjectCode > 0 {
			projectIdSet[r.ProjectCode] = struct{}{}
		}
		if taskId := extractEventTaskID(r.EventContent); taskId > 0 {
			taskIdSet[taskId] = struct{}{}
		}
	}

	memberMap := make(map[int64]memberBaseRow)
	if len(memberIds) > 0 {
		var members []memberBaseRow
		_ = db.Table("ms_member").Where("id IN ?", memberIds).Find(&members).Error
		for _, member := range members {
			memberMap[member.Id] = member
		}
	}

	projectMap := make(map[int64]projectRow)
	if len(projectIdSet) > 0 {
		ids := make([]int64, 0, len(projectIdSet))
		for id := range projectIdSet {
			ids = append(ids, id)
		}
		var projects []projectRow
		_ = db.Table("ms_project").Where("id IN ?", ids).Find(&projects).Error
		for _, project := range projects {
			projectMap[project.Id] = project
		}
	}

	taskMap := make(map[int64]taskBaseRow)
	if len(taskIdSet) > 0 {
		ids := make([]int64, 0, len(taskIdSet))
		for id := range taskIdSet {
			ids = append(ids, id)
		}
		var tasks []taskBaseRow
		_ = db.Table("ms_task").Select("id, name, project_code").Where("id IN ?", ids).Find(&tasks).Error
		for _, task := range tasks {
			taskMap[task.Id] = task
			if task.ProjectCode > 0 {
				projectIdSet[task.ProjectCode] = struct{}{}
				if _, ok := projectMap[task.ProjectCode]; !ok {
					var project projectRow
					if db.Where("id=?", task.ProjectCode).First(&project).Error == nil {
						projectMap[project.Id] = project
					}
				}
			}
		}
	}

	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		member := memberMap[r.MemberCode]
		project := projectMap[r.ProjectCode]
		taskId := extractEventTaskID(r.EventContent)
		cleanContent := cleanEventContent(r.EventContent)
		remark, isComment := parseProjectEventType(r.EventType, r.EventContent)
		actionType := "project"
		sourceCode := ""
		taskName := ""
		sourceInfo := gin.H{"name": "", "code": ""}
		if taskId > 0 {
			actionType = "task"
			sourceCode = codecs.EncryptInt64(taskId)
			if task, ok := taskMap[taskId]; ok {
				taskName = task.Name
				sourceInfo = gin.H{"name": task.Name, "code": sourceCode}
				if project.Id == 0 && task.ProjectCode > 0 {
					project = projectMap[task.ProjectCode]
				}
			}
		}
		projectId := r.ProjectCode
		if project.Id > 0 {
			projectId = project.Id
		}
		projectName := project.Name
		out = append(out, gin.H{
			"id":            r.Id,
			"project_code":  codecs.EncryptInt64(projectId),
			"project_name":  projectName,
			"member_code":   codecs.EncryptInt64(r.MemberCode),
			"member_name":   member.Name,
			"member_avatar": member.Avatar,
			"event_type":    r.EventType,
			"event_content": r.EventContent,
			"action_type":   actionType,
			"source_code":   sourceCode,
			"task_name":     taskName,
			"sourceInfo":    sourceInfo,
			"remark":        remark,
			"content":       cleanContent,
			"is_comment":    map[bool]int{true: 1, false: 0}[isComment],
			"create_time":   r.CreateTime,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (p *HandlerProject) accountList(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	projectCode := c.PostForm("projectCode")
	pid, err := codecs.DecryptInt64(projectCode)
	if err != nil || pid == 0 {
		c.JSON(http.StatusOK, result.Fail(400, "projectCode无效"))
		return
	}
	db := gorms.GetDB().WithContext(c.Request.Context())
	// 查询项目成员列表
	var rows []struct {
		Id     int64  `gorm:"column:id"`
		Name   string `gorm:"column:name"`
		Avatar string `gorm:"column:avatar"`
		Email  string `gorm:"column:email"`
	}
	var total int64
	query := db.Table("ms_project_member pm").
		Joins("join ms_member m on m.id=pm.member_code").
		Where("pm.project_code=?", pid)
	_ = query.Count(&total).Error
	_ = query.Select("m.id, m.name, m.avatar, m.email").
		Order("pm.id desc").
		Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).
		Scan(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":     r.Id,
			"code":   codecs.EncryptInt64(r.Id),
			"name":   r.Name,
			"avatar": r.Avatar,
			"email":  r.Email,
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total, "authList": []any{}}))
}

func (p *HandlerProject) eventsMyList(c *gin.Context) {
	result := &common.Result{}
	page := &model.Page{}
	page.Bind(c)
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB().WithContext(c.Request.Context())
	var rows []struct {
		Id          int64  `gorm:"column:id"`
		Title       string `gorm:"column:title"`
		BeginTime   string `gorm:"column:begin_time"`
		EndTime     string `gorm:"column:end_time"`
		ProjectCode int64  `gorm:"column:project_code"`
	}
	var total int64
	query := db.Table("ms_project_events e").
		Joins("join ms_project_events_member m on m.events_id=e.id").
		Where("m.member_id=? and e.deleted=0", memberId)
	_ = query.Count(&total).Error
	_ = query.Select("e.id, e.title, e.begin_time, e.end_time, e.project_code").
		Order("e.id desc").
		Limit(int(page.PageSize)).Offset(int((page.Page - 1) * page.PageSize)).
		Scan(&rows).Error
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"id":           r.Id,
			"code":         codecs.EncryptInt64(r.Id),
			"title":        r.Title,
			"begin_time":   r.BeginTime,
			"end_time":     r.EndTime,
			"project_code": codecs.EncryptInt64(r.ProjectCode),
		})
	}
	c.JSON(http.StatusOK, result.Success(gin.H{"list": out, "total": total}))
}

func (p *HandlerProject) notifyNoReads(c *gin.Context) {
	result := &common.Result{}
	memberId := c.GetInt64("memberId")
	db := gorms.GetDB().WithContext(c.Request.Context())
	var notices []struct {
		Id         int64  `gorm:"column:id"`
		Title      string `gorm:"column:title"`
		Content    string `gorm:"column:content"`
		CreateTime int64  `gorm:"column:create_time"`
	}
	_ = db.Table("ms_notify").Where("member_code=? and type=1 and is_read=0", memberId).Order("id desc").Limit(10).Scan(&notices).Error
	var messages []struct {
		Id         int64  `gorm:"column:id"`
		Title      string `gorm:"column:title"`
		Content    string `gorm:"column:content"`
		CreateTime int64  `gorm:"column:create_time"`
	}
	_ = db.Table("ms_notify").Where("member_code=? and type=0 and is_read=0", memberId).Order("id desc").Limit(10).Scan(&messages).Error
	var noticeTotal, messageTotal int64
	_ = db.Table("ms_notify").Where("member_code=? and type=1 and is_read=0", memberId).Count(&noticeTotal).Error
	_ = db.Table("ms_notify").Where("member_code=? and type=0 and is_read=0", memberId).Count(&messageTotal).Error
	if notices == nil {
		notices = []struct {
			Id         int64  `gorm:"column:id"`
			Title      string `gorm:"column:title"`
			Content    string `gorm:"column:content"`
			CreateTime int64  `gorm:"column:create_time"`
		}{}
	}
	if messages == nil {
		messages = []struct {
			Id         int64  `gorm:"column:id"`
			Title      string `gorm:"column:title"`
			Content    string `gorm:"column:content"`
			CreateTime int64  `gorm:"column:create_time"`
		}{}
	}
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list": gin.H{
			"message": messages,
			"notice":  notices,
		},
		"total": noticeTotal + messageTotal,
		"totalSum": gin.H{
			"message": messageTotal,
			"notice":  noticeTotal,
		},
	}))
}

// saveAsTemplate 将项目保存为模板
func (p *HandlerProject) saveAsTemplate(c *gin.Context) {
	result := &common.Result{}
	projectCode := c.PostForm("projectCode")
	name := c.PostForm("name")
	description := c.PostForm("description")

	// 权限校验：只有项目成员才能另存为模板
	projectId, ok := authz.CanOperateProject(c, projectCode)
	if !ok {
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作此项目"))
		return
	}

	if name == "" {
		c.JSON(http.StatusOK, result.Fail(400, "模板名称不能为空"))
		return
	}

	memberId := c.GetInt64("memberId")
	orgVal, _ := c.Get("organizationCode")
	orgCode := orgCodeToInt64(orgVal)
	db := gorms.GetDB().WithContext(c.Request.Context())
	now := time.Now().UnixMilli()

	// 获取项目信息作为封面
	var project projectRow
	_ = db.Where("id=?", projectId).First(&project).Error

	// 创建模板
	type templateRow struct {
		Id               int64  `gorm:"primaryKey;autoIncrement"`
		Name             string
		Description      string
		Cover            string
		CreateTime       int64
		OrganizationCode int64
		MemberCode       int64
		IsSystem         int
	}
	newTemplate := &templateRow{
		Name:             name,
		Description:      description,
		Cover:            project.Cover,
		CreateTime:       now,
		OrganizationCode: orgCode,
		MemberCode:       memberId,
		IsSystem:         0,
	}
	if err := db.Table("ms_project_template").Create(newTemplate).Error; err != nil {
		c.JSON(http.StatusOK, result.Fail(500, "创建模板失败"))
		return
	}

	// 复制任务阶段到模板
	var stages []struct {
		Id   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
		Sort int    `gorm:"column:sort"`
	}
	_ = db.Table("ms_task_stages").Where("project_code=?", projectId).Order("sort asc, id asc").Scan(&stages).Error

	for i, stage := range stages {
		_ = db.Table("ms_task_stages_template").Create(map[string]any{
			"project_template_code": newTemplate.Id,
			"name":                  stage.Name,
			"sort":                  i,
			"create_time":           now,
		}).Error
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":         codecs.EncryptInt64(newTemplate.Id),
		"name":         name,
		"stages_count": len(stages),
	}))
}

// uploadTemplateCover 上传模板封面
func (p *HandlerProject) uploadTemplateCover(c *gin.Context) {
	result := &common.Result{}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(400, "请选择文件"))
		return
	}
	defer file.Close()

	code := c.PostForm("code")
	templateId, _ := codecs.DecryptInt64(code)

	// 简化处理：直接返回文件名作为封面URL
	// 实际项目中应该上传到OSS或本地存储
	coverUrl := "/uploads/template_covers/" + header.Filename

	// 如果有模板ID，更新封面
	if templateId > 0 {
		db := gorms.GetDB()
		_ = db.Table("ms_project_template").Where("id=?", templateId).Update("cover", coverUrl).Error
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"code":  code,
		"cover": coverUrl,
		"url":   coverUrl,
	}))
}

func New() *HandlerProject {
	return &HandlerProject{}
}
