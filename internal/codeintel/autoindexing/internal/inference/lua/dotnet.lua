local path = require("path")
local pattern = require("sg.autoindex.patterns")
local recognizer = require("sg.autoindex.recognizer")

local env_steps = {

  -- Need to set for dotnet restore to work in docker
  -- https://github.com/dotnet/runtime/issues/97828
  "export DOTNET_EnableWriteXorExecute=0",
}

local dotnet_proj_recognizer = recognizer.new_path_recognizer({
  patterns = {
    -- Find all project files
    pattern.new_path_extension("csproj"),
    pattern.new_path_extension("vbproj"),
  },

  generate = function(_, paths)
    local jobs = {}
    for i = 1, #paths do
      table.insert(jobs, {
        indexer = "sourcegraph/scip-dotnet",
        root = path.dirname(paths[i]),
        local_steps = env_steps,
        indexer_args = { "scip-dotnet", "index", paths[i], "--output", "index.scip" },
        outfile = "index.scip",
      })
    end

    return jobs
  end,
})

local dotnet_sln_recognizer = recognizer.new_path_recognizer({
  patterns = {
    -- Find all solution files
    pattern.new_path_extension("sln"),
  },

  generate = function(_, paths)
    local jobs = {}
    for i = 1, #paths do
      table.insert(jobs, {
        indexer = "sourcegraph/scip-dotnet",
        root = path.dirname(paths[i]),
        local_steps = env_steps,
        indexer_args = { "scip-dotnet", "index", paths[i], "--output", "index.scip" },
        outfile = "index.scip",
      })
    end

    return jobs
  end,
})

return recognizer.new_fallback_recognizer({
  dotnet_sln_recognizer,
  dotnet_proj_recognizer,
})
