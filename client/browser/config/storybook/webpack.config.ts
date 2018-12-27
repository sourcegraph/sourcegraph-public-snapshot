import * as webpack from 'webpack'
import { commonStylesheetLoaders, jsRule, tsRule } from '../shared/webpack'

const buildWebpackConfig = (
    baseConfig: webpack.Configuration,
    env: any,
    config: webpack.Configuration
): webpack.Configuration => {
    config.module!.rules.push(tsRule)
    config.module!.rules.push(jsRule)
    config.resolve!.extensions!.push('.ts', '.tsx')

    // Put our style rules at the beginning so they're processed by the time it
    // gets to storybook's style rules.
    config.module!.rules.unshift({
        test: /\.(css|sass|scss)$/,
        use: ['style-loader', ...commonStylesheetLoaders],
    })

    return config
}

export default buildWebpackConfig
