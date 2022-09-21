local path = require "path"
local pattern = require "sg.autoindex.patterns"
local recognizer = require "sg.autoindex.recognizer"

return recognizer.new_path_recognizer {
  patterns = {
    pattern.new_path_basename "sg-test",
  },

  -- Invoked as part of unit tests for the autoindexing service
  generate = function(_, paths)
    local jobs = {}
    for _, p in ipairs(paths) do
      table.insert(jobs, {
        steps = {},
        root = path.dirname(p),
        indexer = "test",
        indexer_args = {},
        outfile = "",
      })
    end

    return jobs
  end,
}
