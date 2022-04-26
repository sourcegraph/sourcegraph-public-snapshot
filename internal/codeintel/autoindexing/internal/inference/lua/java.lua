local patterns = require("sg.patterns")
local recognizers = require("sg.recognizers")

local indexer = "sourcegraph/lsif-java"
local outfile = "dump.lsif"

return recognizers.path_recognizer {
    patterns = {
        patterns.path_literal("lsif-java.json"),
    },

    -- Invoked when lsif-java.json exists in root of repository
    generate = function(api)
        api:register(recognizers.path_recognizer {
            patterns = {
                patterns.path_basename("pom.xml"),
                patterns.path_basename("build.gradle"),
                patterns.path_basename("build.gradle.kts"),
                patterns.path_extension("java"),
                patterns.path_extension("scala"),
                patterns.path_extension("kt"),
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
