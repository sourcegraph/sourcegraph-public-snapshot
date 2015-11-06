# amdetective

Find all calls to `require()` in AMD modules by walking the AST.

This module uses code extracted from [r.js](https://github.com/jrburke/r.js) rather than trying to write it's own version of r.js parsing. It depends on esprima (but not r.js).

# Install

```
npm install amdetective
```

# example

First, create `detect.js` which is just a four line CLI wrapper around `amdetective`:

````js
var fs = require('fs'),
    amdetective = require('amdetective');

console.log('Reading file from first argument: ' + process.argv[2]);
console.log(amdetective(fs.readFileSync(process.argv[2]).toString()));
````

Now, let's run it on a bunch of examples to see some output. You can also run this command on your own files to get more realistic examples.

## Definition Functions with Dependencies (simple.js)

````js
require(['module1', 'path/to/module2'], function(a, b){
  // ...
});
````

Running `node detect.js simple.js` produces:

````
Reading file from first argument: simple.js
[ 'module1', 'path/to/module2' ]
````

## Simplified CommonJS Wrapper (simple2.js)

````js
define(function(require) {
  var a = require('some/file'),
      b = require('json!foo/bar');
  // ...
});
````

Running `node detect.js simple2.js` produces:

````
Reading file from first argument: simple2.js
[ 'require', 'some/file', 'json!foo/bar' ]
````

## Named module (named.js)

````js
define("foo/title",
    ["my/cart", "my/inventory"],
    function(cart, inventory) {
   }
);
````

Running `node detect.js simple2.js` produces:

````
Reading file from first argument: named.js
[ { name: 'foo/title', deps: [ 'my/cart', 'my/inventory' ] } ]
````

Note how named modules are treated differently - this is just something that the underlying resolution code does so be prepared to deal with it.

# Methods

## amdetective(src, opts)

Given some source body `src`, return an array of all the `require()` call arguments detected by AMD/r.js.

The options parameter `opts` is passed along to `parse.recurse()` in [lib/parse.js](https://github.com/mixu/amdetective/blob/master/lib/parse.js#L196). This is normally the build config options if it is passed.

# License

BSD
