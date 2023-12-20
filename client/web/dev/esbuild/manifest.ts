import path from 'path'

import type * as esbuild from 'esbuild'

export interface WebBuildManifest {
    /** Main JS bundle. */
    'main.js': string

    /** Main CSS bundle. */
    'main.css': string

    /** Embed JS bundle. */
    'embed.js': string

    /** Embed CSS bundle. */
    'embed.css': string
}

export const assetPathPrefix = '/.assets'

export const WEB_BUILD_MANIFEST_FILENAME = 'web.manifest.json'

/**
 * Create a web manifest from esbuild build results.
 *
 *
 * @param buildOptions The esbuild options
 * @param outputs The esbuild metafile outputs
 * @param inBazel Whether the build is running in Bazel
 */
export function createManifestFromBuildResult(
    buildOptions: { entryPoints: string[]; outdir: string },
    outputs: esbuild.Metafile['outputs'],
    inBazel = Boolean(process.env.BAZEL_BINDIR)
): WebBuildManifest {
    const outdir = path.relative(process.cwd(), buildOptions.outdir)

    const assetPath = (filePath: string): string => {
        if (inBazel) {
            const BAZEL_PATH_PREFIX = /^client\/web\/(app_)?bundle\//
            if (!BAZEL_PATH_PREFIX.test(filePath)) {
                throw new Error(`expected filePath to match ${BAZEL_PATH_PREFIX}, got ${filePath}`)
            }
            return filePath.replace(BAZEL_PATH_PREFIX, '')
        }
        return path.relative(outdir, filePath)
    }

    if (buildOptions.entryPoints.length !== 2) {
        throw new Error('expected 2 entryPoints (main and embed)')
    }
    const [mainEntrypoint, embedEntrypoint] = buildOptions.entryPoints.map(filePath =>
        path.relative(process.cwd(), filePath)
    )

    const manifest: Partial<WebBuildManifest> = {}

    // Find the entrypoint in the output files
    for (const [asset, output] of Object.entries(outputs)) {
        if (!output.entryPoint) {
            continue
        }
        if (output.entryPoint.endsWith(mainEntrypoint)) {
            manifest['main.js'] = assetPath(asset)
            if (output.cssBundle) {
                manifest['main.css'] = assetPath(output.cssBundle)
            }
        } else if (output.entryPoint.endsWith(embedEntrypoint)) {
            manifest['embed.js'] = assetPath(asset)
            if (output.cssBundle) {
                manifest['embed.css'] = assetPath(output.cssBundle)
            }
        }
    }

    if (!manifest['main.js']) {
        throw new Error('no main.js found in outputs')
    }
    if (!manifest['main.css']) {
        throw new Error('no main.css found in outputs')
    }
    if (!manifest['embed.js']) {
        throw new Error('no embed.js found in outputs')
    }
    if (!manifest['embed.css']) {
        throw new Error('no embed.css found in outputs')
    }

    return manifest as WebBuildManifest
}
