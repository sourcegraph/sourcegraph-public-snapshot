# app/script

## Deprecation warning

This code is being migrated into a better architecture at ``../web_modules/sourcegraph`. _Do not write new code here!_

## Testing

To run all the tests for this code:

```
cd sourcegraph/app
npm test
```

To run a specific test:

```
cd sourcegraph/app
NODE_PATH=web_modules ./node_modules/.bin/mocha --compilers js:babel-register --require babel-polyfill ./script/path/to/The_Test.js
```
