# tslint loader for webpack

## Usage

Apply the tslint loader as pre/postLoader in your webpack configuration:

``` javascript
module.exports = {
    module: {
        preLoaders: [
            {
                test: /\.ts$/,
                loader: "tslint"
            }
        ]
    },
    // more options in the optional tslint object
    tslint: {
        configuration: {
            rules: {
                quotemark: [true, "double"]
            }
        },

        // tslint errors are displayed by default as warnings
        // set emitErrors to true to display them as errors
        emitErrors: false,

        // tslint does not interrupt the compilation by default
        // if you want any file with tslint errors to fail
        // set failOnHint to true
        failOnHint: true,        

        // name of your formatter (optional)
        formatter: "yourformatter",

        // path to directory containing formatter (optional)
        formattersDirectory: "node_modules/tslint-loader/formatters/",
        
        // These options are useful if you want to save output to files
        // for your continuous integration server
        fileOutput: {
            // The directory where each file's report is saved
            dir: "./foo/",

            // The extension to use for each report's filename. Defaults to "txt"
            ext: "xml",

            // If true, all files are removed from the report directory at the beginning of run
            clean: true,

            // A string to include at the top of every report file.
            // Useful for some report formats.
            header: "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n<checkstyle version=\"5.7\">",

            // A string to include at the bottom of every report file.
            // Useful for some report formats.
            footer: "</checkstyle>"
        }
    }
}
```
## Installation

``` shell
npm install tslint-loader --save-dev
```

## License

MIT (http://www.opensource.org/licenses/mit-license.php)


