local bucket_quota_key = KEYS[1]
local replenish_interval_seconds_key = KEYS[2]
local burst_key = KEYS[3]
local bucket_quota = tonumber(ARGV[1])
local bucket_replenish_interval = tonumber(ARGV[2])
local allowed_burst = tonumber(ARGV[3])

redis.call('SET', bucket_quota_key, bucket_quota, 'EX', 86400)
redis.call('SET', replenish_interval_seconds_key, bucket_replenish_interval, 'EX', 86400)
redis.call('SET', burst_key, allowed_burst, 'EX', 86400)
