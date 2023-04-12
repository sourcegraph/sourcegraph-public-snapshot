-- Go libs
local log = require("log")
local time = require("time")
local user = require("user")
local llm = require("llm")
local embeddings = require("embeddings")

-- Lua libs
local context = require("sg.cody.context")

local MAX_HUMAN_INPUT_TOKENS = 1000
local MAX_CURRENT_FILE_TOKENS = 1000
local CHARS_PER_TOKEN = 4

local truncateText = function(text, numTokens)
    return string.sub(text, 0, numTokens * CHARS_PER_TOKEN)
end

local editorContext = function()
    -- TODO
    if true then
        return {}
    end

    local fileName = 'TODO'
    local contents = "TODO"
    local language = "TODO" -- path.extname(filePath).slice(1)

    return {{
        speaker = 'human',
        text = "I have the `" .. filePath .. "` file opened in my editor. You are able to answer questions about `" ..
            filePath .. "`. The following code snippet is from the currently open file in my editor `" .. filePath ..
            "`:\n" .. "```" .. language .. "\n" .. contents .. "\n```",
        fileName = filename
    }, {
        speaker = 'assistant',
        text = "You currently have `" .. filePath ..
            "` open in your editor, and I can answer questions about that file's contents."
    }}
end

-- TODO
local intent = {}
intent.isCodebaseContextRequired = function(text)
    return false
end
intent.isEditorContextRequired = function(text)
    return false
end

local contextMessages = function(text)
    local messages = {}

    if intent.isCodebaseContextRequired(text) then
        for _, value in context.contextMessages(text, {
            numCodeResults = 12,
            numTextResults = 3
        }) do
            table.insert(messages, value)
        end
    end

    if intent.isCodebaseContextRequired(text) or intent.isEditorContextRequired(text) then
        table.insert(messages, editorContext())
    end

    return messages
end

local M = {}

M.capabilities = {"chat-input"}
M.setup = function()
    local timestamp = time.shortTimestamp()

    -- "chat-input", "At " .. timestamp

    return {
        timestamp = timestamp
    }
end

M.run = function(ctx)
    -- TODO - get at input
    local text = user.perform("chat-input", "At " .. ctx.timestamp)
    local truncatedText = truncateText(text, MAX_HUMAN_INPUT_TOKENS)

    local chatContext = {{
        speaker = "human",
        text = truncatedText,
        timestamp = ctx.timestamp
    }, {
        speaker = "assistant",
        text = "",
        timestamp = ctx.timestamp
    }}

    for _, message in pairs(contextMessages(truncatedText)) do
        table.insert(chatContext, message)
    end

    return llm.run(chatContext)
end

return M
