local path = require "path"
local pattern = require "sg.autoindex.patterns"
local recognizer = require "sg.autoindex.recognizer"

local indexer = require("sg.autoindex.indexes").get "clang"

local is_cmakelist_file = function(base)
  return string.lower(base) == "cmakelists.txt"
end

return recognizer.new_path_recognizer {
  patterns = {
    pattern.new_path_extension "cpp",
    pattern.new_path_extension "c",
    pattern.new_path_extension "h",
    pattern.new_path_extension "hpp",
    pattern.new_path_extension "cxx",
    pattern.new_path_extension "cc",
    pattern.new_path_basename "CMakeLists.txt",
    pattern.new_path_basename "CMakelists.txt",
    pattern.new_path_basename "CmakeLists.txt",
    pattern.new_path_basename "Cmakelists.txt",
    pattern.new_path_basename "cMakeLists.txt",
    pattern.new_path_basename "cMakelists.txt",
    pattern.new_path_basename "cmakeLists.txt",
    pattern.new_path_basename "cmakelists.txt",
  },

  -- Invoked when c, cpp, header, or cmakelist files exist
  hints = function(_, paths)
    local hints = {}
    local visited = {}

    for i = 1, #paths do
      local dir = path.dirname(paths[i])
      local base = path.basename(paths[i])

      if visited[dir] == nil and is_cmakelist_file(base) then
        table.insert(hints, {
          root = dir,
          indexer = indexer,
          confidence = "PROJECT_STRUCTURE_SUPPORTED",
        })

        visited[dir] = true
      end
    end

    for i = 1, #paths do
      local dir = path.dirname(paths[i])
      local base = path.basename(paths[i])

      if visited[dir] == nil and not is_cmakelist_file(base) then
        table.insert(hints, {
          root = dir,
          indexer = indexer,
          confidence = "LANGUAGE_SUPPORTED",
        })

        visited[dir] = true
      end
    end

    return hints
  end,
}
