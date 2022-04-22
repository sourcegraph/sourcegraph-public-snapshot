local languages = {
    "go",
    "java",
    "rust",
    "typescript",
}

local recognizers = {}
for _, name in ipairs(languages) do
    recognizers["sg." .. name] = loadfile(name .. ".lua")()
end

return recognizers
