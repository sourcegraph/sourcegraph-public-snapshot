local path = require "path"
local recognizer = require "sg.autoindex.recognizer"
local pattern = require "sg.autoindex.patterns"

local indexer = require("sg.autoindex.indexes").get "java"
local outfile = "index.scip"

local is_project_structure_supported = function(base)
  local supported = {
    ['pom.xml'] = true,
    ['build.gradle'] = true,
    ['build.gradle.kts'] = true,
    ['build.sbt'] = true,
    ['build.sc'] = true
  }
  return supported[base] ~= nil
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
        -- Gradle
        pattern.new_path_basename("build.gradle"),
        pattern.new_path_basename("build.gradle.kts"),
        pattern.new_path_basename("gradlew"),
        pattern.new_path_basename("settings.gradle"),
        -- Maven
        pattern.new_path_basename("pom.xml"),
        -- SBT
        pattern.new_path_basename("build.sbt"),
        -- Mill
        pattern.new_path_basename("build.sc"),
        -- SCIP build tool
        pattern.new_path_basename("lsif-java.json")
      },

      -- Invoked when lsif-java.json exists in root of repository
      generate = function(_, paths)
        local jobs = {}
        local unique_paths = {}
        for i = 1, #paths do
          unique_paths[path.dirname(paths[i])] = true
        end

        for root_path in pairs(unique_paths) do
          table.insert(jobs, {
            steps = {},
            root = root_path,
            outfile = outfile,
            indexer = indexer,
            indexer_args = { "scip-java", "index", "--build-tool=auto" },
          })
        end

        return jobs
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
