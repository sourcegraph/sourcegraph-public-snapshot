local path = require("path")
local pattern = require("sg.autoindex.patterns")
local recognizer = require("sg.autoindex.recognizer")

local indexer = require("sg.autoindex.indexes").get("dotnet")

local env_steps = {

  -- macOS enables W^X (https://en.wikipedia.org/wiki/W%5EX) which means
  -- that when running in development, we need to set the following environment
  -- variable for `dotnet restore` to work correctly. It is technically not needed on
  -- Linux in production, and removing this might improve performance, but we
  -- currently do not have a way of conditionally passing in OS-level configuration
  -- to the Lua script, and doing so would create a divergence between dev vs prod,
  -- so leave this as-is for now.
  --
  -- See also: https://github.com/dotnet/runtime/issues/97828
  "export DOTNET_EnableWriteXorExecute=0",
}

local dotnet_proj_recognizer = recognizer.new_path_recognizer({
  patterns = {
    pattern.new_path_extension("csproj"),
    pattern.new_path_extension("vbproj"),
  },

  generate = function(_, paths)
    local jobs = {}
    for i = 1, #paths do
      table.insert(jobs, {
        indexer = indexer,
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
    pattern.new_path_extension("sln"),
  },

  generate = function(_, paths)
    local jobs = {}
    for i = 1, #paths do
      table.insert(jobs, {
        indexer = indexer,
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
