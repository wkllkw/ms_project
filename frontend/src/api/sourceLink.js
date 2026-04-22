import $http from '@/assets/js/http'

/*获取资源链接列表*/
export function list(data) {
    return $http.post('project/source_link', data);
}

/*创建资源链接*/
export function save(data) {
    return $http.post('project/source_link/save', data);
}

/*编辑资源链接*/
export function edit(data) {
    return $http.post('project/source_link/edit', data);
}

/*删除资源链接*/
export function del(sourceCode) {
    return $http.post('project/source_link/delete', {sourceCode: sourceCode});
}
