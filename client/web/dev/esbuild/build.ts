import * as path from 'path'

import * as esbuild from 'esbuild'
import signale from 'signale'

import { MONACO_LANGUAGES_AND_FEATURES } from '../webpack/monacoWebpack'

import { manifestPlugin } from './manifestPlugin'
import { monacoPlugin } from './monacoPlugin'
import { packageResolutionPlugin } from './packageResolutionPlugin'
import { stylePlugin } from './stylePlugin'
import { workerPlugin } from './workerPlugin'

const rootPath = path.resolve(__dirname, '..', '..', '..', '..')
export const uiAssetsPath = path.join(rootPath, 'ui', 'assets')
const isEnterpriseBuild = process.env.ENTERPRISE && Boolean(JSON.parse(process.env.ENTERPRISE))

export const esbuildOutDirectory = uiAssetsPath

export const BUILD_OPTIONS: esbuild.BuildOptions = {
    entryPoints: {
        // Enterprise vs. OSS builds use different entrypoints. The enterprise entrypoint imports a
        // strict superset of the OSS entrypoint.
        'scripts/app': isEnterpriseBuild
            ? path.join(__dirname, '../../src/enterprise/main.tsx')
            : path.join(__dirname, '../../src/main.tsx'),
    },
    bundle: true,
    format: 'esm',
    logLevel: 'error',
    splitting: true,
    chunkNames: 'chunks/chunk-[name]-[hash]',
    outdir: esbuildOutDirectory,
    plugins: [
        stylePlugin,
        workerPlugin,
        manifestPlugin,
        packageResolutionPlugin({
            path: require.resolve('path-browserify'),
        }),
        monacoPlugin(MONACO_LANGUAGES_AND_FEATURES),
        {
            name: 'buildTimer',
            setup: (build: esbuild.PluginBuild): void => {
                let buildStarted: number
                build.onStart(() => {
                    buildStarted = Date.now()
                })
                build.onEnd(() => console.log(`# esbuild: build took ${Date.now() - buildStarted}ms`))
            },
        },
        {
            name: 'experimentalNotice',
            setup: (): void => {
                signale.info(
                    'esbuild usage is experimental. See https://docs.sourcegraph.com/dev/background-information/web/build#esbuild.'
                )
            },
        },
    ],
    define: {
        'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development'),
        'process.env.PERCY_ON': JSON.stringify(process.env.PERCY_ON),
        'process.env.SOURCEGRAPH_API_URL': JSON.stringify(process.env.SOURCEGRAPH_API_URL),
    },
    loader: {
        '.yaml': 'text',
        '.ttf': 'file',
        '.png': 'file',
    },
    target: 'es2021',
    sourcemap: true,

    // TODO(sqs): When https://github.com/evanw/esbuild/pull/1458 is merged (or the issue is
    // otherwise fixed), we can return to using tree shaking. Right now, esbuild's tree shaking has
    // a bug where the NavBar CSS is not loaded because the @sourcegraph/wildcard uses `export *
    // from` and has `"sideEffects": false` in its package.json.
    treeShaking: 'ignore-annotations',
}

// TODO(sqs): These Monaco Web Workers could be built as part of the main build if we switch to
// using MonacoEnvironment#getWorker (from #getWorkerUrl), which would then let us use the worker
// plugin (and in Webpack the worker-loader) to load these instead of needing to hardcode them as
// build entrypoints.
export const buildMonaco = async (): Promise<void> => {
    await esbuild.build({
        entryPoints: {
            'scripts/editor.worker.bundle': 'monaco-editor/esm/vs/editor/editor.worker.js',
            'scripts/json.worker.bundle': 'monaco-editor/esm/vs/language/json/json.worker.js',
        },
        format: 'iife',
        target: 'es2021',
        bundle: true,
        outdir: esbuildOutDirectory,
    })
}

export const build = async (): Promise<void> => {
    await esbuild.build({
        ...BUILD_OPTIONS,
        outdir: esbuildOutDirectory,
    })
    await buildMonaco()
}

if (require.main === module) {
    build()
        .catch(error => console.error('Error:', error))
        .finally(() => process.exit(0))
}
