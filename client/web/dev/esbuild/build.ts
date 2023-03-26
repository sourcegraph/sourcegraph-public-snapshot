import { writeFileSync } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'

import {
    MONACO_LANGUAGES_AND_FEATURES,
    ROOT_PATH,
    STATIC_ASSETS_PATH,
    stylePlugin,
    packageResolutionPlugin,
    monacoPlugin,
    RXJS_RESOLUTIONS,
    buildMonaco,
    experimentalNoticePlugin,
    buildTimerPlugin,
} from '@sourcegraph/build-config'
import { isDefined } from '@sourcegraph/common'

import { ENVIRONMENT_CONFIG, IS_DEVELOPMENT, IS_PRODUCTION } from '../utils'

import { manifestPlugin } from './manifestPlugin'

const isEnterpriseBuild = ENVIRONMENT_CONFIG.ENTERPRISE
const omitSlowDeps = ENVIRONMENT_CONFIG.DEV_WEB_BUILDER_OMIT_SLOW_DEPS

export const BUILD_OPTIONS: esbuild.BuildOptions = {
    entryPoints: {
        // Enterprise vs. OSS builds use different entrypoints. The enterprise entrypoint imports a
        // strict superset of the OSS entrypoint.
        'scripts/app': isEnterpriseBuild
            ? path.join(ROOT_PATH, 'client/web/src/enterprise/main.tsx')
            : path.join(ROOT_PATH, 'client/web/src/main.tsx'),
    },
    bundle: true,
    minify: IS_PRODUCTION,
    format: 'esm',
    logLevel: 'error',
    jsx: 'automatic',
    jsxDev: IS_DEVELOPMENT,
    splitting: true,
    chunkNames: 'chunks/chunk-[name]-[hash]',
    entryNames: IS_PRODUCTION ? 'scripts/[name]-[hash]' : undefined,
    outdir: STATIC_ASSETS_PATH,
    plugins: [
        stylePlugin,
        manifestPlugin,
        packageResolutionPlugin({
            path: require.resolve('path-browserify'),
            ...RXJS_RESOLUTIONS,
            ...(omitSlowDeps
                ? {
                      // Monaco
                      '@sourcegraph/shared/src/components/MonacoEditor':
                          '@sourcegraph/shared/src/components/NoMonacoEditor',
                      'monaco-editor': '/dev/null',
                      'monaco-editor/esm/vs/editor/editor.api': '/dev/null',
                      'monaco-yaml': '/dev/null',

                      // GraphiQL
                      './api/ApiConsole': path.join(ROOT_PATH, 'client/web/src/api/NoApiConsole.tsx'),
                      '@graphiql/react': '/dev/null',
                      graphiql: '/dev/null',

                      // Misc.
                      recharts: '/dev/null',
                      '@visx/annotation': '/dev/null',
                      '@visx/axis': '/dev/null',
                      '@visx/event': '/dev/null',
                      '@visx/glyph': '/dev/null',
                      '@visx/grid': '/dev/null',
                      '@visx/group': '/dev/null',
                      '@visx/responsive': '/dev/null',
                      '@visx/scale': '/dev/null',
                      '@visx/shape': '/dev/null',
                      '@visx/text': '/dev/null',
                      '@visx/voronoi': '/dev/null',
                      'd3-format': '/dev/null',
                      'd3-scale': '/dev/null',
                      'd3-shape': '/dev/null',
                      'd3-time-format': '/dev/null',
                      'd3-voronoi': '/dev/null',
                      'sanitize-html': '/dev/null',
                  }
                : null),
        }),
        omitSlowDeps ? null : monacoPlugin(MONACO_LANGUAGES_AND_FEATURES),
        buildTimerPlugin,
        experimentalNoticePlugin,
    ].filter(isDefined),
    define: {
        ...Object.fromEntries(
            Object.entries({
                ...ENVIRONMENT_CONFIG,
                SOURCEGRAPH_API_URL: undefined,
                DISABLE_NOTEBOOKS: true,
                DISABLE_BATCH_CHANGES: true,
                DISABLE_CODE_MONITORING: true,
                DISABLE_CODE_INSIGHTS: true,
                DISABLE_ADMIN_ANALYTICS: true,
                DISABLE_SEARCH_AGGREGATIONS: true,
            }).map(([key, value]) => [`process.env.${key}`, JSON.stringify(value === undefined ? null : value)])
        ),
        global: 'window',
    },
    loader: {
        '.yaml': 'text',
        '.ttf': 'file',
        '.png': 'file',
    },
    target: 'esnext',
    sourcemap: false,
}

export const build = async (): Promise<void> => {
    const metafile = process.env.ESBUILD_METAFILE
    const result = await esbuild.build({
        ...BUILD_OPTIONS,
        outdir: STATIC_ASSETS_PATH,
        metafile: Boolean(metafile),
    })
    if (metafile) {
        writeFileSync(metafile, JSON.stringify(result.metafile), 'utf-8')
    }
    if (!omitSlowDeps) {
        const ctx = await buildMonaco(STATIC_ASSETS_PATH)
        await ctx.rebuild()
        await ctx.dispose()
    }
}

if (require.main === module) {
    build()
        .catch(error => console.error('Error:', error))
        .finally(() => process.exit(0))
}
