import fs from 'fs'
import path from 'path'

import type * as esbuild from 'esbuild'

import { STATIC_ASSETS_PATH } from '@sourcegraph/build-config'

import type { WebBuildManifest } from '../utils'

export const assetPathPrefix = '/.assets'

export const getManifest = (
    jsEntrypoint: string,
    cssEntrypoint: string,
    embedEntrypoint?: string
): WebBuildManifest => ({
    'main.js': path.join(assetPathPrefix, jsEntrypoint),
    'main.css': path.join(assetPathPrefix, cssEntrypoint),
    'embed.js': embedEntrypoint ? path.join(assetPathPrefix, embedEntrypoint) : undefined,
    isModule: true,
})

export const WEB_BUILD_MANIFEST_FILENAME = 'web.manifest.json'

/**
 * An esbuild plugin to write a web.manifest.json file.
 */
export const manifestPlugin: esbuild.Plugin = {
    name: 'manifest',
    setup: build => {
        const origMetafile = build.initialOptions.metafile
        build.initialOptions.metafile = true

        build.onEnd(async result => {
            const { entryPoints } = build.initialOptions
            const outputs = result?.metafile?.outputs

            if (!origMetafile) {
                // If we were the only consumers of the metafile, then delete it from the result to
                // avoid unexpected behavior from other downstream consumers relying on the metafile
                // despite not actually enabling it in the config.
                delete result.metafile
            }

            if (!entryPoints) {
                throw new Error('[manifestPlugin] No entrypoints found')
            }
            if (!Array.isArray(entryPoints) || typeof entryPoints[0] !== 'string' || entryPoints.length === 0) {
                throw new Error('[manifestPlugin] Unexpected entryPoints format')
            }
            const mainEntryPoint = entryPoints[0]
            const mainRelativeEntrypoint = path.relative(process.cwd(), mainEntryPoint)

            if (entryPoints[1] && typeof entryPoints[1] !== 'string') {
                throw new Error('[manifestPlugin] Unexpected entryPoints format')
            }
            const embedEntryPoint: string | undefined = entryPoints[1]
            const embedRelativeEntrypoint = embedEntryPoint ? path.relative(process.cwd(), embedEntryPoint) : undefined

            if (!outputs) {
                throw new Error('[manifestPlugin] no outputs')
            }
            let jsEntrypoint: string | undefined
            let cssEntrypoint: string | undefined
            let embedJSEntrypoint: string | undefined

            // Find the entrypoint in the output files
            for (const [asset, output] of Object.entries(outputs)) {
                if (!output.entryPoint) {
                    continue
                }
                if (output.entryPoint.endsWith(mainRelativeEntrypoint)) {
                    jsEntrypoint = path.relative(STATIC_ASSETS_PATH, asset)
                    if (output.cssBundle) {
                        cssEntrypoint = path.relative(STATIC_ASSETS_PATH, output.cssBundle)
                    }
                }
                if (embedRelativeEntrypoint && output.entryPoint.endsWith(embedRelativeEntrypoint)) {
                    embedJSEntrypoint = path.relative(STATIC_ASSETS_PATH, asset)
                }
            }

            if (!jsEntrypoint) {
                throw new Error('[manifestPlugin] Could not find jsEntrypoint in outputs')
            }
            if (!cssEntrypoint) {
                throw new Error('[manifestPlugin] Could not find cssEntrypoint in outputs')
            }

            const { outdir } = build.initialOptions
            if (!outdir) {
                throw new Error('[manifestPlugin] No outdir found')
            }

            const manifest = getManifest(jsEntrypoint, cssEntrypoint, embedJSEntrypoint)
            const manifestPath = path.join(outdir, WEB_BUILD_MANIFEST_FILENAME)
            await fs.promises.writeFile(manifestPath, JSON.stringify(manifest, null, 2))
        })
    },
}
