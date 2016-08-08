# Contributer's Guide

We welcome contributions from the community and have gathered guidelines 
here to help you get started.

## Discussion

While not absolutely required, it is encouraged that you first open an issue 
for any bug or feature request. This allows discussion on the proper course of
action to take before coding begins.

## Building

```shell
npm install
npm run build
```

## Changing

Most of the information you need to contribute code changes can [be found here](https://guides.github.com/activities/contributing-to-open-source/).
In short: fork, branch, make your changes, and submit a pull request.

## Testing

This project makes use of a test suite both to make sure we don't break
anything and also to make sure new versions of webpack and TypeScript don't
break anything. The tests are full integration tests, meaning each test is a
mini project that is run through webpack and then the output is compared to the
expected output. Not all bugs/features necessarily fit into this framework and
that's OK. However, most do and therefore you should make every effort to
create at least one test which demonstrates the issue or exercises the feature.

The test harness uses certain conventions. All tests have their own directory
under `/test`, eg `/test/someFeature`. Each test should have a
`webpack.config.js` file which follows this general convention:

```javascript
module.exports = {
    entry: './app.ts',
    output: {
        filename: 'bundle.js'
    },
    resolve: {
        extensions: ['', '.ts', 'tsx', '.js']
    },
    module: {
        loaders: [
            { test: /\.tsx?$/, loader: 'ts-loader' }
        ]
    }
}

// for test harness purposes only, you would not need this in a normal project
module.exports.resolveLoader = { alias: { 'ts-loader': require('path').join(__dirname, "../../index.js") } }
```

You can run tests with `npm test`. You can also go into an individual test
directory and manually build a project using `webpack` or `webpack --watch`.
This can be useful both when developing the test and also when fixing an issue
or adding a feature.

Each test should have an `expectedOutput` directory which contains any webpack
filesystem output (typically `bundle.js` and possibly `bundle.js.map`) and any 
console output. stdout should go in `output.txt` and stderr should go in
`err.txt`.

As a convenience it is possible to regenerate the expected output from the 
actual output. This is useful when creating new tests and also when making a
change that affects multiple existing tests. To run use 
`npm test -- --save-output`. Note that all tests will automatically pass when
using this feature. You should double check the generated files to make sure
the output is indeed correct.

The test harness additionally supports watch mode since that is such an
integral part of webpack. The initial state is as described above. After the
initial state is compiled, a series of "patches" can be applied and tested. The
patches use the convention of `/patchN` starting with 0. For example:

Initial state:
- test/someFeature/app.ts
- test/someFeature/expectedOutput/bundle.js
- test/someFeature/expectedOutput/output.txt

Patch 0
- test/someFeature/patch0/app.ts - *modified file*
- test/someFeature/expectedOutput/patch0/bundle.js - *bundle after applying patch*
- test/someFeature/expectedOutput/patch0/output.txt - *output after applying patch*