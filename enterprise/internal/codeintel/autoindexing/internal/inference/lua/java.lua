local path = require "path"
local recognizer = require "sg.autoindex.recognizer"
local pattern = require "sg.autoindex.patterns"

local indexer = require("sg.indexes").get "java"
local outfile = "index.scip"

local is_project_structure_supported = function(base)
  return base == "pom.xml" or base == "build.gradle" or base == "build.gradle.kts"
end

return recognizer.new_path_recognizer {
  patterns = {
    pattern.new_path_extension "java",
    pattern.new_path_extension "scala",
    pattern.new_path_extension "kt",
    pattern.new_path_basename "pom.xml",
    pattern.new_path_basename "build.gradle",
    pattern.new_path_basename "build.gradle.kts",
  },

  -- Invoked when Java, Scala, Kotlin, or Gradle build files exist
  generate = function(api)
    api:register(recognizer.new_path_recognizer {
      patterns = {
        pattern.new_path_literal "lsif-java.json",
      },

      -- Invoked when lsif-java.json exists in root of repository
      generate = function(api, paths)
        return {
          steps = {},
          root = "",
          indexer = indexer,
          indexer_args = { "scip-java", "index", "--build-tool=scip" },
          outfile = outfile,
        }
      end,
    })

    return {}
  end,

  -- Invoked when Java, Scala, Kotlin, or Gradle build files exist
  hints = function(_, paths)
    local hints = {}
    local visited = {}

    for i = 1, #paths do
      local dir = path.dirname(paths[i])
      local base = path.basename(paths[i])

      if visited[dir] == nil and is_project_structure_supported(base) then
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

      if visited[dir] == nil and not is_project_structure_supported(base) then
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
