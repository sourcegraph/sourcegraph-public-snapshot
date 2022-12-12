import path from 'path'

import * as esbuild from 'esbuild'

import {
    MONACO_LANGUAGES_AND_FEATURES,
    ROOT_PATH,
    STATIC_ASSETS_PATH,
    stylePlugin,
    packageResolutionPlugin,
    workerPlugin,
    monacoPlugin,
    RXJS_RESOLUTIONS,
    buildMonaco,
    experimentalNoticePlugin,
    buildTimerPlugin,
} from '@sourcegraph/build-config'

import { ENVIRONMENT_CONFIG } from '../utils'

import { manifestPlugin } from './manifestPlugin'

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
    jsx: 'automatic',
    jsxDev: true, // we're only using esbuild for dev server right now
    splitting: true,
    chunkNames: 'chunks/chunk-[name]-[hash]',
    outdir: STATIC_ASSETS_PATH,
    plugins: [
        stylePlugin,
        workerPlugin,
        manifestPlugin,
        packageResolutionPlugin({
            path: require.resolve('path-browserify'),
            ...RXJS_RESOLUTIONS,
        }),
        monacoPlugin(MONACO_LANGUAGES_AND_FEATURES),
        buildTimerPlugin,
        experimentalNoticePlugin,
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

export const build = async (): Promise<void> => {
    await esbuild.build({
        ...BUILD_OPTIONS,
        outdir: STATIC_ASSETS_PATH,
    })
    await buildMonaco(STATIC_ASSETS_PATH)
}

if (require.main === module) {
    build()
        .catch(error => console.error('Error:', error))
        .finally(() => process.exit(0))
}
