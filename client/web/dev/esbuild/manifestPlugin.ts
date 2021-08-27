import fs from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'

import { uiAssetsPath } from './build'

export const assetPathPrefix = '/.assets/'

const writeManifest = async (appJSPath: string): Promise<void> => {
    const manifestPath = path.join(uiAssetsPath, 'webpack.manifest.json')
    await fs.promises.writeFile(manifestPath, JSON.stringify({ 'app.js': path.join(assetPathPrefix, appJSPath) }))
}

export const manifestPlugin: esbuild.Plugin = {
    name: 'manifest',
    setup: build => {
        build.onStart(async () => {
            // TODO(sqs): bug https://github.com/evanw/esbuild/issues/1384 means that onEnd isn't
            // called in serve mode, so write it here because we know what it should be.
            await writeManifest('web/src/main.js')
        })

        // TODO(sqs): bug https://github.com/evanw/esbuild/issues/1384 means that onEnd isn't called
        // in serve mode, so this is never actually called.
        build.onEnd(async result => {
            await writeManifest(result.outputFiles[0].path)
        })
    },
}
