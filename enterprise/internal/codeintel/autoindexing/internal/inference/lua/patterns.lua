local patterns = require "internal_patterns"

local M = {}

local new_pattern = function(prefix, pattern, suffix)
    return patterns.backdoor(prefix .. pattern .. suffix)
end

M.new_path_literal = function(pattern)
    return new_pattern("/", pattern, "")
end

M.new_path_segment = function(pattern)
    return new_pattern("", pattern, "/")
end

M.new_path_basename = function(pattern)
    return new_pattern("", pattern, "")
end

M.new_path_extension = function(pattern)
    return new_pattern("*.", pattern, "")
end

M.new_path_combine = function(pattern)
    return patterns.path_combine(pattern)
end

M.new_path_exclude = function(pattern)
    return patterns.path_exclude(pattern)
end

return M
