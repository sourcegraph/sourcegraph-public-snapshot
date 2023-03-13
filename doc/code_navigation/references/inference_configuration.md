# Auto-indexing inference configuration reference

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span> This feature is in beta for self-hosted customers.
</p>
</aside>

This document details SOMETHING.

Default indexers:

- `clang`
- `go`
- `java`
- `python`
- `ruby`
- `rust`
- `test`
- `typescript`

## Example

TODO

```lua
local path = require("path")
local pattern = require("sg.autoindex.patterns")
local recognizer = require("sg.autoindex.recognizer")

local custom_recognizer = recognizer.new_path_recognizer {
  patterns = { pattern.new_path_basename("sg-test") },

  -- Invoked when go.mod files exist
  generate = function(_, paths)
    local jobs = {}
    for i = 1, #paths do
      table.insert(jobs, {
        steps = {},
        root = path.dirname(paths[i]),
        indexer = "test-override",
        indexer_args = {},
        outfile = "",
      })
    end

    return jobs
  end,
}

return require("sg.autoindex.config").new({
  ["custom.test"] = custom_recognizer,
})
```

## Libraries

TODO

### `sg.autoindex.recognizer`

TODO

```
M.new_path_recognizer = function(config)
M.new_fallback_recognizer = function(config)
```

### `sg.autoindex.patterns`

TODO

```
M.new_path_literal = function(pattern)
M.new_path_segment = function(pattern)
M.new_path_basename = function(pattern)
M.new_path_extension = function(pattern)
```
