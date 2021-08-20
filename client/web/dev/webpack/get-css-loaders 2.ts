import path from 'path'

import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import { RuleSetUseItem } from 'webpack'

/**
 * Generates array of CSS loaders both for regular CSS and CSS modules.
 * Useful to ensure that we use the same configuration for shared loaders: postcss-loader, sass-loader, etc.
 * */
export const getCSSLoaders = (
    rootPath: string,
    isDevelopment: boolean,
    ...loaders: RuleSetUseItem[]
): RuleSetUseItem[] => [
    // Use style-loader for local development as it is significantly faster.
    isDevelopment ? 'style-loader' : MiniCssExtractPlugin.loader,
    ...loaders,
    'postcss-loader',
    {
        loader: 'sass-loader',
        options: {
            sassOptions: {
                // eslint-disable-next-line @typescript-eslint/no-require-imports
                implementation: require('sass'),
                includePaths: [path.resolve(rootPath, 'node_modules'), path.resolve(rootPath, 'client')],
            },
        },
    },
]
