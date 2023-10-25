import fs from 'fs'
import path from 'path'

import type * as esbuild from 'esbuild'

import { WEB_BUILD_MANIFEST_FILENAME, createManifestFromBuildResult } from './manifest'

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

            if (!checkEntryPoints(entryPoints)) {
                throw new Error('[manifestPlugin] Unexpected entryPoints format')
            }

            const { outdir } = build.initialOptions
            if (!outdir) {
                throw new Error('[manifestPlugin] No outdir found')
            }

            if (!outputs) {
                throw new Error('[manifestPlugin] No outputs found')
            }

            const manifest = createManifestFromBuildResult({ entryPoints, outdir }, outputs)
            const manifestPath = path.join(outdir, WEB_BUILD_MANIFEST_FILENAME)
            await fs.promises.writeFile(manifestPath, JSON.stringify(manifest, null, 2))
        })
    },
}

function checkEntryPoints(entryPoints: esbuild.BuildOptions['entryPoints']): entryPoints is string[] {
    return Array.isArray(entryPoints) && typeof entryPoints[0] === 'string'
}
