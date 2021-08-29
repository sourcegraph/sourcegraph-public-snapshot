import * as esbuild from 'esbuild'

import { packageResolutionPlugin } from './packageResolutionPlugin'

async function buildWorker(
    workerPath: string,
    extraConfig: Pick<esbuild.BuildOptions, 'target' | 'format'>
): Promise<string> {
    const results = await esbuild.build({
        entryPoints: [workerPath],
        bundle: true,
        write: false,
        plugins: [packageResolutionPlugin],
        sourcemap: true,
        ...extraConfig,
    })
    return results.outputFiles[0].text
}

export const workerPlugin: esbuild.Plugin = {
    name: 'esbuild-plugin-inline-worker',
    setup: build => {
        build.onLoad({ filter: /\.worker\.ts$/, namespace: 'file' }, async ({ path: workerPath }) => {
            // TODO(sqs): memoize this, and return the metafile deps as the watchDir/watchFiles
            const t0 = Date.now()
            const workerCode = await buildWorker(workerPath, {
                target: build.initialOptions.target,
                format: build.initialOptions.format,
            })
            console.log('Worker took', Date.now() - t0)

            return {
                contents: `import inlineWorker from '__inline-worker'
export default function Worker() {
  return inlineWorker(${JSON.stringify(workerCode)})
}
`,
                loader: 'js',
            }
        })

        build.onResolve({ filter: /^__inline-worker$/ }, ({ path }) => ({ path, namespace: 'inline-worker' }))
        build.onLoad({ filter: /.*/, namespace: 'inline-worker' }, () => ({
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
