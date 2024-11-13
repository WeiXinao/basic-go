-- 是阅读数，点赞数还是收藏数
for i, v in pairs(ARGV) do
    local res = redis.call("SET", KEYS[i], v, "EX", 70)
    if res ~= "OK" then
        return 1
    end
end

return 0

