-- 1, 2, 3, 4, 5, 6, 7 这是你的元素
-- ZREMRANGEBYSCORE key1 0 6
-- 7 执行完之后

-- 限流对象
local key = KEYS[1]
-- 窗口大小
local window = tonumber(ARGV[1])
-- 阈值
local threshold =  tonumber(ARGV[2])
local now = tonumber(ARGV[3])
-- 窗口起始时间
local min = now - window
-- 挪动窗口
redis.call('ZREMRANGEBYSCORE', key, '-inf', min)
-- 查看容量
local cnt = redis.call('ZCOUNT', key, '-inf', '+inf')
if cnt >= threshold then
    -- 执行限流
    return "true"
else
    -- 把score和member都设置成now
    redis.call('ZADD', key, now, now) -- 将当前时间（now）添加到有序集合 key 中，并为其分配分数 now
    redis.call('PEXPIRE', key, window)
    return "false"
end