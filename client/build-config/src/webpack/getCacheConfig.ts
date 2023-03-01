import path from 'path'

import webpack from 'webpack'

import { ROOT_PATH } from '../paths'

// TODO(bazel): drop when non-bazel removed.
const IS_BAZEL = !!process.env.BAZEL_BINDIR

interface CacheConfigOptions {
    invalidateCacheFiles?: string[]
}

export const getCacheConfig = ({ invalidateCacheFiles = [] }: CacheConfigOptions): webpack.Configuration['cache'] => ({
    type: 'filesystem',
    buildDependencies: {
        // Invalidate cache on config change.
        config: IS_BAZEL
            ? invalidateCacheFiles
            : [
                  ...invalidateCacheFiles,
                  path.resolve(ROOT_PATH, 'babel.config.js'),
                  path.resolve(ROOT_PATH, 'postcss.config.js'),
              ],
    },
})
