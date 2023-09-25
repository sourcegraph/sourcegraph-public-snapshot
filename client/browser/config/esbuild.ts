import path from 'path'

import type * as esbuild from 'esbuild'
import shelljs from 'shelljs'
import signale from 'signale'

import { buildTimerPlugin, experimentalNoticePlugin, stylePlugin } from '@sourcegraph/build-config'

import { buildNpm } from '../scripts/build-npm'
import * as tasks from '../scripts/tasks'

import { entrypoints, browserWorkspacePath } from './buildCommon'

/**
 * Returns the esbuild build options for the browser extension build.
 */
export function esbuildBuildOptions(mode: 'dev' | 'prod'): esbuild.BuildOptions {
    return {
        entryPoints: { ...entrypoints, extensionHostWorker: '../shared/src/api/extension/main.worker.ts' },
        format: 'cjs',
        platform: 'browser',
        plugins: [
            stylePlugin,
            copyAssetsPlugin,
            browserBuildStepsPlugin(mode),
            buildTimerPlugin,
            experimentalNoticePlugin,
        ],
        bundle: true,
        minify: false,
        logLevel: 'error',
        jsx: 'automatic',
        outdir: path.join(browserWorkspacePath, 'build/dist'),
        chunkNames: '[ext]/[hash].chunk',
        entryNames: '[ext]/[name].bundle',
        target: 'esnext',
        sourcemap: true,
        alias: { path: 'path-browserify' },
        loader: {
            '.svg': 'text',
        },
    }
}

const copyAssetsPlugin: esbuild.Plugin = {
    name: 'copyAssets',
    setup: (): void => {
        tasks.copyAssets()
    },
}

function browserBuildStepsPlugin(mode: 'dev' | 'prod'): esbuild.Plugin {
    return {
        name: 'browserBuildSteps',
        setup: (build: esbuild.PluginBuild): void => {
            build.onEnd(async result => {
                if (result.errors.length === 0) {
                    signale.info('Running browser build steps...')
                    tasks.buildChrome(mode)()
                    tasks.buildFirefox(mode)()
                    tasks.buildEdge(mode)()
                    if (mode === 'prod') {
                        if (isXcodeAvailable()) {
                            tasks.buildSafari(mode)()
                        } else {
                            signale.debug(
                                'Skipping Safari build because Xcode tools were not found (xcrun, xcodebuild)'
                            )
                        }
                    }
                    tasks.copyIntegrationAssets()
                    if (mode === 'prod') {
                        await buildNpm()
                    }
                }
            })
        },
    }
}

function isXcodeAvailable(): boolean {
    return !!shelljs.which('xcrun') && !!shelljs.which('xcodebuild')
}
