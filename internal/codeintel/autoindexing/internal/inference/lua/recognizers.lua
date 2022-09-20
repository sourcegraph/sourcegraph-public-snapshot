local languages = {
  "clang",
  "go",
  "java",
  "python",
  "rust",
  "test",
  "typescript",
}

local recognizers = require("sg.autoindex.config").new {}
for _, name in ipairs(languages) do
  -- Backdoor set `sg.`-prefixed recognizers
  rawset(recognizers, "sg." .. name, require("sg.autoindex." .. name))
end

return recognizers
