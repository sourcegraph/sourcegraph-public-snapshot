local fun = require "fun"
local path = require "path"
local pattern = require "sg.autoindex.patterns"
local recognizer = require "sg.autoindex.recognizer"

return recognizer.new_path_recognizer {
  patterns = {
    pattern.new_path_basename "sg-test",
  },

  -- Invoked as part of unit tests for the autoindexing service
  generate = function(_, paths)
    return fun.totable(fun.map(function(p)
      return {
        steps = {},
        root = path.dirname(p),
        indexer = "test",
        indexer_args = {},
        outfile = "",
      }
    end, paths))
  end,
}
