local fun = require "fun"
local path = require "path"
local pattern = require "sg.autoindex.patterns"
local recognizer = require "sg.autoindex.recognizer"
local shared = require "sg.autoindex.shared"

local indexer = "sourcegraph/lsif-go:latest"

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
    return fun.totable(fun.map(function(p)
      local root = path.dirname(p)

      return {
        steps = {
          {
            root = root,
            image = indexer,
            commands = { "go mod download" },
          },
        },
        root = root,
        indexer = indexer,
        indexer_args = { "lsif-go", "--no-animation" },
        outfile = "",
      }
    end, paths))
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
    if fun.any(function(p)
      return path.dirname(p) == ""
    end, paths) then
      return {
        steps = {},
        root = "",
        indexer = indexer,
        indexer_args = { "GO111MODULE=off", "lsif-go", "--no-animation" },
        outfile = "",
      }
    end

    return {}
  end,
}

return recognizer.new_fallback_recognizer {
  gomod_recognizer,
  goext_recognizer,
}
