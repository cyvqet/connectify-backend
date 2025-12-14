---@diagnostic disable: undefined-global

-- Rate limiting object
local key = KEYS[1]
-- Window size (milliseconds)
local window = tonumber(ARGV[1])
-- Threshold
local threshold = tonumber(ARGV[2])
-- Current time (milliseconds)
local now = tonumber(ARGV[3])

-- Window start time = (now - window, now]
local min = now - window

-- 1. Delete request records before the window, using "(" to exclude the boundary equal to min
redis.call('ZREMRANGEBYSCORE', key, '-inf', '(' .. min)

-- 2. Current window request count (after cleanup, the entire zset is the data within the window)
local cnt = redis.call('ZCARD', key)

-- 3. If the threshold is reached, execute rate limiting
if cnt >= threshold then
    -- Return string "true" to indicate that the request is rate limited
    return "true"
end

-- 4. Allow through, write current request record
-- To avoid multiple requests in the same millisecond being "overwritten", use an incrementing sequence to ensure member uniqueness
local seqKey = key .. ":seq"
local seq = redis.call('INCR', seqKey)
local member = now .. "-" .. seq

redis.call('ZADD', key, now, member)

-- 5. Set expiration time:
--    - When there are requests in the window, the key will always exist, but old data will be cleaned up
redis.call('PEXPIRE', key, window)
redis.call('PEXPIRE', seqKey, window)

-- Return string "false" to indicate that the request is not rate limited
return "false"
