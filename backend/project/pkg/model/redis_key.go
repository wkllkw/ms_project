package model

import "time"

var (
	RegisterRedisKey     = "REGISTER_"
	CollectRedisKey      = "COLLECT_"      // 收藏状态缓存 key 前缀, 格式: COLLECT_{memberId}_{projectCode}
	CollectRedisExpire   = 7 * 24 * time.Hour   // 收藏缓存过期时间: 7天
)
