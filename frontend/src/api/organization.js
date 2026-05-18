import $http from '@/assets/js/http'
import store from '@/store';

export function list(data) {
    return $http.post('project/organization', data);
}

export function _getOrgList(data) {
    return $http.post('project/organization/_getOrgList', data).then(res => {
        if (res.data) {
            store.dispatch('setOrganizationList', res.data);
        }
        return Promise.resolve(res);
    });
}

export function doData(data) {
    let url = 'project/organization/save';
    if (data.organizationCode) {
        url = 'project/organization/edit';
    }
    return $http.post(url, data);
}

export function del(organizationCode) {
    return $http.post('project/organization/delete', {organizationCode: organizationCode});
}
export function _quitOrganization(data) {
    return $http.post('project/organization/_quitOrganization', data);
}

// ===== 组织级角色管理 API =====

// 获取组织成员列表（含角色信息）
export function _listMembers(data) {
    return $http.post('project/organization/_listMembers', data);
}

// 设置组织成员角色
export function _setMemberAuth(data) {
    return $http.post('project/organization/_setMemberAuth', data);
}

// 移除组织成员角色
export function _removeMemberAuth(data) {
    return $http.post('project/organization/_removeMemberAuth', data);
}

// 获取指定成员在组织中的角色
export function _getMemberAuth(data) {
    return $http.post('project/organization/_getMemberAuth', data);
}
