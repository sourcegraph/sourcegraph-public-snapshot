node-noop
=========
Perfection is achieved, not when there is nothing more to add, but when there is nothing left to take away. 

- Antoine de Saint-Exupery

Installation
------------
```
npm install node-noop
```

Usage
-----
```
var noop = require('node-noop').noop;
require('fs').writeFile('file.out',"Ignore write failure",noop);
```

Alternatives
------------
The npm package `noop` (github
[here](https://github.com/coolaj86/javascript-noop)) already has
this basic functionality, but it doesn't do it in a very node-like way.
He just sticks the noop function on the global scope when you require it.
