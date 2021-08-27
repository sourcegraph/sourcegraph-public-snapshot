// TODO(sqs): copied from https://sourcegraph.com/github.com/endreymarcell/esbuild-plugin-webworker/-/blob/plugin.js
import path from 'path'

import * as esbuild from 'esbuild'

export const workerPlugin: esbuild.Plugin = {
    name: 'worker',
    setup: build => {
        build.onResolve({ filter: /\.worker\.ts$/ }, args => ({
            path: args.path,
            namespace: 'worker',
            pluginData: { resolveDir: args.resolveDir },
        }))

        build.onLoad({ filter: /./, namespace: 'worker' }, async args => {
            const {
                path: importPath,
                pluginData: { resolveDir },
            } = args

            const workerWithFullPath = path.resolve(resolveDir, importPath)

            const result = await esbuild.build({
                entryPoints: [workerWithFullPath],
                bundle: true,
                write: false,
                target: build.initialOptions.target,
                format: build.initialOptions.format,
            })
            const dataURI = `data:text/javascript;base64,${btoa(result.outputFiles[0].text)}`
            return {
                contents: `export default class { constructor() { return new Worker(${JSON.stringify(dataURI)}) } }`,
                loader: 'js',
            }
        })
    },
}
