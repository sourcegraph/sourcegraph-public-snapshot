# glob-expand v0.2.1

A (sync) glob / minimatch / RegExp call using [gruntjs](https://github.com/gruntjs/grunt)'s `file.expand`.

It has only a minimum of dependencies: `glob` (version 4.x that has negation patterns) & `lodash`.

Its almost a copy/paste of 2 functions from Gruntjs's v0.4.1 [grunt/file.js](https://github.com/gruntjs/grunt/blob/master/lib/grunt/file.js)

Additionally you can use [minimatch](http://github.com/isaacs/minimatch/) `String`s or `RegExp`s, either as an Array or as arguments.
*

## Install:

`npm install glob-expand`

## Examples:
```coffeescript
	expand = require 'glob-expand'

	# may the original node-glob be with you (should you need it):
	glob = expand.glob

	expand {filter: 'isFile', cwd: '../'}, ['**/*.*', '!exclude/these/**/*.*']
	# returns all files in cwd ['file1', 'file2',...] but excluding
	# those under directory 'exclude/these'

	# These are the same
	expand {cwd: '../..'}, ['**/*.*', '!node_modules/**/*.*']
	expand {cwd: '../..'}, '**/*.*', '!node_modules/**/*.*'

	# These are the same too:
	expand {}, ['**/*.*', '!**/*.js']
	expand {}, '**/*.*', '!**/*.js'
	expand ['**/*.*', '!**/*.js']
	expand '**/*.*', '!**/*.js'

	# Using Regular Expressions:
	expand '**/*.js', /.*\.(coffee\.md|litcoffee|coffee)$/i, '!DRAFT*.*'
	# -> returns all `.js`, `.coffee`, `.coffee.md` & `.litcoffee` files,
	#    excluding those starting with 'DRAFT'

```

See [gruntjs files configuration](http://gruntjs.com/configuring-tasks#files)
and [node-glob](https://github.com/isaacs/node-glob) for more options.

Sorry no tests, I assumed gruntjs's tests are sufficient ;-)

# License

The MIT License

Copyright (c) 2013-2016 Angelos Pikoulas (agelos.pikoulas@gmail.com)

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the "Software"), to deal in the Software without
restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following
conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
