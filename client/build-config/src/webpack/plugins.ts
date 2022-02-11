import TerserPlugin from 'terser-webpack-plugin'
import webpack from 'webpack'

export const getTerserPlugin = (): TerserPlugin =>
    new TerserPlugin({
        terserOptions: {
            compress: {
                // Don't inline functions, which causes name collisions with uglify-es:
                // https://github.com/mishoo/UglifyJS2/issues/2842
                inline: 1,
            },
        },
    })

export const getProvidePlugin = (): webpack.ProvidePlugin =>
    new webpack.ProvidePlugin({
        process: 'process/browser',
        // Based on the issue: https://github.com/webpack/changelog-v5/issues/10
        Buffer: ['buffer', 'Buffer'],
    })
