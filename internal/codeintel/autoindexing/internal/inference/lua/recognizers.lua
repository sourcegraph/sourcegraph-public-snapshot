local config = require("sg.autoindex.config").new({})

for _, name in ipairs({
  "go",
  "java",
  "python",
  "ruby",
  "rust",
  "test",
  "typescript",
  "dotnet",
}) do
  -- Backdoor set `sg.`-prefixed recognizers
  rawset(config, "sg." .. name, require("sg.autoindex." .. name))
end

return config
