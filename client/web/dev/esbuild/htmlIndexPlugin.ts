import fs from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import * as handlebars from 'handlebars'

import { STATIC_ASSETS_PATH } from '@sourcegraph/build-config'

import { WebpackManifest, HTML_INDEX_PATH } from '../utils'

// Note: This is only valid for Sourcegraph App.
export const assetPathPrefix = '/'

export const getManifest = (jsEntrypoint: string, cssEntrypoint?: string): WebpackManifest => ({
    'shell.js': path.join(assetPathPrefix, jsEntrypoint ?? 'scripts/app.js'),
    'app.css': path.join(assetPathPrefix, cssEntrypoint ?? 'scripts/app.css'),
    isModule: true,
})

const writeHtmlIndex = async (manifest: WebpackManifest): Promise<void> => {
    const template = await fs.promises.readFile('index.html.template', 'utf8')
    const render = handlebars.compile(template)
    const content = render({
        cssBundle: manifest['app.css'],
        jsBundle: manifest['shell.js'],
        isModule: manifest.isModule,
    })
    await fs.promises.writeFile(HTML_INDEX_PATH, content)
}

const ENTRYPOINT_NAME = 'scripts/shell'

/**
 * An esbuild plugin to write a index.html file for Sourcegraph, for compatibility with the current
 * Go backend template system.
 *
 * This is only used in Sourcegraph App, currently.
 */
export const htmlIndexPlugin: esbuild.Plugin = {
    name: 'htmlIndex',
    setup: build => {
        build.initialOptions.metafile = true

        build.onEnd(async result => {
            const { entryPoints } = build.initialOptions
            const outputs = result?.metafile?.outputs

            if (!entryPoints) {
                console.error('[htmlIndexPlugin] No entrypoints found')
                return
            }
            const absoluteEntrypoint: string | undefined = (entryPoints as any)[ENTRYPOINT_NAME]
            if (!absoluteEntrypoint) {
                console.error('[htmlIndexPlugin] No entrypoint found with the name scripts/app')
                return
            }
            const relativeEntrypoint = path.relative(process.cwd(), absoluteEntrypoint)

            if (!outputs) {
                return
            }
            let jsEntrypoint: string | undefined
            let cssEntrypoint: string | undefined

            // Find the entrypoint in the output files
            for (const [asset, output] of Object.entries(outputs)) {
                if (output.entryPoint === relativeEntrypoint) {
                    jsEntrypoint = path.relative(STATIC_ASSETS_PATH, asset)
                    if (output.cssBundle) {
                        cssEntrypoint = path.relative(STATIC_ASSETS_PATH, output.cssBundle)
                    }
                }
            }

            await writeHtmlIndex(getManifest(jsEntrypoint, cssEntrypoint))
        })
    },
}
