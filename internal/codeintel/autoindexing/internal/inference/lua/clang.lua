local path = require("path")
local patterns = require("sg.patterns")
local recognizers = require("sg.recognizers")

local indexer = "sourcegraph/lsif-clang"

local is_cmakelist_file = function(base)
    return string.lower(base) == "cmakelists.txt"
end

return recognizers.path_recognizer {
    patterns = {
        patterns.path_extension("cpp"),
        patterns.path_extension("c"),
        patterns.path_extension("h"),
        patterns.path_extension("hpp"),
        patterns.path_extension("cxx"),
        patterns.path_extension("cc"),
        patterns.path_basename("CMakeLists.txt"),
        patterns.path_basename("CMakelists.txt"),
        patterns.path_basename("CmakeLists.txt"),
        patterns.path_basename("Cmakelists.txt"),
        patterns.path_basename("cMakeLists.txt"),
        patterns.path_basename("cMakelists.txt"),
        patterns.path_basename("cmakeLists.txt"),
        patterns.path_basename("cmakelists.txt"),
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
