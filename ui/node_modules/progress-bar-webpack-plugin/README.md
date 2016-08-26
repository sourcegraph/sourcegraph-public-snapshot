# progress-bar-webpack-plugin
![progress-bar-webpack-plugin](http://i.imgur.com/OIP1gnj.gif)

## Installation

```
npm i -D progress-bar-webpack-plugin
```

## Usage

Include the following in your Webpack config.

```javascript
var ProgressBarPlugin = require('progress-bar-webpack-plugin');

...

plugins: [
  new ProgressBarPlugin()
]
```

## Options

Accepts almost all of the same options as [node-progress](https://github.com/tj/node-progress#options).

- `format` the format of the progress bar
- `width` the displayed width of the progress bar defaulting to total
- `complete` completion character defaulting to "="
- `incomplete` incomplete character defaulting to " "
- `renderThrottle` minimum time between updates in milliseconds defaulting to 16
- `clear` option to clear the bar on completion defaulting to true
- `callback` optional function to call when the progress bar completes
- `stream` the output stream defaulting to stderr
- `summary` option to show summary of time taken defaulting to true
- `summaryContent` optional custom summary message if summary option is false
- `customSummary` optional function to display a custom summary (passed build time)

The `format` option accepts the following tokens:

- `:bar` the progress bar itself
- `:current` current tick number
- `:total` total ticks
- `:elapsed` time elapsed in seconds
- `:percent` completion percentage
- `:msg` current progress message

The default format uses the `:bar` and `:percent` tokens.

Use [chalk](https://github.com/chalk/chalk) to sprinkle on a few colors.

To include the time elapsed and prevent the progress bar from being cleared on build completion:

```javascript
new ProgressBarPlugin({
  format: '  build [:bar] ' + chalk.green.bold(':percent') + ' (:elapsed seconds)',
  clear: false
})
```

## License

MIT
