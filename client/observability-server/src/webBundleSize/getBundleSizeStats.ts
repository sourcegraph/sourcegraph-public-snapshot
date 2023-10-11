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
        gzip: number
        brotli: number
        isInitial: boolean
        isDynamicImport: boolean
        isDefaultVendors: boolean
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
        Object.values(webBuildManifest).map(resourcePath =>
            path.join(staticAssetsPath, resourcePath.replace('/.assets/', ''))
        )
    )

    return bundleSizeConfig.files.reduce<BundleSizeStats>((result, file) => {
        const filePaths = glob.sync(file.path)

        const fileStats = filePaths.reduce((fileStats, brotliFilePath) => {
            const { dir, name } = path.parse(brotliFilePath)
            const noCompressionFilePath = path.join(dir, name)
            const gzipFilePath = `${noCompressionFilePath}.gz`

            return {
                ...fileStats,
                [noCompressionFilePath.replace(`${staticAssetsPath}/`, '')]: {
                    raw: statSync(noCompressionFilePath).size,
                    gzip: statSync(gzipFilePath).size,
                    brotli: statSync(brotliFilePath).size,
                    isInitial: initialResources.has(noCompressionFilePath),
                    isDynamicImport: name.startsWith('sg_'),
                    isDefaultVendors: /\d+.chunk.js/.test(name),
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
