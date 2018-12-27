import UglifyJsPlugin from 'uglifyjs-webpack-plugin'
import * as webpack from 'webpack'
import baseConfig from './base.config'
import { generateBundleUID } from './utils'

const { plugins, ...base } = baseConfig

const config: webpack.Configuration = {
    ...base,
    mode: 'production',
    optimization: {
        minimize: true,
        minimizer: [
            new UglifyJsPlugin({
                uglifyOptions: {
                    output: {
                        // Without this, Uglify will change \u0000 to \0 (NULL byte),
                        // which causes Chrome to complain that the bundle is not UTF8
                        ascii_only: true,
                        beautify: false,
                    },
                },
            }),
        ],
    },
    plugins: (plugins || []).concat(
        ...[
            new webpack.DefinePlugin({
                'process.env': {
                    NODE_ENV: JSON.stringify('production'),
                    BUNDLE_UID: JSON.stringify(generateBundleUID()),
                    USE_EXTENSIONS: JSON.stringify(process.env.USE_EXTENSIONS),
                },
            }),
            new webpack.ProvidePlugin({
                // tslint:disable-next-line object-literal-key-quotes
                $: 'jquery',
                // tslint:disable-next-line object-literal-key-quotes
                jQuery: 'jquery',
                '$.fn.pjax': 'jquery-pjax',
            }),
        ]
    ),
}
export default config
