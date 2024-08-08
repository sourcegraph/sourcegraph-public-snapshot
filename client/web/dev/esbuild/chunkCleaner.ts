import fs from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'

/**
 * chunkCleaner is a small plugin that removes the 'chunks' directory when the build starts from the configured `outDir` if it exists
 **/
const chunkCleaner: esbuild.Plugin = {
    name: 'chunkCleaner',
    setup(build: esbuild.PluginBuild) {
        build.onStart(() => {
            const outDir = build.initialOptions.outdir
            if (!outDir) {
                // nothing to clean if no outdir is defined
                return
            }

            const chunkDir = path.join(outDir, 'chunks')

            console.debug(`[chunk cleaner] removing '${chunkDir}'`)
            fs.rm(chunkDir, { recursive: true, force: true }, err => {
                if (err) {
                    console.warn(`[chunk cleaner] failed to remove ${chunkDir} - it might not exist: ${err}`)
                } else {
                    console.log(`[chunk cleaner] removed '${chunkDir}'`)
                }
            })
        })
    },
}

export default chunkCleaner
