local rate = tonumber(ARGV[1])
local cap = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])
local fill_time = cap/rate
local ttl = math.floor(fill_time*2)

-- KEYS[1]: token key
local last_tokens = tonumber(redis.call("get",KEYS[1]))
if last_tokens == nil then
	last_tokens = cap
end

-- KEYS[2]: token refreshed timestamp
local last_refreshed = tonumber(redis.call("get",KEYS[2]))
if last_refreshed == nil then
	last_refreshed = 0
end

local delta = math.max(0, now-last_refreshed)
local left = math.min(cap,last_tokens+delta*rate)
local new_tokens = left
local allowed = left >= requested
if allowed then
  new_tokens = left - requested
end

redis.call("setex", KEYS[1], new_tokens)
redis.call("setex", KEYS[2], last_refreshed)