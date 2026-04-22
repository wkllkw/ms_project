import $http from '@/assets/js/http'

/*收藏项目*/
export function collect(code, type = 'collect') {
    return $http.post('project/project_collect/collect', {type: type, projectCode: code});
}

/*获取收藏项目列表*/
export function list(data) {
    return $http.post('project/project_collect/list', data);
}
