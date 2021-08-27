import fs from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'

import { uiAssetsPath } from './build'

export const assetPathPrefix = '/.assets/'

interface Manifest {
    'app.js': string
    'app.css': string
    isModule: boolean
}

const writeManifest = async (manifest: Manifest): Promise<void> => {
    const manifestPath = path.join(uiAssetsPath, 'webpack.manifest.json')
    await fs.promises.writeFile(manifestPath, JSON.stringify(manifest, null, 2))
}

export const manifestPlugin: esbuild.Plugin = {
    name: 'manifest',
    setup: build => {
        build.onStart(async () => {
            // TODO(sqs): bug https://github.com/evanw/esbuild/issues/1384 means that onEnd isn't
            // called in serve mode, so write it here because we know what it should be. When that
            // is fixed, add this to an onEnd hook when we know the exact filenames without
            // hard-coding.
            await writeManifest({
                'app.js': path.join(assetPathPrefix, 'app.js'),
                'app.css': path.join(assetPathPrefix, 'app.css'),
                isModule: true,
            })
        })
    },
}
