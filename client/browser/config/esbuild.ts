import path from 'path'

import type * as esbuild from 'esbuild'
import shelljs from 'shelljs'
import signale from 'signale'

import { ROOT_PATH, stylePlugin } from '@sourcegraph/build-config'

import { buildNpm } from './build-npm'
import * as tasks from './tasks'
import { generateBundleUID } from './utils'

const browserWorkspacePath = path.resolve(ROOT_PATH, 'client/browser')
const browserSourcePath = path.resolve(browserWorkspacePath, 'src')

/**
 * Returns the esbuild build options for the browser extension build.
 */
export function esbuildBuildOptions(mode: 'dev' | 'prod'): esbuild.BuildOptions {
    return {
        entryPoints: {
            // Browser extension
            background: path.resolve(browserSourcePath, 'browser-extension/scripts/backgroundPage.main.ts'),
            inject: path.resolve(browserSourcePath, 'browser-extension/scripts/contentPage.main.ts'),
            options: path.resolve(browserSourcePath, 'browser-extension/scripts/optionsPage.main.tsx'),
            'after-install': path.resolve(browserSourcePath, 'browser-extension/scripts/afterInstallPage.main.tsx'),

            // Common native integration entry point (Gitlab, Bitbucket)
            integration: path.resolve(browserSourcePath, 'native-integration/integration.main.ts'),
            // Phabricator-only native integration entry point
            phabricator: path.resolve(browserSourcePath, 'native-integration/phabricator/integration.main.ts'),

            // Styles
            style: path.join(browserSourcePath, 'app.scss'),
            'branded-style': path.join(browserSourcePath, 'branded.scss'),

            // Worker
            extensionHostWorker: path.resolve(browserSourcePath, 'shared/main.worker.ts'),
        },
        format: 'cjs',
        platform: 'browser',
        plugins: [stylePlugin, copyAssetsPlugin, browserBuildStepsPlugin(mode)],
        define: {
            'process.env.NODE_ENV': mode === 'dev' ? 'development' : 'production',
            'process.env.BUNDLE_UID': JSON.stringify(generateBundleUID()),
        },
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
