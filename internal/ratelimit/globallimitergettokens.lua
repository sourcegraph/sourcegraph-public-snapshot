local bucket_key = KEYS[1]
local last_replenishment_timestamp_key = KEYS[2]
local bucket_rate_key = KEYS[3]
local bucket_replenishment_interval_key = KEYS[4]
local burst_key = KEYS[5]
local current_time = tonumber(ARGV[1])
local max_time_to_wait_for_tokens = tonumber(ARGV[2])
local default_rate = tonumber(ARGV[3])
local default_replenishment_interval = tonumber(ARGV[4])
local default_burst = tonumber(ARGV[5])
local tokens_to_grant = tonumber(ARGV[6])

-- Ensure the bucket burst capacity is configured. Otherwise,
-- fall back to the provided default.
local burst_exists = redis.call('EXISTS', burst_key)
if burst_exists == 0 then
  redis.call('SET', burst_key, default_burst)
end
redis.call('EXPIRE', burst_key, 86400)

local burst = tonumber(redis.call('GET', burst_key))

-- Check if the bucket exists.
local bucket_exists = redis.call('EXISTS', bucket_key)
-- If the bucket does not exist, create the bucket, fill it up to burst, and set
-- the last replenishment time.
if bucket_exists == 0 then
    redis.call('SET', bucket_key, burst)
    redis.call('SET', last_replenishment_timestamp_key, current_time)
end
redis.call('EXPIRE', bucket_key, 86400)
redis.call('EXPIRE', last_replenishment_timestamp_key, 86400)

-- Check if bucket quota key and replenishment interval keys both exist
local rate_exists = redis.call('EXISTS', bucket_rate_key)
local bucket_replenishment_interval_exists = redis.call('EXISTS', bucket_replenishment_interval_key)
if rate_exists == 0 or bucket_replenishment_interval_exists == 0 then
	-- Otherwise, use default values.
	redis.call('SET', bucket_rate_key, default_rate)
	redis.call('SET', bucket_replenishment_interval_key, default_replenishment_interval)
end
redis.call('EXPIRE', bucket_rate_key, 86400)
redis.call('EXPIRE', bucket_replenishment_interval_key, 86400)

local bucket_rate = tonumber(redis.call('GET', bucket_rate_key))

if bucket_rate == -1 then
  return {1, 0} -- -1 means rate.Inf, we return early, and don't track usage further.
end

if bucket_rate == 0 then
  return {-3, 0} -- -3: You shall not pass.
end

-- Calculate the time difference in seconds since last replenishment
local last_replenishment_timestamp = tonumber(redis.call('GET', last_replenishment_timestamp_key))
local time_difference = current_time - last_replenishment_timestamp
-- Shouldn't happen, but check just in case.
if time_difference < 0 then
	return {-2, 0} -- Return -2 (negative time difference)
end

-- Get the rate (tokens/second) that the bucket should replenish
local bucket_replenishment_interval = tonumber(redis.call('GET', bucket_replenishment_interval_key))
local replenishment_rate = bucket_rate / bucket_replenishment_interval

-- Calculate the amount of tokens to replenish, round down for the number of 'full' tokens.
local num_tokens_to_replenish = math.floor(replenishment_rate * time_difference)

-- Get the current token count in the bucket.
local current_tokens = tonumber(redis.call('GET', bucket_key))

-- Replenish the bucket if there are tokens to replenish
if num_tokens_to_replenish > 0 then
    local available_capacity = burst - current_tokens
    if available_capacity > 0 then
      -- The number of tokens we add is either the number of tokens we have replenished over
      -- the last time_difference, or enough tokens to refill the bucket completely, whichever
      -- is lower.
		  current_tokens = math.min(burst, current_tokens + num_tokens_to_replenish)
    	redis.call('SET', bucket_key, current_tokens)
    	redis.call('SET', last_replenishment_timestamp_key, current_time)
    end
end

local time_to_wait_for_tokens = 0
-- This is for calculations with us removing a token.
local tokens_after_consumption = current_tokens - tokens_to_grant

-- If the bucket will be negative, calculate the needed to 'wait' before using the token.
-- i.e. if we are going to be at -15 tokens after this consumption, and the token replenishment
-- rate is 0.33/s, then we need to wait 45.45 (46 because we round up) seconds before making the request.
if tokens_after_consumption < 0 then
    time_to_wait_for_tokens = math.ceil((tokens_after_consumption * -1) / replenishment_rate)
end

-- If the deadline is not infinite, and the wait time exceeds the deadline, return -1.
if max_time_to_wait_for_tokens ~= -1 and time_to_wait_for_tokens >= max_time_to_wait_for_tokens then
    return {-1, time_to_wait_for_tokens} -- Return -1 (token grant wait time exceeds limit)
end

-- Decrement the token bucket by tokens_to_grant, we are granted the tokens
redis.call('DECRBY', bucket_key, tokens_to_grant)

return {1, time_to_wait_for_tokens}
