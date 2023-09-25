import path from 'path'

import * as esbuild from 'esbuild'
import signale from 'signale'

import { buildTimerPlugin, experimentalNoticePlugin, stylePlugin } from '@sourcegraph/build-config'

import { browserWorkspacePath, config } from '../config/webpack/base.config'

import * as tasks from './tasks'

signale.config({ displayTimestamp: true })

const buildChrome = tasks.buildChrome('dev')
const buildFirefox = tasks.buildFirefox('dev')
const buildEdge = tasks.buildEdge('dev')

tasks.copyAssets()

signale.info('Running esbuild')

const browserBuildStepsPlugin: esbuild.Plugin = {
    name: 'browserBuildSteps',
    setup: (build: esbuild.PluginBuild): void => {
        build.onEnd(result => {
            if (result.errors.length === 0) {
                signale.info('Running browser build steps...')
                buildChrome()
                buildEdge()
                buildFirefox()
                tasks.copyIntegrationAssets()
            }
        })
    },
}

const COMMON_BUILD_OPTIONS: esbuild.BuildOptions = {
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

const BUILD_OPTIONS: esbuild.BuildOptions = {
    ...COMMON_BUILD_OPTIONS,
    entryPoints: { ...config.entry, extensionHostWorker: '../shared/src/api/extension/main.worker.ts' },
    format: 'cjs',
    platform: 'browser',
    plugins: [stylePlugin, browserBuildStepsPlugin, buildTimerPlugin, experimentalNoticePlugin],
}

async function build(): Promise<void> {
    const ctx = await esbuild.context(BUILD_OPTIONS)
    await ctx.watch()
    signale.info('Watching...')
    await new Promise(() => {}) // wait forever
}

if (require.main === module) {
    build()
        .catch(error => console.error('Error:', error))
        .finally(() => process.exit(0))
}
