local patterns = require "internal_patterns"

local M = {}

local new_pattern = function(glob, pathspecs)
    return patterns.backdoor(glob, pathspecs)
end

-- glob:     /BUILD.bazel
-- pathspec:  BUILD.bazel
M.new_path_literal = function(pattern)
    return new_pattern("/" .. pattern, {pattern})
end

-- glob:       web/
-- pathspec:   web/* (root)
-- pathspec: */web/* (non-root)
M.new_path_segment = function(pattern)
    return new_pattern(pattern .. "/", {pattern .. "/*", "*/" .. pattern .. "/*"})
end

-- glob:       gen.go
-- pathspec:   gen.go (root)
-- pathspec: */gen.go (non-root)
M.new_path_basename = function(pattern)
    return new_pattern(pattern, {pattern, "*/" .. pattern})
end

-- glob:     *.md
-- pathspec: *.md
M.new_path_extension = function(pattern)
    return new_pattern("*." .. pattern, {"*." .. pattern})
end

M.new_path_combine = function(pattern)
    return patterns.path_combine(pattern)
end

M.new_path_exclude = function(pattern)
    return patterns.path_exclude(pattern)
end

return M
