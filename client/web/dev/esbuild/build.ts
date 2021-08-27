import * as path from 'path'

import * as esbuild from 'esbuild'

import { manifestPlugin } from './manifestPlugin'
import { packageResolutionPlugin } from './packageResolutionPlugin'
import { sassPlugin } from './sassPlugin'
import { workerPlugin } from './workerPlugin'

const rootPath = path.resolve(__dirname, '..', '..', '..', '..')
export const uiAssetsPath = path.join(rootPath, 'ui', 'assets')
const isEnterpriseBuild = process.env.ENTERPRISE && Boolean(JSON.parse(process.env.ENTERPRISE))
const enterpriseDirectory = path.resolve(__dirname, '..', '..', 'src', 'enterprise')

export const BUILD_OPTIONS: esbuild.BuildOptions = {
    entryPoints: [
        // Enterprise vs. OSS builds use different entrypoints. The enterprise entrypoint imports a
        // strict superset of the OSS entrypoint.
        isEnterpriseBuild
            ? path.join(enterpriseDirectory, 'main.tsx')
            : path.join(__dirname, '..', '..', 'src', 'main.tsx'),
        path.join(__dirname, '..', '..', '..', 'shared/src/api/extension/main.worker.ts'),
        // TODO(sqs): webpack has some monaco entrypoints, do we need these?
    ],
    bundle: true,
    format: 'esm',
    outdir: path.join(uiAssetsPath, 'esbuild'),
    logLevel: 'error',
    splitting: false, // TODO(sqs): need to have splitting:false for main.worker.ts entrypoint
    plugins: [sassPlugin, workerPlugin, manifestPlugin, packageResolutionPlugin],
    define: {
        'process.env.NODE_ENV': '"development"',
        global: 'window',
        'process.env.SOURCEGRAPH_API_URL': JSON.stringify(process.env.SOURCEGRAPH_API_URL),
    },
    loader: {
        '.yaml': 'text',
        '.ttf': 'file',
        '.png': 'file',
    },
    target: 'es2020',
    sourcemap: true,
    incremental: true,
}
