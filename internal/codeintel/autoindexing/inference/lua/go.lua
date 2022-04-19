local path = require("path")
local path_patterns = require("sg.path_patterns")
local recognizers = require("sg.recognizers")

local shared = loadfile("shared.lua")()

local indexer = "sourcegraph/lsif-go:latest"

local exclude_paths = path_patterns.combine(shared.exclude_paths, {
    path_patterns.segment("vendor"),
})

local gomod_recognizer = recognizers.path_recognizer {
    patterns = {
        path_patterns.basename("go.mod"),
        path_patterns.exclude(exclude_paths),
    },

    -- Invoked when go.mod files exist
    generate = function(api, paths)
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
        path_patterns.extension("go"),
        path_patterns.exclude(exclude_paths),
    },

    -- Invoked when no go.mod files exist but go extensions exist somewhere
    -- in the repository. Within this function we filter out files that are
    -- not directly in the root of the repository (the simple pre-mod libs).
    generate = function(api, paths)
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
