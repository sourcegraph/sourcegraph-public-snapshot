local config = require("sg.autoindex.config").new {}

for _, name in ipairs {
  "clang",
  "go",
  "java",
  "python",
  "rust",
  "test",
  "typescript",
} do
  -- Backdoor set `sg.`-prefixed recognizers
  rawset(config, "sg." .. name, require("sg.autoindex." .. name))
end

return config
