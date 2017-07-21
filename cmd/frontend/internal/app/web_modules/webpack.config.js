var path = require('path');
var webpack = require('webpack');

var plugins;
if (process.env.NODE_ENV === 'production') {
    plugins = [
        new webpack.optimize.UglifyJsPlugin({
            sourceMap: true,
            compressor: {
                warnings: false
            }
        }),
    ];
} else {
    plugins = [
        new webpack.NoEmitOnErrorsPlugin(),
    ]
}

var devtool = process.env.NODE_ENV === 'production' ? undefined : 'cheap-module-source-map';

module.exports = {
    entry: {
        app: path.join(__dirname, 'app.tsx'),
        highlighter: path.join(__dirname, 'highlighter.tsx'),
    },
    output: {
        path: path.join(__dirname, '../../../../../ui/assets/scripts'),
        filename: '[name].bundle.js',
        chunkFilename: '[id].chunk.js'
    },
    devtool: devtool,
    plugins: plugins,
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
        alias: {
            app: path.resolve(__dirname),
        }
    },
    module: {
        loaders: [{
            test: /\.tsx?$/,
            loader: 'ts-loader?' + JSON.stringify({
                compilerOptions: {
                    noEmit: false, // tsconfig.json sets this to true to avoid output when running tsc manually
                },
                transpileOnly: true, // type checking is only done as part of linting or testing
            }),
        }]
    }
};
