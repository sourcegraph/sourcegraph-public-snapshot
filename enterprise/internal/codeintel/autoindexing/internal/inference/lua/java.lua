local path = require("path")
local pattern = require("sg.autoindex.patterns")

local is_project_structure_supported = function(base)
  local supported = {
    ["pom.xml"] = true,
    ["build.gradle"] = true,
    ["build.gradle.kts"] = true,
    ["build.sbt"] = true,
    ["build.sc"] = true,
  }
  return supported[base] ~= nil
end

local recognizer = require("sg.autoindex.recognizer")

local java_indexer = require("sg.autoindex.indexes").get "java"


local patterns = require "internal_patterns"

local new_rooted_extension = function(root, ext)
  if root == "" then
    return patterns.backdoor("**/*." .. ext, { "**/*." .. ext })
  else
    return patterns.backdoor("/" .. root .. "/**/*." .. ext, { root .. "/**/*." .. ext })
  end
end

-- This recogniser works in two steps:
-- 1. Identify build roots - paths that contain build files for any of the supported build tools
-- 2. Among those build roots select only those that have any java/scala/kotlin files in there
-- We are doing this to avoid creating an indexing job that will fail because there are no sources.
return recognizer.new_path_recognizer {
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
  generate = function(api, paths)
    local unique_paths = {}
    for i = 1, #paths do
      unique_paths[path.dirname(paths[i])] = true
    end

    local roots = {}

    for project_root in pairs(unique_paths) do
      api:register(recognizer.new_path_recognizer {
        patterns = {
          new_rooted_extension(project_root, "java"),
          new_rooted_extension(project_root, "scala"),
          new_rooted_extension(project_root, "kt"),
        },

        generate = function(_, _)
          if roots[project_root] == nil then
            roots[project_root] = true
            return {
              steps = {},
              root = project_root,
              outfile = "index.scip",
              indexer = java_indexer,
              indexer_args = { "scip-java", "index", "--build-tool=auto" },
            }
          else
            return {}
          end
        end
      })
    end

    return {}
  end,

  hints = function(_, paths)
    local hints = {}
    local visited = {}

    for i = 1, #paths do
      local dir = path.dirname(paths[i])
      local base = path.basename(paths[i])

      if visited[dir] == nil and is_project_structure_supported(base) then
        table.insert(hints, {
          root = dir,
          indexer = java_indexer,
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
          indexer = java_indexer,
          confidence = "LANGUAGE_SUPPORTED",
        })

        visited[dir] = true
      end
    end

    return hints
  end,
}
