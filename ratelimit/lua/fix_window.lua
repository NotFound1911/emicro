-- 查询有没有限流对象的设置
local val = redis.call('get', KEYS[1]) -- 获取Redis中指定键的值
local expiration = ARGV[1]  -- 获取第一个参数作为过期时间
local limit = tonumber(ARGV[2]) -- 获取第二个参数作为限流阈值，并转换为数字
if val == false then

    if limit < 1 then -- 如果限流阈值小于1
        -- 执行限流
        return "true"
    else
        -- set your_service 1 px 100s
        -- 设置限流对象的值为1，过期时间为expiration（以秒为单位）
        redis.call('set', KEYS[1], 1, 'PX', expiration)
        -- 不执行限流
        return "false"
    end

elseif tonumber(val) < limit then

    -- 有这个限流对象，但尚未达到阈值
    redis.call('incr', KEYS[1])
    -- 不指定限流
    return "false"

else

    -- 执行限流
    return "true"

end