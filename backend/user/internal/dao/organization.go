package dao

import (
	"context"
	"test.com/project-user/internal/data/organization"
	"test.com/project-user/internal/database"
	"test.com/project-user/internal/database/gorms"
)

type OrganizationDao struct {
	conn *gorms.GormConn
}

func (o *OrganizationDao) FindOrganizationByMemId(ctx context.Context, memId int64) ([]*organization.Organization, error) {
	var orgs []*organization.Organization
	// 方式1: 用户创建的组织
	var createdOrgs []*organization.Organization
	err := o.conn.Session(ctx).Where("member_id=?", memId).Find(&createdOrgs).Error
	if err != nil {
		return nil, err
	}

	// 方式2: 用户通过部门参与的组织
	// 通过 ms_department_member -> ms_department -> ms_organization 查找
	var joinedOrgIds []int64
	err = o.conn.Session(ctx).
		Table("ms_department_member").
		Select("DISTINCT ms_department.organization_code").
		Joins("JOIN ms_department ON ms_department.id = ms_department_member.department_id").
		Where("ms_department_member.member_id=? AND ms_department.deleted=0", memId).
		Scan(&joinedOrgIds).Error
	if err != nil {
		return nil, err
	}

	// 合并组织ID
	orgIdSet := make(map[int64]bool)
	for _, org := range createdOrgs {
		orgIdSet[org.Id] = true
	}
	for _, id := range joinedOrgIds {
		orgIdSet[id] = true
	}

	// 如果没有任何组织，返回空列表
	if len(orgIdSet) == 0 {
		return []*organization.Organization{}, nil
	}

	// 查询所有相关组织
	allOrgIds := make([]int64, 0, len(orgIdSet))
	for id := range orgIdSet {
		allOrgIds = append(allOrgIds, id)
	}
	err = o.conn.Session(ctx).Where("id IN ?", allOrgIds).Find(&orgs).Error
	return orgs, err
}

func NewOrganizationDao() *OrganizationDao {
	return &OrganizationDao{
		conn: gorms.New(),
	}
}

func (o *OrganizationDao) SaveOrganization(conn database.DbConn, ctx context.Context, org *organization.Organization) error {
	o.conn = conn.(*gorms.GormConn)
	return o.conn.Tx(ctx).Create(org).Error
}
