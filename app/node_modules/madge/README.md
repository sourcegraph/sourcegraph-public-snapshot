# MaDGe - Module Dependency Graph

[![Build Status](https://secure.travis-ci.org/pahen/madge.png)](http://travis-ci.org/pahen/madge)

Create graphs from your [CommonJS](http://nodejs.org/api/modules.html), [AMD](https://github.com/amdjs/amdjs-api/wiki/AMD) or [ES6](https://people.mozilla.org/~jorendorff/es6-draft.html) module dependencies. Could also be useful for finding circular dependencies in your code. Tested on [Node.js](http://nodejs.org/) and [RequireJS](http://requirejs.org/) projects. Dependencies are calculated using static code analysis. CommonJS dependencies are found using James Halliday's [detective](https://github.com/substack/node-detective), for AMD I'm using [amdetective](https://www.npmjs.org/package/amdetective) and for ES6 [detective-es6](https://www.npmjs.com/package/detective-es6) is used. Modules written in [CoffeeScript](http://coffeescript.org/) with extension .coffee are supported and will automatically be compiled on-the-fly.

## Examples
Here's a very simple example of a generated image.

![](https://github.com/pahen/node-madge/raw/master/examples/small.png)

 - blue = has dependencies
 - green = has no dependencies
 - red = has circular dependencies

Here's an example generated from the [Express](https://github.com/visionmedia/express) project.

![](https://github.com/pahen/node-madge/raw/master/examples/express.png)

And some terminal usage.

![](https://github.com/pahen/node-madge/raw/master/examples/terminal.png)

# Installation

To install as a library:

	$ npm install madge

To install the command-line tool:

	$ sudo npm -g install madge

## Graphviz (optional)

Only required if you want to generate the visual graphs using [Graphviz](http://www.graphviz.org/).

### Mac OS X

	$ sudo port install graphviz

### Ubuntu

	$ sudo apt-get install graphviz

# API

	var madge = require('madge');
	var dependencyObject = madge('./');
	console.log(dependencyObject.tree);

## madge(src, opts)

{Object|Array|String} **src** (required)

- Object - a dependency tree.
- Array - an Array of directories to scan.
- String - a directory to scan.

{Object} **opts** (optional)

- {String} **format**. The module format to expect, 'cjs', 'amd' or 'es6'. Commonjs (cjs) is the default format.
- {String} **exclude**. String from which a regex will be constructed for excluding files from the scan.
- {Boolean} **breakOnError**. True if the parser should stop on parse errors and when modules are missing, false otherwise. Defaults to false.
- {Boolean} **optimized**. True if the parser should read modules from a optimized file (r.js). Defaults to false.
- {Boolean} **findNestedDependencies**. True if nested dependencies should be found in AMD modules. Defaults to false.
- {String} **mainRequireModule**. Name of the module if parsing an optimized file (r.js), where the main file used `require()` instead of `define`. Defaults to `''`.
- {String} **requireConfig**. Path to RequireJS config used to find shim dependencies and path aliases. Not used by default.
- {Function} **onParseFile**. Function to be called when parsing a file (argument will be an object with "filename" and "src" property set).
- {Function} **onAddModule** . Function to be called when adding a module to the module tree (argument will be an object with "id" and "dependencies" property set).
- {Array} **extensions**. List of file extensions which are considered. Defaults to `['.js']`.

## dependency object (returned from madge)

#### .opts

Options object passed used in the constructor.

#### .tree

Dependency tree object. Can be overwritten with an object in the format:

	{
	     'module1': ['dep1a', 'dep1b'],
	     'module2': ['dep2a']
	}

#### .obj()

Alias to the tree property.

#### .circular()

Returns all the modules with circular dependencies.

#### .depends()

Returns a list of modules that depends on a given module.

#### .dot()

Get a DOT representation of the module dependency graph.

#### .image(opts, callback)

Get an image representation of the module dependency graph.

- {Object} **opts** (required).
	- {String} **layout**. The layout to use. Defaults to 'DOT'.
	- {String} **fontFace**. The font face to use. Defaults to 'Times-Roman'.
	- {Object} **imageColors**. Object with color information (all colors are strings containing hex values).
		- {String} **bgcolor**. The backgound color.
		- {String} **edge**. The edge color.
		- {String} **dependencies**. The color for dependencies and for text if fontColor is not present.
		- {String} **fontColor**. The color for text.
- {Function} **callback** (required). Receives the rendered image as the first argument.

# CLI

	Usage: madge [options] <file|dir ...>

	Options:

	-h, --help                       output usage information
	-V, --version                    output the version number
	-f, --format <name>              format to parse (amd/cjs/es6)
	-s, --summary                    show summary of all dependencies
	-c, --circular                   show circular dependencies
	-d, --depends <id>               show modules that depends on the given id
	-x, --exclude <regex>            a regular expression for excluding modules
	-t, --dot                        output graph in the DOT language
	-i, --image <filename>           write graph to file as a PNG image
	-l, --layout <name>              layout engine to use for image graph (dot/neato/fdp/sfdp/twopi/circo)
	-b, --break-on-error             break on parse errors & missing modules
	-n, --no-colors                  skip colors in output and images
	-r, --read                       skip scanning folders and read JSON from stdin
	-C, --config <filename>          provide a config file
	-R, --require-config <filename>  include shim dependencies and path aliases found in RequireJS config file
	-O, --optimized                  if given file is optimized with r.js
	-M  --main-require-module        name of the primary RequireJS module, if it's included with `require()`
	-j  --json                       output dependency tree in json


## Examples:

### List all module dependencies (CommonJS)

	$ madge /path/src

### List all module dependencies (AMD)

	$ madge --format amd /path/src

### List all module dependencies (ES6)

	$ madge --format es6 /path/src

### Finding circular dependencies

	$ madge --circular /path/src

### Show modules that depends on a given module

	$ madge --depends 'wheels' /path/src

### Excluding modules

	$ madge --exclude '^foo$|^bar$|^tests' /path/src

### Save graph as a PNG image (graphviz required)

	$ madge --image graph.png /path/src

### Save graph as a [DOT](http://en.wikipedia.org/wiki/DOT_language) file for further processing (graphviz required)

	$ madge --dot /path/src > graph.gv

### Run on optimized file by r.js (RequireJS optimizer)
	$ r.js -o app-build.js
	$ madge --format amd --optimized app-build.js

### Include shim dependencies found in RequireJS config
	$ madge --format amd --require-config path/config.js path/src

### Pipe predefined results (the example image was produced with the following command)

	$ cat << EOF | madge --read --image example.png
	{
		"a": ["b", "c", "d"],
		"b": ["c"],
		"c": [],
		"d": ["a"]
	}
	EOF

## Config (use with --config)

	{
	    "format": "amd",
	    "image": "dependencyMap.png",
	    "fontFace": "Arial",
	    "fontSize": "14px",
	    "imageColors": {
	        "noDependencies" : "#0000ff",
	        "dependencies" : "#00ff00",
	        "circular" : "#bada55",
	        "edge" : "#666666",
	        "bgcolor": "#ffffff"
	    }
	}

# FAQ

## The image produced by madge is very hard to read, what's wrong?

Try running madge with a different layout, here's a list of the ones you can try:

* **dot**	"hierarchical" or layered drawings of directed graphs. This is the default tool to use if edges have directionality.

* **neato** "spring model'' layouts.  This is the default tool to use if the graph is not too large (about 100 nodes) and you don't know anything else about it. Neato attempts to
minimize a global energy function, which is equivalent to statistical multi-dimensional scaling.

* **fdp**	"spring model'' layouts similar to those of neato, but does this by reducing forces rather than working with energy.

* **sfdp** multiscale version of fdp for the layout of large graphs.

* **twopi** radial layouts, after Graham Wills 97. Nodes are placed on concentric circles depending their distance from a given root node.

* **circo** circular layout, after Six and Tollis 99, Kauffman and Wiese 02. This is suitable for certain diagrams of multiple cyclic structures, such as certain telecommunications networks.

# Running tests

	$ npm test

# Release Notes

## v0.5.2 (October 16, 2015)
Updated dependency resolve to latest version.

## v0.5.1 (October 15, 2015)
Updated dependencies to newer versions (Thanks to Martin Kapp).

## v0.5.0 (April 2, 2015)
Added support for ES6 modules (Thanks to Marc Laval).
Added support for setting custom file extension name (Thanks to Marc Laval).

## v0.4.1 (December 19, 2014)
Fixed issues with absolute paths for modules IDs in Windows (all tests should now pass on Windows too).

## v0.4.0 (December 19, 2014)
Add support for JSX (React) and additional module paths (Thanks to Ben Lowery).
Fix for detecting presence of AMD or CommonJS modules (Thanks to Aaron Russ).
Now resolves the module IDs from the RequireJS paths-config properly (Thanks to russaa).
Added support for option findNestedDependencies to find nested dependencies in AMD modules.

## v0.3.5 (Septemper 22, 2014)
Fix issue with number of graph node lines increased with each render (Thanks to Colin H. Fredericks).

## v0.3.4 (Septemper 04, 2014)
Correctly detect circular dependencies when using path aliases in RequireJS config (Thanks to Nicolas Ramz).

## v0.3.3 (July 11, 2014)
Fixed bug with relative paths in AMD not handled properly when checking for cyclic dependencies.

## v0.3.2 (June 25, 2014)
Handle anonymous require() as entry in the RequireJS optimized file (Thanks to Benjamin Horsleben).

## v0.3.1 (June 03, 2014)
Apply exclude to RequireJS shim dependencies (Thanks to Michael White).

## v0.3.0 (May 25, 2014)
Added support for onParseFile and onAddModule options (Thanks to Brandon Selway).
Added JSON output option (Thanks to Drew Foehn).
Fix for optimized files including dependency information for excluded modules (Thanks to Drew Foehn). Fixes [issue](https://github.com/pahen/madge/issues/26).

## v0.2.0 (April 17, 2014)
Added support for including shim dependencies found in RequiredJS config (specify with option -R).

## v0.1.9 (February 17, 2014)
Ensure forward slashes are used in modules paths (Windows).

## v0.1.8 (January 27, 2014)
Added support for reading AMD dependencies from a r.js optimized file by using option -O.

## v0.1.7 (September 20, 2013)
Added missing fontsize option when generating images.

## v0.1.6 (September 04, 2013)
AMD plugins are now ignored as dependencies. Fixes [issue](https://github.com/pahen/grunt-madge/issues/1).

## v0.1.5 (September 04, 2013)
Fixed Windows [issue](https://github.com/pahen/node-madge/issues/17) when reading from standard input with --read.

## v0.1.4 (January 10, 2013)
Switched library for walking directory tree which should solve issues on [Windows](https://github.com/pahen/node-madge/issues/8).

## v0.1.3 (December 28, 2012)
Added proper exit code when running "madge --circular" so it can be used in build scripts.

## v0.1.2 (November 15, 2012)
Relative AMD module identifiers (if the first term is "." or "..") are now resolved.

## v0.1.1 (September 3, 2012)
Tweaked circular dependency path output.

## v0.1.0 (September 3, 2012)
Complete path in circular dependencies is now printed (and marked as red in image graphs).

## v0.0.5 (August 8, 2012)
Added support for CoffeeScript. Files with extension .coffee will automatically be compiled on-the-fly.

## v0.0.4 (August 17, 2012)
Fixed dependency issues with Node.js v0.8.

## v0.0.3 (July 01, 2012)
Added support for Node.js v0.8 and dropped support for lower versions.

## v0.0.2 (May 21, 2012)
Added ability to read config file and customize colors.

## v0.0.1 (May 20, 2012)
Initial release.

# License

(The MIT License)

Copyright (c) 2012 Patrik Henningsson &lt;patrik.henningsson@gmail.com&gt;

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
