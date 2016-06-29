Flow Status Webpack Plugin
==========================

[![npm version](https://img.shields.io/npm/v/flow-status-webpack-plugin.svg?style=flat-square)](https://www.npmjs.com/package/flow-status-webpack-plugin) [![npm downloads](https://img.shields.io/npm/dm/flow-status-webpack-plugin.svg?style=flat-square)](https://www.npmjs.com/package/flow-status-webpack-plugin)

This [webpack](http://webpack.github.io/) plugin will automatically start a [Flow](http://flowtype.org/) server (or restart if one is running) when webpack starts up, and run `flow status` after each webpack build. Still experimental.

If you have any idea on how to get it better, you're welcome to contribute!

Requirements
------------

You need to have Flow installed. To do that, follow [these steps](http://flowtype.org/docs/getting-started.html#_).

Installation
------------
`npm install flow-status-webpack-plugin --save-dev`

Usage
-----

```js
var FlowStatusWebpackPlugin = require('flow-status-webpack-plugin');

module.exports = {
    ...
    plugins: [
        new FlowStatusWebpackPlugin()
    ]
}
```

It will generate an output like this:

![Flow has no errors](http://i.imgur.com/GX2xg8J.png?1)

or, in case of some error:

![Flow has errors](http://i.imgur.com/4cnu50c.png?1)

Configuration
-------------

If you want to pass additional command-line arguments to `flow start`, you can pass a `flowArgs` option to the plugin:

```js
var FlowStatusWebpackPlugin = require('flow-status-webpack-plugin');

module.exports = {
    ...
    plugins: [
        new FlowStatusWebpackPlugin({
            flowArgs: '--lib path/to/interfaces/directory'
        })
    ]
}
```

If you don't want the plugin to automatically restart any running Flow server, pass `restartFlow: false`:

```js
var FlowStatusWebpackPlugin = require('flow-status-webpack-plugin');

module.exports = {
    ...
    plugins: [
        new FlowStatusWebpackPlugin({
            restartFlow: false
        })
    ]
}
```

If provided a binary path, will run Flow from this path instead of running it from any global installation.

```js
var FlowStatusWebpackPlugin = require('flow-status-webpack-plugin');

module.exports = {
    ...
    plugins: [
        new FlowStatusWebpackPlugin({
            binaryPath: '/path/to/your/flow/installation'
        })
    ]
}
```

If you don't want the plugin to display a message on success, pass
`quietSuccess: true`:

```js
var FlowStatusWebpackPlugin = require('flow-status-webpack-plugin');

module.exports = {
    ...
    plugins: [
        new FlowStatusWebpackPlugin({
            quietSuccess: true
        })
    ]
}
```

License
-------
This plugin is released under the [MIT License](https://opensource.org/licenses/MIT).
