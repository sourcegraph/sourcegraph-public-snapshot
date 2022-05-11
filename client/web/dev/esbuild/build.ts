import path from 'path'

import * as esbuild from 'esbuild'
import signale from 'signale'

import { MONACO_LANGUAGES_AND_FEATURES } from '@sourcegraph/build-config'

import { ENVIRONMENT_CONFIG, ROOT_PATH, STATIC_ASSETS_PATH } from '../utils'

import { manifestPlugin } from './manifestPlugin'
import { monacoPlugin } from './monacoPlugin'
import { packageResolutionPlugin } from './packageResolutionPlugin'
import { stylePlugin } from './stylePlugin'
import { workerPlugin } from './workerPlugin'

const isEnterpriseBuild = ENVIRONMENT_CONFIG.ENTERPRISE

export const BUILD_OPTIONS: esbuild.BuildOptions = {
    entryPoints: {
        // Enterprise vs. OSS builds use different entrypoints. The enterprise entrypoint imports a
        // strict superset of the OSS entrypoint.
        'scripts/app': isEnterpriseBuild
            ? path.join(ROOT_PATH, 'client/web/src/enterprise/main.tsx')
            : path.join(ROOT_PATH, 'client/web/src/main.tsx'),
    },
    bundle: true,
    format: 'esm',
    logLevel: 'error',
    splitting: true,
    chunkNames: 'chunks/chunk-[name]-[hash]',
    outdir: STATIC_ASSETS_PATH,
    plugins: [
        stylePlugin,
        workerPlugin,
        manifestPlugin,
        packageResolutionPlugin({
            path: require.resolve('path-browserify'),

            // Needed because imports of rxjs/internal/... actually import a different variant of
            // rxjs in the same package, which leads to observables from combineLatestOrDefault (and
            // other places that use rxjs/internal/...) not being cross-compatible. See
            // https://stackoverflow.com/questions/53758889/rxjs-subscribeto-js-observable-check-works-in-chrome-but-fails-in-chrome-incogn.
            'rxjs/internal/OuterSubscriber': require.resolve('rxjs/_esm5/internal/OuterSubscriber'),
            'rxjs/internal/util/subscribeToResult': require.resolve('rxjs/_esm5/internal/util/subscribeToResult'),
            'rxjs/internal/util/subscribeToArray': require.resolve('rxjs/_esm5/internal/util/subscribeToArray'),
            'rxjs/internal/Observable': require.resolve('rxjs/_esm5/internal/Observable'),
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
        ...Object.fromEntries(
            Object.entries({ ...ENVIRONMENT_CONFIG, SOURCEGRAPH_API_URL: undefined }).map(([key, value]) => [
                `process.env.${key}`,
                JSON.stringify(value),
            ])
        ),
        global: 'window',
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
    ignoreAnnotations: true,
    treeShaking: false,
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
        outdir: STATIC_ASSETS_PATH,
    })
}

export const build = async (): Promise<void> => {
    await esbuild.build({
        ...BUILD_OPTIONS,
        outdir: STATIC_ASSETS_PATH,
    })
    await buildMonaco()
}

if (require.main === module) {
    build()
        .catch(error => console.error('Error:', error))
        .finally(() => process.exit(0))
}
