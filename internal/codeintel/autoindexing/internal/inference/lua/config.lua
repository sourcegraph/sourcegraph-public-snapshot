-- Makes settings and/or reading values from the table error
--  This prevents a whole host of simple footguns for users
--  (and we can access w/ rawget and rawset)
local config_mt = {
  __index = function(_, key)
    error("unknown key for recognizers: " .. tostring(key))
  end,

  __newindex = function(_, key, _)
    error("unable to set values on recognizer config: " .. tostring(key))
  end,
}

local M = {}

M.is_config = function(tbl)
  return type(tbl) == "table" and getmetatable(tbl) == config_mt
end

M.new = function(config)
  assert(type(config) == "table",  "config must be a table, got: " .. tostring(config))

  for key, v in pairs(config) do
    if string.sub(key, 1, 3) == "sg." then
      if value ~= false then
        error("Only allowed to set `sg.` prefixed recognizers to false")
      end
    end
  end

  config.__type = "sg.recognizer"
  return setmetatable(config, config_mt)
end

return M
