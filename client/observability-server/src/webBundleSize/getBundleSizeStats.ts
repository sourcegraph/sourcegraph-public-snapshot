/* eslint-disable @typescript-eslint/no-require-imports, import/no-dynamic-require, @typescript-eslint/no-var-requires */
import { statSync } from 'fs'
import path from 'path'

import glob from 'glob'

interface BundleSizeConfig {
    files: {
        path: string
        maxSize: string
    }[]
}

interface BundleSizeStats {
    [baseFilePath: string]: {
        raw: number
        isInitial: boolean
        isDynamicImport: boolean
        isCss: boolean
        isJs: boolean
    }
}

interface GetBundleSizeStatsOptions {
    staticAssetsPath: string
    bundlesizeConfigPath: string
    webBuildManifestPath: string
}

/**
 * Get a list of files specified by bundlesize config glob and their respective sizes.
 */
export function getBundleSizeStats(options: GetBundleSizeStatsOptions): BundleSizeStats {
    const { bundlesizeConfigPath, webBuildManifestPath, staticAssetsPath } = options
    const bundleSizeConfig = require(bundlesizeConfigPath) as BundleSizeConfig
    const webBuildManifest = require(webBuildManifestPath) as Record<string, string>

    const initialResources = new Set(
        Object.values(webBuildManifest)
            .filter(value => typeof value === 'string')
            .map(resourcePath => path.join(staticAssetsPath, resourcePath.replace('/.assets/', '')))
    )

    return bundleSizeConfig.files.reduce<BundleSizeStats>((result, file) => {
        const filePaths = glob.sync(file.path)
        const fileStats = filePaths.reduce((fileStats, noCompressionFilePath) => {
            const name = path.basename(noCompressionFilePath)
            return {
                ...fileStats,
                [noCompressionFilePath.replace(`${staticAssetsPath}/`, '')]: {
                    raw: statSync(noCompressionFilePath).size,
                    isInitial: initialResources.has(noCompressionFilePath),
                    isDynamicImport: name.startsWith('chunk-'),
                    isCss: path.parse(name).ext === '.css',
                    isJs: path.parse(name).ext === '.js',
                },
            }
        }, {})

        return {
            ...result,
            ...fileStats,
        }
    }, {})
}
