import $http from '@/assets/js/http'

export function list(data) {
    return $http.post('project/project_template', data);
}

export function doData(data) {
    let url = 'project/project_template/save';
    if (data.code) {
        url = 'project/project_template/edit';
    }
    return $http.post(url, data);
}
export function del(code) {
    return $http.post('project/project_template/delete', {code: code});
}

/*上传模板封面*/
export function uploadCover(data) {
    return $http.post('project/project_template/uploadCover', data);
}
