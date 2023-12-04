local internal_path = require "internal_path"

local M = {}

-- type: (string) -> array[string]
M.ancestors = function(path)
  local ax = internal_path.ancestors(path)

  local t = {}
  for i = 1, #ax do
    table.insert(t, ax[i])
  end

  return t
end

-- type: (string) -> string
M.basename = function(path)
  return internal_path.basename(path)
end

-- type: (string) -> string
M.dirname = function(path)
  return internal_path.dirname(path)
end

-- type: (string, string) -> string
M.join = function(p1, p2)
  return internal_path.join(p1, p2)
end

-- type: string -> string, string
M.split = function(path)
  return M.dirname(path), M.basename(path)
end

return M
