local path = require("path")
local patterns = require("sg.patterns")
local recognizers = require("sg.recognizers")

local shared = loadfile("shared.lua")()

local indexer = "sourcegraph/lsif-go:latest"

local exclude_paths = patterns.path_combine(shared.exclude_paths, {
    patterns.path_segment("vendor"),
})

local gomod_recognizer = recognizers.path_recognizer {
    patterns = {
        patterns.path_basename("go.mod"),
        patterns.path_exclude(exclude_paths),
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
                        commands = { "go mod download" },
                    },
                },
                root = root,
                indexer = indexer,
                indexer_args = { "lsif-go", "--no-animation" },
                outfile = "",
            })
        end

        return jobs
    end,
}

local goext_recognizer = recognizers.path_recognizer {
    patterns = {
        patterns.path_extension("go"),
        patterns.path_exclude(exclude_paths),
    },

    -- Invoked when no go.mod files exist but go extensions exist somewhere
    -- in the repository. Within this function we filter out files that are
    -- not directly in the root of the repository (the simple pre-mod libs).
    generate = function(_, paths)
        for i = 1, #paths do
            if path.dirname(paths[i]) == "" then
                return {
                    steps = {},
                    root = "",
                    indexer = indexer,
                    indexer_args = { "GO111MODULE=off", "lsif-go", "--no-animation" },
                    outfile = "",
                }
            end
        end

        return {}
    end,
}

return recognizers.fallback_recognizer {
    gomod_recognizer,
    goext_recognizer,
}
