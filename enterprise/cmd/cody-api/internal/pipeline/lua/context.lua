-- TODO
local llm = require("llm")
local embeddings = require("embeddings")
local keywords = require("keywords")

local contextType = 'embeddings' -- 'keyword' 'none' 'blended'

local M = {}

local embeddingsContextMessages = function(query, numCodeResults, numTextResults)
    -- TODO - group and map
    return embeddings.search(query, numCodeResults, numTextResults)
end

local keywordContextMessages = function(query)
    -- TODO - group and map
    return keywords.search(query)
end

M.contextMessages = function(query)
    if contextType == 'embeddings' or (contextType == 'blended' and embeddings.available()) then
        return embeddingsContextMessages(query, 12, 3)
    elseif contextType == 'keyword' or (contextType == 'blended' and not embeddings.available()) then
        return keywordContextMessages(query)
    else
        return {}
    end
end

return M
