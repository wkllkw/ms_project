package menus

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"test.com/project-api/internal/database/gorms"
	"test.com/project-api/pkg/model/menu"
)

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

type defaultMenuDef struct {
	PreferredID int64
	ParentNode  string
	Title       string
	Icon        string
	Url         string
	FilePath    string
	Params      string
	Node        string
	Sort        int
	Status      int
	CreateBy    int64
	IsInner     int
	Values      string
	ShowSlider  int
}

func defaultMenuDefs() []defaultMenuDef {
	return []defaultMenuDef{
		{PreferredID: 1, Title: "工作台", Icon: "home", Url: "home", FilePath: "home/index", Node: "home", Sort: 1, Status: 1, ShowSlider: 1},
		{PreferredID: 10, Title: "项目", Icon: "project", Url: "project/list", FilePath: "project/list/index", Params: "my", Node: "project", Sort: 2, Status: 1, ShowSlider: 1},
		{PreferredID: 11, ParentNode: "project", Title: "项目列表", Icon: "project", Url: "project/list", FilePath: "project/list/index", Params: "my", Node: "project.list", Sort: 1, Status: 1, ShowSlider: 1},
		{PreferredID: 12, ParentNode: "project", Title: "项目模板", Icon: "appstore", Url: "project/template", FilePath: "project/template/index", Node: "project.template", Sort: 2, Status: 1, ShowSlider: 1},
		{PreferredID: 13, ParentNode: "project", Title: "数据分析", Icon: "bar-chart", Url: "project/analysis", FilePath: "project/analysis/index", Node: "project.analysis", Sort: 3, Status: 1, ShowSlider: 1},
		{PreferredID: 14, ParentNode: "project", Title: "归档项目", Icon: "container", Url: "project/list", FilePath: "project/list/index", Node: "project.archive", Sort: 4, Status: 0, Values: "archive", ShowSlider: 1},
		{PreferredID: 15, ParentNode: "project", Title: "回收站", Icon: "delete", Url: "project/list", FilePath: "project/list/index", Node: "project.recycle", Sort: 5, Status: 0, Values: "deleted", ShowSlider: 1},
		{PreferredID: 20, Title: "通知", Icon: "bell", Url: "notify/notice", FilePath: "notify/notice", Node: "notify", Sort: 4, Status: 1, ShowSlider: 1},
		{PreferredID: 21, ParentNode: "notify", Title: "通知列表", Icon: "bell", Url: "notify/notice", FilePath: "notify/notice", Node: "notify.notice", Sort: 1, Status: 1, ShowSlider: 1},
		{PreferredID: 22, ParentNode: "notify", Title: "系统消息", Icon: "notification", Url: "notify/system", FilePath: "notify/system", Node: "notify.system", Sort: 2, Status: 1, ShowSlider: 1},
		{PreferredID: 25, Title: "日程", Icon: "calendar", Url: "calendar", FilePath: "common/calendar", Node: "calendar", Sort: 3, Status: 1, ShowSlider: 1},
		{PreferredID: 30, Title: "团队", Icon: "team", Url: "members", FilePath: "members/index", Node: "members", Sort: 5, Status: 1, ShowSlider: 1},
		{PreferredID: 31, ParentNode: "members", Title: "成员", Icon: "team", Url: "members", FilePath: "members/index", Node: "members.index", Sort: 1, Status: 1, ShowSlider: 1},
		{PreferredID: 40, Title: "系统", Icon: "setting", Url: "system/account", FilePath: "system/account/index", Node: "system", Sort: 6, Status: 1, ShowSlider: 1},
		{PreferredID: 41, ParentNode: "system", Title: "账号管理", Icon: "user", Url: "system/account", FilePath: "system/account/index", Node: "system.account", Sort: 1, Status: 1, ShowSlider: 1},
		{PreferredID: 42, ParentNode: "system", Title: "菜单管理", Icon: "menu", Url: "system/config/menu", FilePath: "system/config/menu", Node: "system.menu", Sort: 2, Status: 1, ShowSlider: 1},
		{PreferredID: 43, ParentNode: "system", Title: "节点管理", Icon: "apartment", Url: "system/config/node", FilePath: "system/config/node", Node: "system.node", Sort: 3, Status: 1, ShowSlider: 1},
		{PreferredID: 44, ParentNode: "system", Title: "权限管理", Icon: "safety", Url: "system/account/auth", FilePath: "system/account/auth", Node: "system.account.auth", Sort: 4, Status: 1, ShowSlider: 1},
	}
}

