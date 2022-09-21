local pattern = require "sg.autoindex.patterns"
local recognizers = require "sg.recognizers"

local indexer = "sourcegraph/lsif-rust"
local outfile = "dump.lsif"

return recognizers.path_recognizer {
  patterns = {
    pattern.new_path_basename "Cargo.toml",
  },

  -- Invoked when Cargo.toml exists anywhere in repository
  generate = function(_, paths)
    return {
      steps = {},
      root = "",
      indexer = indexer,
      indexer_args = { "lsif-rust", "index" },
      outfile = outfile,
    }
  end,
}
