local fun = require "fun"
local json = require "json"
local path = require "path"
local pattern = require "sg.autoindex.patterns"
local recognizer = require "sg.autoindex.recognizer"
local shared = require "sg.autoindex.shared"

local indexer = "sourcegraph/scip-typescript:autoindex"
local n_node_mirror = "https://unofficial-builds.nodejs.org/download/release"
local typescript_nmusl_command = "N_NODE_MIRROR=" .. n_node_mirror .. " n --arch x64-musl auto"
local node_derivable_basenames = {
  ".nvmrc",
  ".node-version",
  ".n-node-version",
}

local exclude_paths = pattern.new_path_combine(shared.exclude_paths, {
  pattern.new_path_segment "node_modules",
})

local safe_decode = function(encoded)
  local _, payload = pcall(function()
    return json.decode(encoded)
  end)
  return payload
end

local check_lerna_file_contents = function(contents)
  local payload = safe_decode(contents)
  if payload and payload["npmClient"] == "yarn" then
    return true
  end

  return false
end

local check_lerna_file = function(root, contents_by_path)
  return fun.any(function(a)
    return check_lerna_file_contents(contents_by_path[path.join(a, "lerna.json")] or "")
  end, path.ancestors(root))
end

local check_package_json_contents = function(contents)
  local payload = safe_decode(contents)
  if payload and payload["engines"] and payload["engines"]["node"] then
    return true
  end

  return false
end

local contains = function(table, element)
  return fun.any(function(v)
    return v == element
  end, table)
end

local can_derive_node_version = function(root, paths, contents_by_path)
  return fun.any(function(a)
    if check_package_json_contents(contents_by_path[path.join(a, "package.json")] or "") then
      return true
    end

    return fun.any(function(v)
      return contains(paths, path.join(a, v))
    end, node_derivable_basenames)
  end, path.ancestors(root))
end

local reverse = function(slice)
  local reversed = {}
  for i = #slice, 1, -1 do
    reversed[#reversed + 1] = slice[i]
  end

  return reversed
end

local infer_typescript_job = function(api, tsconfig_path, should_infer_config)
  local root = path.dirname(tsconfig_path)
  local reverse_ancestors = reverse(path.ancestors(tsconfig_path))
  local basenames = {
    -- To disambiguate installation steps
    "yarn.lock",
    -- To reinvoke simple cases with no other files
    "tsconfig.json",
  }
  for _, name in ipairs(node_derivable_basenames) do
    table.insert(basenames, name)
  end

  api:register(recognizer.new_path_recognizer {
    patterns = {
      pattern.new_path_basename_set(basenames),
      pattern.new_path_exclude(exclude_paths),
    },

    patterns_for_content = {
      -- To read explicitly configured engines and npm client
      pattern.new_path_basename_set { "package.json", "lerna.json" },
      pattern.new_path_exclude(exclude_paths),
    },

    -- Invoked when any of the files listed in the patterns above exist.
    generate = function(api, paths, contents_by_path)
      local is_yarn = check_lerna_file(root, contents_by_path)

      local docker_steps = fun.totable(fun.map(function(ra)
        if contents_by_path[path.join(ra, "package.json")] then
          local install_command = ""
          if is_yarn or contains(paths, path.join(ra, "yarn.lock")) then
            install_command = "yarn --ignore-engines"
          else
            install_command = "npm install"
          end

          local install_command_suffix = ""
          if should_infer_config then
            install_command_suffix = " --ignore-scripts"
          end

          return {
            root = ra,
            image = indexer,
            commands = { install_command .. install_command_suffix },
          }
        end
      end, reverse_ancestors))

      local local_steps = {}
      if can_derive_node_version(root, paths, contents_by_path) then
        docker_steps = fun.totable(fun.map(function(s)
          -- Add `n` invocation command before each docker step
          s.commands = fun.totable(fun.chain({ typescript_nmusl_command }, s.commands))
          return s
        end, docker_steps))

        -- Add `n` invocation (in indexing container) before indexer runs
        local_steps = { typescript_nmusl_command }
      end

      local args = { "scip-typescript", "index" }
      if should_infer_config then
        table.insert(args, "--infer-tsconfig")
      end

      return {
        steps = docker_steps,
        local_steps = local_steps,
        root = root,
        indexer = indexer,
        indexer_args = args,
        outfile = "index.scip",
      }
    end,
  })

  return {}
end

return recognizer.new_path_recognizer {
  patterns = {
    pattern.new_path_basename_set { "package.json", "tsconfig.json" },
    pattern.new_path_exclude(exclude_paths),
  },

  -- Invoked when package.json or tsconfig.json files exist
  generate = function(api, paths)
    local has_tsconfig = false
    fun.each(function(p)
      if path.basename(p) == "tsconfig.json" then
        -- Infer typescript jobs
        infer_typescript_job(api, p, false)
        has_tsconfig = true
      end
    end, paths)

    if not has_tsconfig then
      -- Infer javascript jobs if there's no tsconfig.json found
      infer_typescript_job(api, "tsconfig.json", true)
    end

    return {}
  end,
}
