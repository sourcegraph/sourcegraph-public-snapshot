import path from 'path'
import webpack from 'webpack'

export default ({ config }: { config: webpack.Configuration }) => {
    if (!config.module || !config.resolve?.extensions) {
        throw new Error('unexpected config')
    }

    config.module.rules.push({
        test: /\.tsx?$/,
        loader: require.resolve('babel-loader'),
        options: {
            presets: [['react-app', { flow: false, typescript: true }]],
        },
    })
    config.resolve.extensions.push('.ts', '.tsx')

    // Put our style rules at the beginning so they're processed by the time it
    // gets to storybook's style rules.
    config.module.rules.unshift({
        test: /\.(css|sass|scss)$/,
        use: [
            'style-loader',
            'css-loader',
            {
                loader: 'sass-loader',
                options: {
                    sassOptions: {
                        includePaths: [path.resolve(__dirname, '..', 'node_modules')],
                    },
                },
            },
        ],
        // Make sure Storybook styles get handled by the Storybook config
        exclude: /node_modules\/@storybook\//,
    })

    return config
}
