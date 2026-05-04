package dal

import (
	"context"
	"fmt"
)

// LuaStockDeduct Lua 原子扣减库存 + 防重
// KEYS[1] = seckill:stock:{activityId}   (String)
// KEYS[2] = seckill:users:{activityId}   (Set)
// ARGV[1] = userId
// 返回：1=成功, 0=库存不足, -1=已参与
const LuaStockDeduct = `
local stockKey = KEYS[1]
local usersKey = KEYS[2]
local userId = ARGV[1]

if redis.call('sismember', usersKey, userId) == 1 then
    return -1
end

local stock = tonumber(redis.call('get', stockKey))
if stock == nil or stock <= 0 then
    return 0
end

redis.call('decr', stockKey)
redis.call('sadd', usersKey, userId)

return 1
`

func EvalLuaStockDeduct(ctx context.Context, activityId int64, userId string) (int64, error) {
	stockKey := fmt.Sprintf("seckill:stock:%d", activityId)
	usersKey := fmt.Sprintf("seckill:users:%d", activityId)

	result, err := RDB.Eval(ctx, LuaStockDeduct, []string{stockKey, usersKey}, userId).Result()
	if err != nil {
		return 0, err
	}
	return result.(int64), nil
}
