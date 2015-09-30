# Sass loader for [webpack](http://webpack.github.io/)

## Install

`npm install sass-loader --save-dev`

Starting with `1.0.0`, the sass-loader requires [node-sass](https://github.com/sass/node-sass) as [`peerDependency`](https://docs.npmjs.com/files/package.json#peerdependencies). Thus you are able to specify the required version accurately.

---

## Usage

[Documentation: Using loaders](http://webpack.github.io/docs/using-loaders.html)

``` javascript
var css = require('!raw!sass!./file.scss');
// => returns compiled css code from file.scss, resolves imports
var css = require('!css!sass!./file.scss');
// => returns compiled css code from file.scss, resolves imports and url(...)s
```

Use in tandem with the [`style-loader`](https://github.com/webpack/style-loader) to add the css rules to your document:

``` javascript
require('!style!css!sass!./file.scss');
```

### Apply via webpack config

It's recommended to adjust your `webpack.config` so `style!css!sass!` is applied automatically on all files ending on `.scss`:

``` javascript
module.exports = {
  module: {
    loaders: [
      {
        test: /\.scss$/,
        loader: 'style!css!sass'
      }
    ]
  }
};
```

Then you only need to write: `require('./file.scss')`.

### Sass options

You can pass any Sass specific configuration options through to the render function via [query parameters](http://webpack.github.io/docs/using-loaders.html#query-parameters).

``` javascript
module.exports = {
  module: {
    loaders: [
      {
        test: /\.scss$/,
        loader: "style!css!sass?outputStyle=expanded&" +
          "includePaths[]=" +
            encodeURIComponent(path.resolve(__dirname, "./some-folder")) + "&" +
          "includePaths[]=" +
            encodeURIComponent(path.resolve(__dirname, "./another-folder"))
      }
    ]
  }
};
```

See [node-sass](https://github.com/andrew/node-sass) for all available options.

### Imports

webpack provides an [advanced mechanism to resolve files](http://webpack.github.io/docs/resolving.html). The sass-loader uses node-sass' custom importer feature to pass all queries to the webpack resolving engine. Thus you can import your sass-modules from `node_modules`. Just prepend them with a `~` which tells webpack to look-up the [`modulesDirectories`](http://webpack.github.io/docs/configuration.html#resolve-modulesdirectories).

```css
@import "~bootstrap/less/bootstrap";
```

It's important to only prepend it with `~`, because `~/` resolves to the home-directory. webpack needs to distinguish between `bootstrap` and `~bootstrap` because CSS- and Sass-files have no special syntax for importing relative files. Writing `@import "file"` is the same as `@import "./file";`

### .sass files

For requiring `.sass` files, add `indentedSyntax` as a loader option:

``` javascript
module.exports = {
  module: {
    loaders: [
      {
        test: /\.sass$/,
        // Passing indentedSyntax query param to node-sass
        loader: "style!css!sass?indentedSyntax"
      }
    ]
  }
};
```

### Problems with `url(...)`

Since Sass/[libsass](https://github.com/sass/libsass) does not provide [url rewriting](https://github.com/sass/libsass/issues/532), all linked assets must be relative to the output.

- If you're just generating CSS without passing it to the css-loader, it must be relative to your web root.
- If you pass the generated CSS on to the css-loader, all urls must be relative to the entry-file (e.g. `main.scss`).

Library authors usually provide a variable to modify the asset path. [bootstrap-sass](https://github.com/twbs/bootstrap-sass) for example has an `$icon-font-path`. Check out [this example](https://github.com/jtangelder/sass-loader/tree/master/test/bootstrapSass).

## Source maps

Because of browser limitations, source maps are only available in conjunction with the [extract-text-webpack-plugin](https://github.com/webpack/extract-text-webpack-plugin). Use that plugin to extract the CSS code from the generated JS bundle into a separate file (which even improves the perceived performance because JS and CSS are downloaded in parallel).

Then your `webpack.config.js` should look like this:

```javascript
var ExtractTextPlugin = require('extract-text-webpack-plugin');

module.exports = {
    ...
    // must be 'source-map' or 'inline-source-map'
    devtool: 'source-map',
    module: {
        loaders: [
            {
                test: /\.scss$/,
                loader: ExtractTextPlugin.extract(
                    // activate source maps via loader query
                    'css?sourceMap!' +
                    'sass?sourceMap'
                )
            }
        ]
    },
    plugins: [
        // extract inline css into separate 'styles.css'
        new ExtractTextPlugin('styles.css')
    ]
};
```

If you want to edit the original Sass files inside Chrome, [there's a good blog post](https://medium.com/@toolmantim/getting-started-with-css-sourcemaps-and-in-browser-sass-editing-b4daab987fb0). Checkout [test/sourceMap](https://github.com/jtangelder/sass-loader/tree/master/test) for a running example. Make sure to serve the content with an HTTP server.

## License

MIT (http://www.opensource.org/licenses/mit-license.php)
