import path from 'path'
import webpack from 'webpack'

export default ({ config }: { config: webpack.Configuration }) => {
    if (!config.module || !config.resolve || !config.resolve.extensions) {
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

    config.resolve.alias = {
        ...config.resolve.alias, // HACK: This is required because the codeintellify package has a hardcoded import that assumes that
        // ../node_modules/@sourcegraph/react-loading-spinner is a valid path. This is not a correct assumption
        // in general, and it also breaks in this build because CSS imports URLs are not resolved (we would
        // need to use resolve-url-loader). There are many possible fixes that are more complex, but this hack
        // works fine for now.
        '../node_modules/@sourcegraph/react-loading-spinner/lib/LoadingSpinner.css': require.resolve(
            '@sourcegraph/react-loading-spinner/lib/LoadingSpinner.css'
        ),
    }

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
                    includePaths: [path.resolve(__dirname, '..', 'node_modules')],
                },
            },
        ],
    })

    return config
}
