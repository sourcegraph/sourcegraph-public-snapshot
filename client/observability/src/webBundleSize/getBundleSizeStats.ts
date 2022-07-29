import { statSync } from 'fs'
import path from 'path'

import glob from 'glob'

import { STATIC_ASSETS_PATH } from '@sourcegraph/build-config'

interface BundleSizeConfig {
    files: {
        path: string
        maxSize: string
    }[]
}

interface BundleSizeStats {
    [baseFilePath: string]: {
        raw: number
        gzip: number
        brotli: number
    }
}

/**
 * Get a list of files specified by bundlesize config glob and their respective sizes.
 */
export function getBundleSizeStats(bundlesizeConfigPath: string): BundleSizeStats {
    // eslint-disable-next-line import/no-dynamic-require, @typescript-eslint/no-var-requires, @typescript-eslint/no-require-imports
    const bundleSizeConfig = require(bundlesizeConfigPath) as BundleSizeConfig

    return bundleSizeConfig.files.reduce<BundleSizeStats>((result, file) => {
        const filePaths = glob.sync(file.path)

        const fileStats = filePaths.reduce((fileStats, brotliFilePath) => {
            const { dir, name } = path.parse(brotliFilePath)
            const noCompressionFilePath = path.join(dir, name)
            const gzipFilePath = `${noCompressionFilePath}.gz`

            return {
                ...fileStats,
                [noCompressionFilePath.replace(`${STATIC_ASSETS_PATH}/`, '')]: {
                    raw: statSync(noCompressionFilePath).size,
                    gzip: statSync(gzipFilePath).size,
                    brotli: statSync(brotliFilePath).size,
                },
            }
        }, {})

        return {
            ...result,
            ...fileStats,
        }
    }, {})
}
