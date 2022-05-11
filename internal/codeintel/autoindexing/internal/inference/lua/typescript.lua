
local json = require("json")
local path = require("path")
local patterns = require("sg.patterns")
local recognizers = require("sg.recognizers")

local shared = loadfile("shared.lua")()
local util = loadfile("util.lua")()

local indexer = "sourcegraph/lsif-typescript:autoindex"
local typescript_nmusl_command = "N_NODE_MIRROR=https://unofficial-builds.nodejs.org/download/release n --arch x64-musl auto"

local exclude_paths = patterns.path_combine(shared.exclude_paths, {
    patterns.path_segment("node_modules"),
})

local check_lerna_file = function(root, contents_by_path)
    local ancestors = path.ancestors(root)
    for i = 1, #ancestors do
        local payload = json.decode(contents_by_path[path.join(ancestors[i], "lerna.json")] or "")
        if payload and payload["npmClient"] == "yarn" then
            return true
        end
    end

    return false
end

local can_derive_node_version = function(root, paths, contents_by_path)
    local ancestors = path.ancestors(root)

    for i = 1, #ancestors do
        local payload = json.decode(contents_by_path[path.join(ancestors[i], "package.json")] or "")
        if payload and payload["engines"] and payload["engines"]["node"] then
            return true
        end
    end

    for i = 1, #ancestors do
        local candidates = {
            path.join(ancestors[i], ".nvmrc"),
            path.join(ancestors[i], ".node-version"),
            path.join(ancestors[i], ".n-node-version"),
        }
        if util.contains_any(paths, candidates) then
            return true
        end
    end

    return false
end

local infer_typescript_job = function(api, tsconfig_path, should_infer_config)
    local root = path.dirname(tsconfig_path)
    local reverse_ancestors = util.reverse(path.ancestors(tsconfig_path))

    api:register(recognizers.path_recognizer {
        patterns = {
            -- To disambiguate installation steps
            patterns.path_basename("yarn.lock"),
            -- Try to determine version
            patterns.path_basename(".n-node-version"),
            patterns.path_basename(".node-version"),
            patterns.path_basename(".nvmrc"),
            -- To reinvoke simple cases with no other files
            patterns.path_basename("tsconfig.json"),
            patterns.path_exclude(exclude_paths),
        },

        patterns_for_content = {
            -- To read explicitly configured engines and npm client
            patterns.path_basename("package.json"),
            patterns.path_basename("lerna.json"),
            patterns.path_exclude(exclude_paths),
        },

        -- Invoked when any of the files listed in the patterns above exist.
        generate = function(api, paths, contents_by_path)
            local is_yarn = check_lerna_file(root, contents_by_path)

            local docker_steps = {}
            for i = 1, #reverse_ancestors do
                if contents_by_path[path.join(reverse_ancestors[i], "package.json")] then
                    local install_command = ""
                    if is_yarn or util.contains(paths, path.join(reverse_ancestors[i], "yarn.lock")) then
                        install_command = "yarn --ignore-engines"
                    else
                        install_command = "npm install"
                    end

                    local install_command_suffix = ""
                    if should_infer_config then
                        install_command_suffix = " --ignore-scripts"
                    end

                    table.insert(docker_steps, {
                        root = reverse_ancestors[i],
                        image = indexer,
                        commands = { install_command .. install_command_suffix },
                    })
                end
            end

            local local_steps = {}
            if can_derive_node_version(root, paths, contents_by_path) then
                for i = 1, #docker_steps do
                    -- Add `n` invocation command before each docker step
                    docker_steps[i].commands = util.with_new_head(docker_steps[i].commands, typescript_nmusl_command)
                end

                -- Add `n` invocation (in indexing container) before indexer runs
                local_steps = { typescript_nmusl_command }
            end

            local args = { "lsif-typescript-autoindex", "index" }
            if should_infer_config then
                table.insert(args, "--infer-tsconfig")
            end

            return {
                steps = docker_steps,
                local_steps = local_steps,
                root = root,
                indexer = indexer,
                indexer_args = args,
                outfile = "",
            }
        end,
    })

    return {}
end

return recognizers.path_recognizer {
    patterns = {
        patterns.path_basename("package.json"),
        patterns.path_basename("tsconfig.json"),
        patterns.path_exclude(exclude_paths),
    },

    -- Invoked when package.json or tsconfig.json files exist
    generate = function(api, paths)
        local has_tsconfig = false
        for i = 1, #paths do
            if path.basename(paths[i]) == "tsconfig.json" then
                -- Infer typescript jobs
                infer_typescript_job(api, paths[i], false)
                has_tsconfig = true
            end
        end

        if not has_tsconfig then
            -- Infer javascript jobs if there's no tsconfig.json found
            infer_typescript_job(api, "tsconfig.json", true)
        end

        return {}
    end,
}
