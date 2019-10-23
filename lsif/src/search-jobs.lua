--[[
    Returns raw, flattened job data for each job matching the
    provided query term with the provided status.

    Input:
        KEYS[1] key prefix
        KEYS[2] name of the queue
        ARGV[1] search query
        ARGV[2] number of results to return
        ARGV[3] maximum number of jobs to search
]]

local searchableArguments = {
    'repository',
    'commit',
    'root',
}

local commandsByQueue = {
    ['active'] = 'LRANGE',
    ['waiting'] = 'LRANGE',
    ['delayed'] = 'ZRANGE',
    ['completed'] = 'ZREVRANGE',
    ['failed'] = 'ZREVRANGE',
}

local matching = {}
local numMatching = 0
local offset = 0
local chunkSize = min(tonumber(ARGV[2]), tonumber(ARGV[3]))
local command = commandsByQueue[KEYS[2]] -- determine command for this queue
local terms = string.gmatch(ARGV[1], '%S+') -- split search query into words

-- Extract values that we'll search over from a job.
local function extractValues(id, key)
    local values = {
        id,
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

    for term in terms do
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

-- Search while we don't have enough results and haven't seen too many jobs
while numMatching < tonumber(ARGV[2]) and offset < ARGV[3] do
    local endIndex = min(offset + limit - 1, tonumber(ARGV[3]))
    local ids = redis.call(command, KEYS[1] .. KEYS[2], offset, endIndex)
    offset = endIndex + 1

    for _, v in pairs(ids) do
        if jobMatches(v, KEYS[1] .. v) then
            -- Get job data and add the job id to the payload
            local data = redis.call('HGETALL', KEYS[1] .. v)
            table.insert(data, 'id')
            table.insert(data, v)

            -- Accumulate matching data
            table.insert(matching, data)
            numMatching = numMatching + 1

            if numMatching >= tonumber(ARGV[2]) then
                break
            end
        end
    end
end

return matching
