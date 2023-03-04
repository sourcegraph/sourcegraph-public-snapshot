import path from 'path'

import webpack from 'webpack'

import { ROOT_PATH, NODE_MODULES_PATH } from '../paths'

export const getBasicCSSLoader = (): webpack.RuleSetUseItem => ({
    loader: 'css-loader',
    options: { url: false },
})

interface CSSModulesLoaderOptions {
    /**
     * Allow to enable/disable handling the CSS functions `url` and `image-set`.
     * Documentation: https://webpack.js.org/loaders/css-loader/#url.
     */
    url?: boolean
    sourceMap?: boolean
}

export const getCSSModulesLoader = ({ sourceMap, url }: CSSModulesLoaderOptions): webpack.RuleSetUseItem => ({
    loader: 'css-loader',
    options: {
        sourceMap,
        modules: {
            exportLocalsConvention: 'camelCase',
            localIdentName: '[name]__[local]_[hash:base64:5]',
        },
        url,
    },
})

/**
 * Generates array of CSS loaders both for regular CSS and CSS modules.
 * Useful to ensure that we use the same configuration for shared loaders: postcss-loader, sass-loader, etc.
 * */
export const getCSSLoaders = (...loaders: webpack.RuleSetUseItem[]): webpack.RuleSetUse => [
    ...loaders,
    'postcss-loader',
    {
        loader: 'sass-loader',
        options: {
            sassOptions: {
                includePaths: [NODE_MODULES_PATH, path.resolve(ROOT_PATH, 'client')],
            },
        },
    },
]

export const getBazelCSSLoaders = (...loaders: webpack.RuleSetUseItem[]): webpack.RuleSetUse => loaders
