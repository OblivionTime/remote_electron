import request from '@/utils/request'
//连接服务器
export function ConnectServer(data) {
    return request({
        url: '/connect_server',
        method: 'post',
        data
    })
}
//连接目标机
export function ConnectRemoteDevice(data) {
    return request({
        url: '/remote',
        method: 'post',
        data
    })
}