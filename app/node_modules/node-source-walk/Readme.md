### node-source-walk [![npm](http://img.shields.io/npm/v/node-source-walk.svg)](https://npmjs.org/package/node-source-walk) [![npm](http://img.shields.io/npm/dm/node-source-walk.svg)](https://npmjs.org/package/node-source-walk)

> Execute a callback on every node of a file's AST and stop walking whenever you see fit.

*A variation of [substack/node-detective](https://github.com/substack/node-detective)
and simplification of [substack/node-falafel](https://github.com/substack/node-falafel).*

`npm install node-source-walk`

### Usage

```javascript
  var Walker = require('node-source-walk');

  var walker = new Walker();

  // Assume src is the string contents of myfile.js
  // or the AST of an outside parse of myfile.js

  walker.walk(src, function (node) {
    if (/* some condition */) {
      // No need to keep traversing since we found what we wanted
      walker.stopWalking();
    }
  });

```

By default, Walker will use `acorn` supporting ES6 and the `sourceType: module`, but you can change any of the defaults as follows:

```js
var walker = new Walker({
  ecmaVersion: 5,
  sourceType: 'script'
});
```

* The supplied options are passed through to the parser, so you can configure it according
to acorn's documentation: https://github.com/marijnh/acorn

### Public Members

`walk(src, cb)`

* src: the contents of a file OR its AST (via Esprima or Acorn)
* cb: a function that is called for every visited node

`stopWalking()`

* Halts further walking of the AST until another manual call of `walk`.
* This is super-beneficial when dealing with large source files

`traverse(node, cb)`

* Allows you to traverse an AST node and execute a callback on it
* Callback should expect the first argument to be an AST node, similar to `walk`'s callback.
