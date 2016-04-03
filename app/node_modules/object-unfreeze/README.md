# object-unfreeze

[![Travis build status](http://img.shields.io/travis/gajus/object-unfreeze/master.svg?style=flat-square)](https://travis-ci.org/gajus/object-unfreeze)
[![NPM version](http://img.shields.io/npm/v/object-unfreeze.svg?style=flat-square)](https://www.npmjs.org/package/object-unfreeze)

Unfreeze an object frozen with Object.freeze.

## Usage

```js
import objectUnfreeze from 'object-unfreeze';

let subject,
    shallowCopy;

subject = {};

Object.freeze(subject);

// Throws an error.
// subject.foo = 'FOO';

shallowCopy = objectUnfreeze(subject);

shallowCopy.foo = 'FOO';
```

## Download

Download using NPM:

```sh
npm install object-unfreeze
```