func getFullUrl(url string, params string, values string) string {
	if (params != "" && values != "") || values != "" {
		return url + "/" + values
	}
	return url
}

func getInnerText(inner int) string {
	if inner == 0 {
		return "导航"
	}
	if inner == 1 {
		return "内页"
	}
	return ""
}

func getStatus(status int) string {
	if status == 0 {
		return "禁用"
	}
	if status == 1 {
		return "使用中"
	}
	return ""
}

func FindMenus(ctx context.Context) ([]*ProjectMenu, error) {
	var pms []*ProjectMenu
	err := gorms.GetDB().WithContext(ctx).Order("pid,sort asc, id asc").Find(&pms).Error
	return pms, err
}

func BuildTree(pms []*ProjectMenu) []*menu.Menu {
	if len(pms) == 0 {
		return []*menu.Menu{}
	}
	items := make([]*menu.Menu, 0, len(pms))
	byID := map[int64]*menu.Menu{}
	for _, pm := range pms {
		m := &menu.Menu{
			Id:         pm.Id,
			Pid:        pm.Pid,
			Title:      pm.Title,
			Icon:       pm.Icon,
			Url:        pm.Url,
			FilePath:   pm.FilePath,
			Params:     pm.Params,
			Node:       pm.Node,
			Sort:       int32(pm.Sort),
			Status:     int32(pm.Status),
			CreateBy:   pm.CreateBy,
			IsInner:    int32(pm.IsInner),
			Values:     pm.Values,
			ShowSlider: int32(pm.ShowSlider),
		}
		m.StatusText = getStatus(pm.Status)
		m.InnerText = getInnerText(pm.IsInner)
		m.FullUrl = getFullUrl(pm.Url, pm.Params, pm.Values)
		items = append(items, m)
		byID[m.Id] = m
	}
	roots := make([]*menu.Menu, 0)
	for _, m := range items {
		if m.Pid == 0 {
			roots = append(roots, m)
			continue
		}
		if parent, ok := byID[m.Pid]; ok {
			parent.Children = append(parent.Children, m)
		} else {
			roots = append(roots, m)
		}
	}
	return roots
}

func upsertDefaultMenu(ctx context.Context, def defaultMenuDef, parents map[string]*ProjectMenu) (*ProjectMenu, error) {
	db := gorms.GetDB().WithContext(ctx)
	pm := &ProjectMenu{}
	err := db.Where("node = ?", def.Node).First(pm).Error
	pid := int64(0)
	if def.ParentNode != "" {
		if parent, ok := parents[def.ParentNode]; ok {
			pid = parent.Id
		}
	}
	if err == nil {
		pm.Pid = pid
		pm.Title = def.Title
		pm.Icon = def.Icon
		pm.Url = def.Url
		pm.FilePath = def.FilePath
		pm.Params = def.Params
		pm.Node = def.Node
		pm.Sort = def.Sort
		pm.Status = def.Status
		pm.CreateBy = def.CreateBy
		pm.IsInner = def.IsInner
		pm.Values = def.Values
		pm.ShowSlider = def.ShowSlider
		if saveErr := db.Save(pm).Error; saveErr != nil {
			return nil, saveErr
		}
		return pm, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	pm = &ProjectMenu{
		Id:         def.PreferredID,
		Pid:        pid,
		Title:      def.Title,
		Icon:       def.Icon,
		Url:        def.Url,
		FilePath:   def.FilePath,
		Params:     def.Params,
		Node:       def.Node,
		Sort:       def.Sort,
		Status:     def.Status,
		CreateBy:   def.CreateBy,
		IsInner:    def.IsInner,
		Values:     def.Values,
		ShowSlider: def.ShowSlider,
	}
	if createErr := db.Create(pm).Error; createErr != nil {
		pm.Id = 0
		if retryErr := db.Create(pm).Error; retryErr != nil {
			return nil, retryErr
		}
	}
	return pm, nil
}

func SeedDefaultIfEmpty(ctx context.Context) error {
	parents := make(map[string]*ProjectMenu)
	for _, def := range defaultMenuDefs() {
		pm, err := upsertDefaultMenu(ctx, def, parents)
		if err != nil {
			return err
		}
		parents[def.Node] = pm
	}
	return nil
}
