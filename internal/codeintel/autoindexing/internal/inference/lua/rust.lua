local patterns = require "sg.patterns"
local recognizer = require "sg.autoindex.recognizer"

local indexer = "sourcegraph/lsif-rust"
local outfile = "dump.lsif"

return recognizer.new_path_recognizer {
  patterns = {
    patterns.path_basename "Cargo.toml",
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
