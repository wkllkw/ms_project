import $http from '@/assets/js/http'

export function chat(data) {
    // http.js 拦截器会将 POST body 转为 form-urlencoded 格式
    // 对象数组需要先序列化为 JSON 字符串，后端再反序列化
    const params = { ...data }
    if (Array.isArray(params.messages)) {
        params.messages = JSON.stringify(params.messages)
    }
    return $http.post('project/assistant/chat', params);
}
