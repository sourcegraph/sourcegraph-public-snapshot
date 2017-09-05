import ExtractTextPlugin = require('extract-text-webpack-plugin');
import * as path from 'path';
import Tapable = require('tapable');
import TSLintPlugin = require('tslint-webpack-plugin');
import * as webpack from 'webpack';

const plugins: webpack.Plugin[] = [
    // Print some output for VS Code tasks to know when a build started
    function(this: Tapable): void {
        this.plugin('watch-run', (watching: any, cb: () => void) => {
            console.log('Begin compile at ' + new Date());
            cb();
        });
    }
];

if (process.env.NODE_ENV === 'production') {
    plugins.push(
        new webpack.DefinePlugin({
            'process.env': {
                NODE_ENV: JSON.stringify('production')
            }
        }),
        new webpack.optimize.UglifyJsPlugin({
            sourceMap: false
        })
    );
} else {
    plugins.push(
        new webpack.DefinePlugin({
            'process.env': {
                NODE_ENV: JSON.stringify('development')
            }
        }),
        new webpack.NoEmitOnErrorsPlugin(),
        new TSLintPlugin({
            files: ['**/*.tsx'],
            exclude: ['node_modules/**']
        })
    );
}

plugins.push(new ExtractTextPlugin({
    filename: 'ui/assets/dist/[name].bundle.css',
    allChunks: true
}));

const devtool = process.env.NODE_ENV === 'production' ? undefined : 'cheap-module-eval-source-map';

const config: webpack.Configuration = {
    entry: {
        app: path.join(__dirname, 'src/app.tsx'),
        style: path.join(__dirname, 'src/app.scss')
    },
    output: {
        path: path.join(__dirname, '../ui/assets/scripts'),
        filename: '[name].bundle.js',
        chunkFilename: '[id].chunk.js'
    },
    devtool,
    plugins,
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
        alias: {
            sourcegraph: path.resolve(__dirname, 'src')
        }
    },
    module: {
        loaders: [{
            test: /\.tsx?$/,
            loader: 'ts-loader?' + JSON.stringify({
                compilerOptions: {
                    noEmit: false // tsconfig.json sets this to true to avoid output when running tsc manually
                },
                transpileOnly: process.env.DISABLE_TYPECHECKING === 'true'
            })
        }, {
            // sass / scss loader for webpack
            test: /\.(css|sass|scss)$/,
            loader: ExtractTextPlugin.extract(['css-loader', 'postcss-loader', 'sass-loader'])
        }]
    }
};

export default config;
