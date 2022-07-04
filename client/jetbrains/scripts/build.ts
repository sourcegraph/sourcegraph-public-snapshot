import { existsSync } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import { rm } from 'shelljs'

import { packageResolutionPlugin, stylePlugin, workerPlugin, buildTimerPlugin } from '@sourcegraph/build-config'

const rootPath = path.resolve(__dirname, '../../../')
const jetbrainsWorkspacePath = path.resolve(rootPath, 'client', 'jetbrains')
const webviewSourcePath = path.resolve(jetbrainsWorkspacePath, 'webview', 'src')

// Build artifacts are put directly into the JetBrains resources folder
const distributionPath = path.resolve(jetbrainsWorkspacePath, 'src', 'main', 'resources', 'dist')

export async function build(): Promise<void> {
    if (existsSync(distributionPath)) {
        rm('-rf', distributionPath)
    }

    const buildPromises = []

    buildPromises.push(
        esbuild.build({
            entryPoints: {
                search: path.resolve(webviewSourcePath, 'search', 'index.tsx'),
                bridgeMock: path.resolve(webviewSourcePath, 'bridge-mock', 'index.ts'),
                style: path.join(webviewSourcePath, 'index.scss'),
            },
            bundle: true,
            format: 'esm',
            platform: 'browser',
            define: {
                'process.env.IS_TEST': 'false',
                global: 'globalThis',
            },
            splitting: true,
            inject: ['./scripts/react-shim.js', './scripts/process-shim.js', './scripts/buffer-shim.js'],
            plugins: [
                stylePlugin,
                workerPlugin,
                packageResolutionPlugin({
                    process: require.resolve('process/browser'),
                    path: require.resolve('path-browserify'),
                    http: require.resolve('stream-http'),
                    https: require.resolve('https-browserify'),
                    // stream: require.resolve('stream-browserify'),
                    util: require.resolve('util'),
                    // events: require.resolve('events'),
                    // buffer: require.resolve('buffer/'),
                }),
                //     './Link': require.resolve('../src/webview/search-panel/alias/Link'),
                //     '../Link': require.resolve('../src/webview/search-panel/alias/Link'),
                //     './RepoSearchResult': require.resolve('../src/webview/search-panel/alias/RepoSearchResult'),
                //     './CommitSearchResult': require.resolve('../src/webview/search-panel/alias/CommitSearchResult'),
                //     './FileMatchChildren': require.resolve('../src/webview/search-panel/alias/FileMatchChildren'),
                //     './RepoFileLink': require.resolve('../src/webview/search-panel/alias/RepoFileLink'),
                //     '../documentation/ModalVideo': require.resolve('../src/webview/search-panel/alias/ModalVideo'),
                // }),
                // Note: leads to "file has no exports" warnings
                // monacoPlugin(MONACO_LANGUAGES_AND_FEATURES),
                buildTimerPlugin,
                // {
                //     name: 'codiconsDeduplicator',
                //     setup(build): void {
                //         build.onLoad({ filter: /\.ttf$/ }, args => {
                //             // Both `@jetbrains/codicons` and `monaco-editor`
                //             // node modules include a `codicons.ttf` file,
                //             // so null one out.
                //             if (!args.path.includes('@jetbrains/codicons')) {
                //                 return {
                //                     contents: '',
                //                     loader: 'text',
                //                 }
                //             }
                //             return null
                //         })
                //     },
                // },
            ],
            loader: {
                '.ttf': 'file',
            },
            assetNames: '[name]',
            ignoreAnnotations: true,
            treeShaking: false,
            watch: !!process.env.WATCH,
            minify: true, // process.env.NODE_ENV === 'production',
            sourcemap: true,
            outdir: distributionPath,
        })
    )

    // buildPromises.push(buildMonaco(outdir))

    await Promise.all(buildPromises)
}
