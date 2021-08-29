import fs from 'fs'
import os from 'os'
import path from 'path'

import * as esbuild from 'esbuild'

import { newEsbuildPluginCache } from './pluginCache'

// TODO(sqs): this was found to NOT be faster than not using it

/**
 * An esbuild plugin that builds Monaco as a single entrypoint for the main thread code instead of
 * multiple entrypoints, to reduce the number of chunks and simplify the build.
 */
export const monacoPlugin: esbuild.Plugin = {
    name: 'monaco',
    setup: async build => {
        /// ////// const cache = newEsbuildPluginCache('esbuild-monaco-plugin')

        // eslint-disable-next-line no-sync
        const temporaryDirectoryPath = fs.mkdtempSync(path.join(os.tmpdir(), 'esbuild-plugin-monaco'))
        process.addListener('exit', () => {
            // eslint-disable-next-line no-sync
            fs.rmdirSync(temporaryDirectoryPath, { recursive: true })
        })

        const monacoEditorEntrypoint = path.join(
            __dirname,
            '../../../../node_modules/monaco-editor/esm/vs/editor/editor.main.js'
        )
        const t0 = Date.now()
        const incrementalBuild = await esbuild.build({
            entryPoints: {
                'monaco-editor.bundle': monacoEditorEntrypoint,
            },
            incremental: true,
            bundle: true,
            outdir: temporaryDirectoryPath,
            target: build.initialOptions.target,
            format: build.initialOptions.format,
            loader: { '.ttf': 'base64' },
        })
        console.log('Initial build took', Date.now() - t0)

        const bundlePath = (extension: 'js' | 'css'): string =>
            path.join(temporaryDirectoryPath, `monaco-editor.bundle.${extension}`)
        const TMPFILE = '/tmp/x.js'
        await fs.promises.writeFile(
            TMPFILE,
            `
        import ${JSON.stringify(bundlePath('css'))}
        export * from ${JSON.stringify(bundlePath('js'))}
    `
        )

        build.onResolve({ filter: /^monaco-editor($|\/)/, namespace: 'file' }, ({ path, importer }) => {
            if (path !== 'monaco-editor') {
                throw new Error(
                    `Import "monaco-editor" instead of ${JSON.stringify(
                        path
                    )} in ${importer}. Importing a submodule of "monaco-editor" is not allowed because a single bundle is created for all of monaco-editor (for faster builds and page loads).`
                )
            }

            /* return {
                path: 'main',
                namespace: 'monaco-editor',
            } */
            return { path: TMPFILE }
        })

        if (false) {
            build.onLoad({ filter: /^main$/, namespace: 'monaco-editor' }, async () => {
                const t0 = Date.now()
                // await incrementalBuild.rebuild()
                console.log('Rebuild took', Date.now() - t0)

                return {
                    contents: `
                    import ${JSON.stringify(bundlePath('css'))}
                    export * from ${JSON.stringify(bundlePath('js'))}
                `,
                    resolveDir: __dirname,
                    loader: 'js',
                }
            })
        }
    },
}
