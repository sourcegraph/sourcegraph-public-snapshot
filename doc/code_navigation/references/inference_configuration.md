# Auto-indexing inference configuration reference

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span> This feature is in beta for self-hosted customers.

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>
</p>
</aside>

This document details how a site administrator can supply a Lua script to customize the way [Sourcegraph detects precise code intelligence indexing jobs from repository contents](http://localhost:5080/code_navigation/explanations/auto_indexing_inference).

By default, Sourcegraph will attempt to infer (or hint) index jobs for the following languages:

- `C++`
- [`Go`](../explanations/auto_indexing_inference#go)
- [`Java`/`Scala`/`Kotlin`](../explanations/auto_indexing_inference#java)
- `Python`
- `Ruby`
- [`Rust`](../explanations/auto_indexing_inference#rust)
- [`TypeScript`/`JavaScript`](../explanations/auto_indexing_inference#typescript)

Inference logic can be disabled or altered in the case when the target repositories do not conform to a pattern that the Sourcegraph default inference logic recognizes. Inference logic is controlled by a **Lua override script** that can be supplied in the UI under `Admin > Code graph > Inference`.

> NOTE: While the change is self-service, **Sourcegraph support is more than happy to help write custom behaviors with you**. Do not hesitate to contact us to get the inference logic behaving how you would expect for **your** repositories.

## Example

The **Lua override script** ultimately must return an _auto-indexing config object_. An empty configuration does not change the default behavior.

```lua
return require("sg.autoindex.config").new({
  -- Empty configuration
})
```

To **disable** default behaviors, you can re-assign a recognizer value to `false`. Each of the built-in recognizers are prefixed with `sg.` (and are the only ones allowed to be).

```lua
return require("sg.autoindex.config").new({
  -- Disable default Python inference
  ["sg.python"] = false
})
```

To **add** additional behaviors, you can create and register a new **recognizer**. A recognizer is an interface that requests some set of files from a repository, and returns a set of auto-indexing job configurations that could produce a precise code intelligence index.

A _path recognizer_ is a concrete recognizer that advertises a set of path _globs_ it is interested in, then invokes its `generate` function with matching paths from a repository. In the following, all files matching `Snek.module` (`Snek.module`, `proj/Snek.module`, `proj/sub/Snek.module`, etc) are passed to a call to `generate` (if non-empty). The generate function will then return a list of indexing job descriptions. The [guide for auto-indexing jobs configuration](auto_indexing_configuration#keys-1) gives detailed descriptions on the fields of this object.

```lua
local path = require("path")
local pattern = require("sg.autoindex.patterns")
local recognizer = require("sg.autoindex.recognizer")

local snek_recognizer = recognizer.new_path_recognizer {
  patterns = {
    -- Look for Snek.module files
    pattern.new_path_basename("Snek.module"),

    -- Ignore any files in test or vendor directories
    pattern.new_path_exclude(pattern.new_path_combine {
      pattern.new_path_segment("test"),
      pattern.new_path_segment("vendor"),
    }),
  },

  -- Called with list of matching Snek.module files
  generate = function(_, paths)
    local jobs = {}
    for i = 1, #paths do
      -- Create indexing job description for each matching file
      table.insert(jobs, {
        indexer = "acme/snek:latest",  -- Run this indexer...
        root = path.dirname(paths[i]), -- ...in this directory
        steps = {},
        indexer_args = {},
        outfile = "",
      })
    end

    return jobs
  end
}

return require("sg.autoindex.config").new({
  -- Register new recognizer
  ["acme.snek"] = snek_recognizer,
})
```

## Available libraries

There are a number of specific and general-purpose Lua libraries made accessible via the built-in `require`.

### `sg.autoindex.recognizer`

This auto-indexing-specific library defines the following two functions.

<!-- TODO - document paths_for_content;api.register -->
- `new_path_recognizer` creates a recognizer from a config object containing `patterns` and `generate` fields. See the [example](#example) above for basic usage.
- `new_fallback_recognizer` creates a recognizer from an ordered list of recognizers. Each recognizer is called sequentially, and halts after the recognizer emitting the first non-empty set of results.

### `sg.autoindex.patterns`

This auto-indexing-specific library defines the following four path pattern constructors.

- `new_path_literal(pattern)` creates a pattern that matches an exact filepath.
- `new_path_segment(pattern)` creates a pattern that matches a directory name.
- `new_path_basename(pattern)` creates a pattern that matches a basename exactly.
- `new_path_extension(ext_no_dot)` creates a pattern that matches files with a given extension.

This library also defines the following two pattern collection constructors.

- `new_path_combine(patterns)` creates a pattern collection object (to be used with [recognizers](#sg-autoindex-recognizers)) from the given set of path patterns.
- `new_path_exclude(patterns)` creates a new _inverted_ pattern collection object. Paths matching these patterns are filtered out from the set of matching filepaths given to a recognizer's `generate` function.

### `paths`

This library defines the following five path utility functions:

- `ancestors(path)` returns a path's parent, grandparent, etc as a list.
- `basename(path)` returns the basename of the given path.
- `dirname(path)` returns the dirname of the given path.
- `join(paths)` returns a filepath created by joining the given path segments via filepath separator.
- `split(path)` split a path into an ordered sequence of path segments.

### `json`

This library defines the following two JSON utility functions:

- `encode(val)` returns a JSON-ified version of the given Lua object.
- `decode(json)` returns a Lua table representation of the given JSON text.

### `fun`

[Lua Functional](https://github.com/luafun/luafun/tree/cb6a7e25d4b55d9578fd371d1474b00e47bd29f3#lua-functional) is a high-performance functional programming library accessible via `local fun = require("fun")`. This library has a number of functional utilities to help make recognizer code a bit more expressive.
