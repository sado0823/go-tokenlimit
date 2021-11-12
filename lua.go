package main

const (
	tokenLimiter =`local rate = tonumber(ARGV[1])
local cap = tonumber(ARGV[2]) -- capacity
local now = tonumber(ARGV[3]) -- now timestamp
local requested = tonumber(ARGV[4]) -- needed token
local fill_time = cap/rate
local ttl = math.floor(fill_time*2) -- give more time for more stability

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

redis.call("setex", KEYS[1], ttl, new_tokens)
redis.call("setex", KEYS[2], ttl, now)

return allowed`
)

