---
--- Generated by EmmyLua(https://github.com/EmmyLua)
--- Created by 小新.
--- DateTime: 2024/8/6 21:48
---
local key = KEYS[1]
-- 用户输入的 code
local expectedCode = ARGV[1]
local code = redis.call("get", key)
local cntKey = key..":cnt"
-- 转成一个数字
local cnt = tonumber(redis.call("get", cntKey))
if cnt <= 0 then
    -- 说明，用户一直输错, 有人搞你
    -- 或者已经用过了，也是有人搞你
    return -1
elseif expectedCode == code then
    -- 输入对了
    redis.call("set", cntKey, -1)
    return 0
else
    -- 用户手一抖，输错了
    redis.call("decr", cntKey, -1)
    return -2
end
