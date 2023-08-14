local pattern_lib = require "internal_patterns"

local M = {}

-- type: (string, array[string]) -> pattern
local new_pattern = function(glob, pathspecs)
    return pattern_lib.backdoor(glob, pathspecs)
end

-- glob:     /BUILD.bazel
-- pathspec:  BUILD.bazel
-- type: (string) -> pattern
M.new_path_literal = function(globlike)
    return new_pattern("/" .. globlike, {globlike})
end

-- glob:       web/
-- pathspec:   web/* (root)
-- pathspec: */web/* (non-root)
-- type: (string) -> pattern
M.new_path_segment = function(globlike)
    return new_pattern(globlike .. "/", {globlike .. "/*", "*/" .. globlike .. "/*"})
end

-- glob:       gen.go
-- pathspec:   gen.go (root)
-- pathspec: */gen.go (non-root)
-- type: (string) -> pattern
M.new_path_basename = function(globlike)
    return new_pattern(globlike, {globlike, "*/" .. globlike})
end

-- glob:     *.md
-- pathspec: *.md
-- type: (string) -> pattern
M.new_path_extension = function(globlike)
    return new_pattern("*." .. globlike, {"*." .. globlike})
end

-- type: ((pattern | table[pattern])...) -> pattern
M.new_path_combine = function(one_or_more_patterns)
    return pattern_lib.path_combine(one_or_more_patterns)
end

-- type: ((pattern | table[pattern])...) -> pattern
M.new_path_exclude = function(one_or_more_patterns)
    return pattern_lib.path_exclude(one_or_more_patterns)
end

return M
