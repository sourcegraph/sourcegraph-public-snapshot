local path = require "path"
local pattern = require "sg.autoindex.patterns"

local recognizer = require "sg.autoindex.recognizer"

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
    pattern.new_path_basename "build.gradle",
    pattern.new_path_basename "build.gradle.kts",
    pattern.new_path_basename "gradlew",
    pattern.new_path_basename "settings.gradle",
    -- Maven
    pattern.new_path_basename "pom.xml",
    -- SBT
    pattern.new_path_basename "build.sbt",
    -- Mill
    pattern.new_path_basename "build.sc",
    -- SCIP build tool
    pattern.new_path_basename "lsif-java.json",
  },
  generate = function(api, paths)
    local unique_paths = {}

    for i = 1, #paths do
      unique_paths[path.dirname(paths[i])] = true
    end

    local unique_paths_array = {}

    for path in pairs(unique_paths) do
      table.insert(unique_paths_array, path)
    end

    table.sort(unique_paths_array, function(l, r)
      return string.len(l) < string.len(r)
    end)

    local roots = {}

    for i = 1, #unique_paths_array do
      local project_root = unique_paths_array[i]
      api:register(recognizer.new_path_recognizer {
        patterns = {
          new_rooted_extension(project_root, "java"),
          new_rooted_extension(project_root, "scala"),
          new_rooted_extension(project_root, "kt"),
        },

        generate = function(_, _)
          local is_nested_root = project_root ~= ""
          local is_toplevel_root = project_root == ""
          local top_level_root_is_already_registerd = roots[""] ~= nil

          local this_root_already_registered = roots[project_root] ~= nil

          local job = {
            steps = {},
            root = project_root,
            outfile = "index.scip",
            indexer = java_indexer,
            indexer_args = { "scip-java", "index", "--build-tool=auto" },
          }
          -- top level root should be registered anyways if it has build files and source files
          if is_toplevel_root and not this_root_already_registered then
            roots[project_root] = true
            return job
            -- nested roots are only registered if the top level root WASN'T
            -- this is to account for multi-module builds like ones present in Maven, where
            -- nested roots might have build files but cannot be built independently
            -- in the future these situations should be handled with the auto-indexer itself
          elseif is_nested_root and not this_root_already_registered and not top_level_root_is_already_registerd then
            roots[project_root] = true
            return job
          else
            return {}
          end
        end,
      })
    end

    return {}
  end,
}
