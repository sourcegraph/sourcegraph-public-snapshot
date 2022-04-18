import { existsSync, rmdirSync } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'

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

export async function build(): Promise<void> {
    const outdir = path.join(__dirname, '../dist')

    const SHARED_CONFIG = {
        outdir,
        watch,
        minify,
    }

    if (existsSync(outdir)) {
        rmdirSync(outdir, { recursive: true })
    }

    // TODO: build webworker AND node
    await esbuild.build({
        entryPoints: { extension: path.join(__dirname, '/../src/extension.ts') },
        bundle: true,
        format: 'cjs',
        platform: 'node',
        external: ['vscode'],
        ...SHARED_CONFIG,
    })

    await esbuild.build({
        entryPoints: {
            helpSidebar: path.join(__dirname, '../src/webview/sidebars/help'),
            searchSidebar: path.join(__dirname, '../src/webview/sidebars/search'),
            searchPanel: path.join(__dirname, '../src/webview/search-panel'),
            // For our style-plugin, has to be relative to resolveDir.
            style: './src/webview/index.scss',
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
    })

    await buildMonaco(outdir)
}
