local path = require "path"
local recognizer = require "sg.autoindex.recognizer"
local pattern = require "sg.autoindex.patterns"

local indexer = require("sg.autoindex.indexes").get "ruby"
local outfile = "index.scip"

return recognizer.new_path_recognizer {
  patterns = {
    pattern.new_path_extension "gemspec",
    pattern.new_path_basename "rubygems-metadata.yml",
    pattern.new_path_basename "Gemfile.lock",
    pattern.new_path_basename "Gemfile",
  },

  generate = function(_api, paths)
    local roots = {}
    for i = 1, #paths do
      roots[path.dirname(paths[i])] = true
    end

    local jobs = {}
    for root in pairs(roots) do
      table.insert(jobs, {
        steps = {},
        root = root,
        indexer = indexer,
        indexer_args = { "scip-ruby-autoindex" },
        outfile = outfile,
      })
    end
    return jobs
  end,
}
