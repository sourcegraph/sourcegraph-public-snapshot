local recognizer = require "sg.autoindex.recognizer"
local pattern = require "sg.autoindex.patterns"

local indexer = require("sg.autoindex.indexes").get "rust"
local outfile = "index.scip"

return recognizer.new_path_recognizer {
  patterns = {
    pattern.new_path_basename "Cargo.toml",
  },

  -- Invoked when Cargo.toml exists anywhere in repository
  generate = function(_, paths)
    return {
      steps = {},
      root = "",
      indexer = indexer,
      indexer_args = { "scip-rust", "index" },
      outfile = outfile,
    }
  end,
}
