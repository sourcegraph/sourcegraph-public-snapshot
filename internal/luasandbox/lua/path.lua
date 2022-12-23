local internal_path = require "internal_path"

local M = {}

M.ancestors = function(path)
  local ax = internal_path.ancestors(path)

  local t = {}
  for i = 1, #ax do
    table.insert(t, ax[i])
  end

  return t
end

M.basename = function(path)
  return internal_path.basename(path)
end

M.dirname = function(path)
  return internal_path.dirname(path)
end

M.join = function(p1, p2)
  return internal_path.join(p1, p2)
end

M.split = function(path)
  return M.dirname(path), M.basename(path)
end

return M
