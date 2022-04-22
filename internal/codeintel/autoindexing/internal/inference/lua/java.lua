local path_patterns = require("sg.path_patterns")
local recognizers = require("sg.recognizers")

local indexer = "sourcegraph/lsif-java"
local outfile = "dump.lsif"

return recognizers.path_recognizer {
    patterns = {
        path_patterns.literal("lsif-java.json"),
    },

    -- Invoked when lsif-java.json exists in root of repository
    generate = function(api)
        api:callback(recognizers.path_recognizer {
            patterns = {
                path_patterns.basename("pom.xml"),
                path_patterns.basename("build.gradle"),
                path_patterns.basename("build.gradle.kts"),
                path_patterns.extension("java"),
                path_patterns.extension("scala"),
                path_patterns.extension("kt"),
            },

            -- Invoked when filenames/extensions exist
            generate = function(api, paths)
                return {
                    steps = {},
                    root = "",
                    indexer = indexer,
                    indexer_args = { "lsif-java", "index", "--build-tool=lsif" },
                    outfile = outfile,
                }
            end
        })

        return {}
    end,
}
