import { existsSync } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import { rm } from 'shelljs'

import {
    buildMonaco,
    monacoPlugin,
    MONACO_LANGUAGES_AND_FEATURES,
    packageResolutionPlugin,
    stylePlugin,
    workerPlugin,
    RXJS_RESOLUTIONS,
    buildTimerPlugin,
} from '@sourcegraph/build-config'

const watch = !!process.env.WATCH
const minify = process.env.NODE_ENV === 'production'
const outdir = path.join(__dirname, '../dist')
const isTest = !!process.env.IS_TEST

const TARGET_TYPE = process.env.TARGET_TYPE

const SHARED_CONFIG = {
    outdir,
    watch,
    minify,
    sourcemap: true,
}

export async function build(): Promise<void> {
    if (existsSync(outdir)) {
        rm('-rf', outdir)
    }

    const buildPromises = []

    if (TARGET_TYPE === 'node' || !TARGET_TYPE) {
        buildPromises.push(
            esbuild.build({
                entryPoints: { extension: path.join(__dirname, '/../src/extension.ts') },
                bundle: true,
                format: 'cjs',
                platform: 'node',
                external: ['vscode'],
                banner: { js: 'global.Buffer = require("buffer").Buffer' },
                define: {
                    'process.env.IS_TEST': isTest ? 'true' : 'false',
                },
                ...SHARED_CONFIG,
                outdir: path.join(SHARED_CONFIG.outdir, 'node'),
            })
        )
    }
    if (TARGET_TYPE === 'webworker' || !TARGET_TYPE) {
        buildPromises.push(
            esbuild.build({
                entryPoints: { extension: path.join(__dirname, '/../src/extension.ts') },
                bundle: true,
                format: 'cjs',
                platform: 'browser',
                external: ['vscode'],
                define: {
                    'process.env.IS_TEST': isTest ? 'true' : 'false',
                    global: 'globalThis',
                },
                inject: ['./scripts/process-shim.js', './scripts/buffer-shim.js'],
                plugins: [
                    packageResolutionPlugin({
                        process: require.resolve('process/browser'),
                        path: require.resolve('path-browserify'),
                        http: require.resolve('stream-http'),
                        https: require.resolve('https-browserify'),
                        stream: require.resolve('stream-browserify'),
                        util: require.resolve('util'),
                        events: require.resolve('events'),
                        buffer: require.resolve('buffer/'),
                        './browserActionsNode': path.resolve(__dirname, '../src', 'commands', 'browserActionsWeb'),
                    }),
                ],
                ...SHARED_CONFIG,
                outdir: path.join(SHARED_CONFIG.outdir, 'webworker'),
            })
        )
    }

    buildPromises.push(
        esbuild.build({
            entryPoints: {
                helpSidebar: path.join(__dirname, '../src/webview/sidebars/help'),
                searchSidebar: path.join(__dirname, '../src/webview/sidebars/search'),
                searchPanel: path.join(__dirname, '../src/webview/search-panel'),
                style: path.join(__dirname, '../src/webview/index.scss'),
            },
            bundle: true,
            format: 'esm',
            platform: 'browser',
            splitting: true,
            plugins: [
                stylePlugin,
                workerPlugin,
                packageResolutionPlugin({
                    path: require.resolve('path-browserify'),
                    ...RXJS_RESOLUTIONS,
                    './RepoSearchResult': require.resolve('../src/webview/search-panel/alias/RepoSearchResult'),
                    './CommitSearchResult': require.resolve('../src/webview/search-panel/alias/CommitSearchResult'),
                    './FileMatchChildren': require.resolve('../src/webview/search-panel/alias/FileMatchChildren'),
                    './RepoFileLink': require.resolve('../src/webview/search-panel/alias/RepoFileLink'),
                    '../documentation/ModalVideo': require.resolve('../src/webview/search-panel/alias/ModalVideo'),
                }),
                // Note: leads to "file has no exports" warnings
                monacoPlugin(MONACO_LANGUAGES_AND_FEATURES),
                buildTimerPlugin,
                {
                    name: 'codiconsDeduplicator',
                    setup(build): void {
                        build.onLoad({ filter: /\.ttf$/ }, args => {
                            // Both `@vscode/codicons` and `monaco-editor`
                            // node modules include a `codicons.ttf` file,
                            // so null one out.
                            if (!args.path.includes('@vscode/codicons')) {
                                return {
                                    contents: '',
                                    loader: 'text',
                                }
                            }
                            return null
                        })
                    },
                },
            ],
            loader: {
                '.ttf': 'file',
            },
            assetNames: '[name]',
            ignoreAnnotations: true,
            treeShaking: false,
            ...SHARED_CONFIG,
            outdir: path.join(SHARED_CONFIG.outdir, 'webview'),
        })
    )

    buildPromises.push(buildMonaco(outdir))

    await Promise.all(buildPromises)
}
