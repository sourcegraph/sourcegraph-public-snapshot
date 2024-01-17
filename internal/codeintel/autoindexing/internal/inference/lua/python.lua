local path = require "path"
local pattern = require "sg.autoindex.patterns"
local recognizer = require "sg.autoindex.recognizer"

local indexer = require("sg.autoindex.indexes").get "python"
local outfile = "index.scip"

-- TODO: This could probably be part of the provided libs once we have that.
local function split_string(inputstr, sep)
  local t = {}

  for str in string.gmatch(inputstr, "([^" .. sep .. "]+)") do
    table.insert(t, str)
  end

  return t
end

--- PKG-INFO looks like:
---
--- ...
--- Metadata-Version: 1.1
--- Name: easydict
--- Version: 1.9
--- ...
local get_name_and_version_from_content = function(content)
  local name, version = nil, nil

  -- So we just assume we can read this and split it by ": "
  --  If we can't find these lines, then we just skip trying to auto-index for now.
  local lines = split_string(content, "\n")
  for _, line in ipairs(lines) do
    local split = split_string(line, ": ")
    if split[1] == "Version" then
      version = split[2]
    elseif split[1] == "Name" then
      name = split[2]
    end

    -- As soon as we have found both name and version, break the loop.
    --  This is so that we just grab the first Name and Version that we can find in the document
    --  to give us the best chance of reading the correct ones.
    if name and version then
      break
    end
  end

  return name, version
end

local handle_one_pkg_info = function(libraries, filepath, content)
  -- Only pkg infos INSIDE of an egg are reliable
  --    That's common knowledge for 1337 python hackers like Me
  --        - Noah
  if not string.find(filepath, ".egg%-info") or not content then
    return
  end

  local name, version = get_name_and_version_from_content(content)

  -- Only index if we know the name and the version
  if not name or not version then
    return
  end

  -- in:    x/y/foo.egg-info/PKG-INFO
  -- out:   x/y/
  local info_parts = split_string(filepath, "/")
  local library_parts = {}
  for i = 1, #info_parts - 2 do
    table.insert(library_parts, info_parts[i])
  end

  table.insert(libraries, {
    root = table.concat(library_parts, "/"),
    name = name,
    version = version,
  })
end

local increase_node_mem_step =
  'if [ -n "${VM_MEM_MB:-}" ]; then export NODE_OPTIONS="--max-old-space-size=$VM_MEM_MB"; fi'

local make_job = function(root, name, version, additional_args)
  return {
    steps = {
      {
        root = "",
        image = indexer,
        -- It's ok if pip install fails, we can still do our best attempt.
        commands = { "pip install . || true" },
      },
    },
    local_steps = { increase_node_mem_step },
    root = root,
    indexer = indexer,
    indexer_args = {
      "scip-python",
      "index",
      ".",
      "--project-name",
      name,
      "--project-version",
      version,
      unpack(additional_args or {}),
    },
    outfile = outfile,
  }
end

return recognizer.new_path_recognizer {
  patterns = {
    pattern.new_path_basename "PKG-INFO",
    pattern.new_path_basename "requirements.txt",
    pattern.new_path_basename "pyproject.toml",
    pattern.new_path_basename "setup.py",
  },

  patterns_for_content = {
    pattern.new_path_basename "PKG-INFO",
  },

  generate = function(_, paths, contents_by_path)
    local roots = {}
    local has_package_info = false
    local libraries = {}

    for i = 1, #paths do
      roots[path.dirname(paths[i])] = true

      if path.basename(paths[i]) == "PKG-INFO" then
        has_package_info = true
        local pkg_info_filepath = paths[i]
        local content = contents_by_path[pkg_info_filepath]
        handle_one_pkg_info(libraries, pkg_info_filepath, content)
      end
    end

    -- If we didn't find any libraries, just insert this as the index job.
    -- Don't worry about excluding or including anything in particular
    if #libraries == 0 and contents_by_path["PKG-INFO"] then
      local name, version = get_name_and_version_from_content(contents_by_path["PKG-INFO"])
      return { make_job("", name, version) }
    end

    -- If we only have one library that we've found, and it's at the root
    -- then we're only going to issue one autoindex job, which will be
    -- run from the root of the library with the found name and version
    if #libraries == 1 and libraries[1].root == "" then
      local lib = libraries[1]
      return make_job(lib.root, lib.name, lib.version)
    end

    -- If we have a base PKG-INFO in addition
    -- to the other PKG-INFOs that are inside of .egg-info,
    -- then we're going to add a job that runs from the base
    -- but excludes all of the paths that we've seen thus far.
    local jobs = {}
    for _, lib in ipairs(libraries) do
      table.insert(jobs, make_job(lib.root, lib.name, lib.version))
    end
    local to_exclude = {}
    for _, lib in ipairs(libraries) do
      table.insert(to_exclude, lib.root)
    end
    -- Sort to_exclude list so the tests are stable
    table.sort(to_exclude)
    local exclude = table.concat(to_exclude, ",")
    if contents_by_path["PKG-INFO"] and exclude ~= "" then
      local name, version = get_name_and_version_from_content(contents_by_path["PKG-INFO"])
      if name and version then
        table.insert(jobs, make_job("", name, version, { "--exclude", exclude }))
      end
    end

    -- Only consider pyproject.toml etc. if we're not looking at a package.
    -- This is because these config files may have been accidentally
    -- bundled in with the package; the PKG-INFO should contain the
    -- canonicalized information regardless of the exact config file.
    if not has_package_info then
      for root in pairs(roots) do
        table.insert(jobs, {
          steps = {},
          local_steps = { "pip install . || true", increase_node_mem_step },
          root = root,
          indexer = indexer,
          indexer_args = { "scip-python", "index" },
          outfile = outfile,
        })
      end
    end

    return jobs
  end,
}
