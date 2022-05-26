local languages = {
    "clang",
    "go",
    "java",
    "rust",
    "test",
    "typescript",
}

local recognizers = require("sg.autoindex.config").new({})
for _, name in ipairs(languages) do
  -- Using `rawset` to skip over validation preventing OTHER people
  -- from setting things starting with `sg.`. I'M the only one allowed
  -- to do that!
  rawset(recognizers, "sg." .. name, require("sg.autoindex." .. name))
end

return recognizers
