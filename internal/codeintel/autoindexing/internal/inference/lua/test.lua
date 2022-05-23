local path = require("path")
local patterns = require("sg.patterns")
local recognizers = require("sg.recognizers")

return recognizers.path_recognizer {
    patterns = { patterns.path_basename("sg-test") },

    -- Invoked as part of unit tests for the autoindexing service
    generate = function(_, paths)
        local jobs = {}
        for i = 1, #paths do
            table.insert(jobs, {
                steps = {},
                root = path.dirname(paths[i]),
                indexer = "test",
                indexer_args = {},
                outfile = "",
            })
        end

        return jobs
    end,
}
