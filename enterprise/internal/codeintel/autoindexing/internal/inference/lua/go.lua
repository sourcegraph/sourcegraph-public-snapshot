local path = require "path"
local recognizer = require "sg.autoindex.recognizer"
local pattern = require "sg.autoindex.patterns"

local shared = require "sg.autoindex.shared"

local indexer = require("sg.autoindex.indexes").get "go"

local netrc_steps = [[
if [ "$NETRC_DATA" ]; then
  echo "Writing netrc config to $HOME/.netrc"
  echo "$NETRC_DATA" > ~/.netrc
else
  echo "No netrc config set, continuing"
fi
]]

local exclude_paths = pattern.new_path_combine(shared.exclude_paths, {
  pattern.new_path_segment "vendor",
})

local gomod_recognizer = recognizer.new_path_recognizer {
  patterns = {
    pattern.new_path_basename "go.mod",
    pattern.new_path_exclude(exclude_paths),
  },

  -- Invoked when go.mod files exist
  generate = function(_, paths)
    local jobs = {}
    for i = 1, #paths do
      local root = path.dirname(paths[i])

      table.insert(jobs, {
        steps = {
          {
            root = root,
            image = indexer,
            commands = { netrc_steps, "go mod download" },
          },
        },
        local_steps = { netrc_steps },
        root = root,
        indexer = indexer,
        indexer_args = { "scip-go", "--no-animation" },
        outfile = "index.scip",
        requested_envvars = { "GOPRIVATE", "GOPROXY", "GONOPROXY", "GOSUMDB", "GONOSUMDB", "NETRC_DATA" },
      })
    end

    return jobs
  end,
}

local goext_recognizer = recognizer.new_path_recognizer {
  patterns = {
    pattern.new_path_extension "go",
    pattern.new_path_exclude(exclude_paths),
  },

  -- Invoked when no go.mod files exist but go extensions exist somewhere
  -- in the repository. Within this function we filter out files that are
  -- not directly in the root of the repository (the simple pre-mod libs).
  generate = function(_, paths)
    for i = 1, #paths do
      if path.dirname(paths[i]) == "" then
        return {
          local_steps = { netrc_steps },
          root = "",
          indexer = indexer,
          indexer_args = { "GO111MODULE=off", "scip-go", "--no-animation" },
          outfile = "index.scip",
          requested_envvars = { "GOPRIVATE", "GOPROXY", "GONOPROXY", "GOSUMDB", "GONOSUMDB", "NETRC_DATA" },
        }
      end
    end

    return {}
  end,
}

return recognizer.new_fallback_recognizer {
  gomod_recognizer,
  goext_recognizer,
}
