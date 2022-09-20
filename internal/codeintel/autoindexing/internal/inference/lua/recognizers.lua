local languages = {
  "clang",
  "go",
  "java",
  "python",
  "rust",
  "test",
  "typescript",
}

local recognizers = {}
for _, name in ipairs(languages) do
  recognizers["sg." .. name] = require("sg.autoindex." .. name)
end

return recognizers
