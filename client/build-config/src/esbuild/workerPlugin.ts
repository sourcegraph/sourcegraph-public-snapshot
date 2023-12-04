import path from 'path'

import * as esbuild from 'esbuild'

import { packageResolutionPlugin, RXJS_RESOLUTIONS } from './packageResolutionPlugin'

async function buildWorker(
    workerPath: string,
    extraConfig: Pick<esbuild.BuildOptions, 'target' | 'format'>
): Promise<{ workerBundle: string } & Pick<esbuild.OnResolveResult, 'watchFiles' | 'errors' | 'warnings'>> {
    const results = await esbuild.build({
        entryPoints: [workerPath],
        bundle: true,
        write: false,
        plugins: [
            packageResolutionPlugin({
                path: require.resolve('path-browserify'),
                ...RXJS_RESOLUTIONS,
            }),
        ],
        sourcemap: true,
        metafile: true,
        ...extraConfig,
    })
    return {
        workerBundle: results.outputFiles[0].text,
        watchFiles: results.metafile && Object.keys(results.metafile.inputs),
        errors: results.errors,
        warnings: results.warnings,
    }
}

/**
 * An esbuild plugin that bundles a Web Worker (classic, not a module worker) given an import of a
 * `.worker.ts` file.
 *
 * TODO(sqs): This could be improved to use the new `new Worker(new URL(..., import.meta.url))`
 * worker syntax that many bundlers are starting to prefer.
 */
export const workerPlugin: esbuild.Plugin = {
    name: 'worker',
    setup: build => {
        build.onResolve(
            { filter: /\.worker\.(js|ts)$/, namespace: 'file' },
            ({ path: workerPath, importer, resolveDir }) => {
                // This plugin is only needed for classic Web Workers (which need to be bundled
                // independently), not module Web Workers (which don't need a separate bundle
                // because they can `import` shared modules just like non-worker code). Assume that
                // workers specified as entrypoints in the build are module workers, and skip them.
                const isEntrypoint = importer === ''
                return isEntrypoint ? undefined : { path: path.join(resolveDir, workerPath), namespace: 'worker' }
            }
        )
        build.onLoad({ filter: /./, namespace: 'worker' }, async ({ path: workerPath }) => {
            const { workerBundle, watchFiles, errors, warnings } = await buildWorker(workerPath, {
                target: build.initialOptions.target,
                format: build.initialOptions.format,
            })

            return {
                contents: `
                    import inlineWorker from '__inline-worker'
                    export default function Worker() {
                        return inlineWorker(${JSON.stringify(workerBundle)})
                    }
`,
                loader: 'js',
                watchFiles,
                errors,
                warnings,
            }
        })

        build.onResolve({ filter: /^__inline-worker$/ }, ({ path }) => ({ path, namespace: 'inline-worker' }))
        build.onLoad({ filter: /./, namespace: 'inline-worker' }, () => ({
            contents: `
                export default function inlineWorker(scriptText) {
                    const blob = new Blob([scriptText], { type: 'text/javascript' })
                    const url = URL.createObjectURL(blob)
                    const worker = new Worker(url)
                    URL.revokeObjectURL(url)
                    return worker
                }`,
            loader: 'js',
        }))
    },
}
