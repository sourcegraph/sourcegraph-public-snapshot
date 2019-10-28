--[[
    Return the data for jobs that match the given query. Will scan
    between the indices `[ARGV[2], ARGV[2]+ARGV[4])` and return up to
    ARGV[3] results.

    Input:
        KEYS[1] key prefix
        KEYS[2] name of the queue
        ARGV[1] search query
        ARGV[2] the index at which to start the search
        ARGV[3] number of results to return
        ARGV[4] maximum number of jobs to search

    Output:
        an array consisting of
            [1]: an array of hgetall responses for matching jobs
            [2]: the first offset _not_ scanned or nil
]]

local searchableArguments = {
    'repository',
    'commit',
    'root',
}

local commandsByQueue = {
    ['active'] = 'LRANGE',
    ['wait'] = 'LRANGE',
    ['delayed'] = 'ZRANGE',
    ['completed'] = 'ZREVRANGE',
    ['failed'] = 'ZREVRANGE',
}

local prefix = KEYS[1]
local queueName = KEYS[2]
local command = commandsByQueue[queueName]
local query = ARGV[1]
local offset = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local maxJobsToSearch = tonumber(ARGV[4])
local maxIndex = offset + maxJobsToSearch
local chunkSize = math.min(limit, maxJobsToSearch)

local terms = {}

-- split search query into words
for term in string.gmatch(query, '%S+') do
    table.insert(terms, term)
end

-- Extract values that we'll search over from a job.
local function extractValues(id, key)
    local values = {
        redis.call('HGET', key, 'name'),
        redis.call('HGET', key, 'failedReason'),
        redis.call('HGET', key, 'stacktrace'),
    }

    local data = cjson.decode(redis.call('HGET', key, 'data'))
    for _, value in pairs(searchableArguments) do
        table.insert(values, data['args'][value])
    end

    return values
end

-- Determine if a job contains every word in the search query.
local function jobMatches(id, key)
    local values = extractValues(id, key)

    for _, term in pairs(terms) do
        local found = false
        for _, value in pairs(values) do
            -- string.find is weird  the last bool argument enables literal mode
            if type(value) == 'string' and string.find(value, term, 1, true) then
                found = true
                break
            end
        end

        -- Didn't contain this term
        if not found then
            return false
        end
    end

    -- All terms found
    return true
end

local matching = {}
local numMatching = 0

-- Search while we don't have enough results and haven't seen too many jobs
while numMatching < limit and offset <= maxIndex do
    local endIndex = math.min(offset + chunkSize - 1, maxIndex)
    local ids = redis.call(command, prefix .. queueName, offset, endIndex)
    offset = endIndex + 1

    for _, id in pairs(ids) do
        local key = prefix .. id

        if jobMatches(id, key) then
            -- Get job data and add the job id to the payload
            local data = redis.call('HGETALL', key)
            table.insert(data, 'id')
            table.insert(data, id)

            -- Accumulate matching data
            table.insert(matching, data)
            numMatching = numMatching + 1

            if numMatching >= limit then
                break
            end
        end
    end
end

if offset <= maxIndex then
    return {matching, offset}
else
    return {matching, nil}
end
