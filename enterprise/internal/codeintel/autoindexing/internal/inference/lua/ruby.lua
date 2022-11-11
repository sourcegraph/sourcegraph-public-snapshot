local path = require("path")
local recognizer = require("sg.autoindex.recognizer")
local pattern = require("sg.autoindex.patterns")

local indexer = require("sg.indexes").get("ruby")
local outfile = "index.scip"

local pattern_list = {
  pattern.new_path_extension("gemspec"),
  pattern.new_path_basename("rubygems-metadata.yml"),
  pattern.new_path_basename("Gemfile.lock"),
  pattern.new_path_basename("Gemfile"),
}

return recognizer.new_path_recognizer {
  patterns = pattern_list,

  generate = function(api, paths)
    local roots = {}
    for i = 1, #paths do
      roots[path.dirname(paths[i])] = true
    end
    for root in pairs(roots) do
      api:register(recognizer.new_path_recognizer({
        patterns = pattern_list, -- unused; only present to trigger generate
        generate = function(api, _paths)
          return {
            steps = {},
            root = root,
            indexer = indexer,
            indexer_args = { "scip-ruby-autoindex" },
            outfile = outfile,
          }
        end,
      }))
    end
    return {}
  end
}
