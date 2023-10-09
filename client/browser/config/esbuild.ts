import path from 'path'

import type * as esbuild from 'esbuild'

import { ROOT_PATH } from '@sourcegraph/build-config'
import { stylePlugin } from '@sourcegraph/build-config/src/esbuild/plugins'

import { generateBundleUID } from './utils'

const browserWorkspacePath = path.resolve(ROOT_PATH, 'client/browser')
const browserSourcePath = path.resolve(browserWorkspacePath, 'src')

/**
 * Returns the esbuild build options for the browser extension build.
 */
export function esbuildBuildOptions(mode: 'dev' | 'prod', extraPlugins: esbuild.Plugin[] = []): esbuild.BuildOptions {
    return {
        entryPoints: [
            // Browser extension
            path.resolve(browserSourcePath, 'browser-extension/scripts/backgroundPage.main.ts'),
            path.resolve(browserSourcePath, 'browser-extension/scripts/contentPage.main.ts'),
            path.resolve(browserSourcePath, 'browser-extension/scripts/optionsPage.main.tsx'),
            path.resolve(browserSourcePath, 'browser-extension/scripts/afterInstallPage.main.tsx'),

            // Common native integration entry point (Gitlab, Bitbucket)
            path.resolve(browserSourcePath, 'native-integration/nativeIntegration.main.ts'),
            // Phabricator-only native integration entry point
            path.resolve(browserSourcePath, 'native-integration/phabricator/phabricatorNativeIntegration.main.ts'),

            // Styles
            path.join(browserSourcePath, 'app.scss'),
            path.join(browserSourcePath, 'branded.scss'),

            // Worker
            path.resolve(browserSourcePath, 'shared/extensionHostWorker.ts'),
        ],
        format: 'cjs',
        platform: 'browser',
        plugins: [stylePlugin, ...extraPlugins],
        define: {
            'process.env.NODE_ENV': JSON.stringify(mode === 'dev' ? 'development' : 'production'),
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
