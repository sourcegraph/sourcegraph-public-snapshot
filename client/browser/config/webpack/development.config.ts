import * as path from 'path'
import * as webpack from 'webpack'
import baseConfig from './base.config'
import { generateBundleUID } from './utils'

const config: webpack.Configuration = {
    ...baseConfig,
    entry:
        process.env.AUTO_RELOAD === 'false'
            ? baseConfig.entry
            : {
                  ...baseConfig.entry,
                  background: [
                      path.resolve(__dirname, '../../src/browser-extension/scripts/auto-reloading.ts'),
                      ...baseConfig.entry.background,
                  ],
              },
    mode: 'development',
    plugins: [
        ...baseConfig.plugins,
        new webpack.DefinePlugin({
            'process.env': {
                NODE_ENV: JSON.stringify('development'),
                BUNDLE_UID: JSON.stringify(generateBundleUID()),
                USE_EXTENSIONS: JSON.stringify(process.env.USE_EXTENSIONS),
            },
        }),
    ],
}
export default config
