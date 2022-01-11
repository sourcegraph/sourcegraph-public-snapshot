import path from 'path'

import { RuleSetRule } from 'webpack'

import { ROOT_PATH } from '../paths'

export const getBabelLoader = (): RuleSetRule => ({
    loader: 'babel-loader',
    options: {
        cacheDirectory: true,
        configFile: path.join(ROOT_PATH, 'babel.config.js'),
    },
})
