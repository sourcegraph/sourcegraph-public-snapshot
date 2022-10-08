import { existsSync } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import { rm } from 'shelljs'

import { packageResolutionPlugin, stylePlugin, workerPlugin, buildTimerPlugin } from '@sourcegraph/build-config'

const rootPath = path.resolve(__dirname, '../../../')
const sourceboxRootPath = path.join(rootPath, 'client', 'sourcebox')
const sourcePath = path.join(sourceboxRootPath, 'src')

const distributionPath = path.join(sourceboxRootPath, 'dist')

export async function build(): Promise<void> {
    if (existsSync(distributionPath)) {
        rm('-rf', distributionPath)
    }

    await esbuild.build({
        entryPoints: {
            sourcebox: path.join(sourcePath, 'index.tsx'),
        },
        bundle: true,
        format: 'esm',
        platform: 'browser',
        splitting: true,
        // TODO(sqs): is `inject` needed?
        inject: ['./scripts/react-shim.js', './scripts/process-shim.js', './scripts/buffer-shim.js'],
        plugins: [
            stylePlugin,
            workerPlugin,
            packageResolutionPlugin({
                // TODO(sqs): is this needed?
                process: require.resolve('process/browser'),
                path: require.resolve('path-browserify'),
                http: require.resolve('stream-http'),
                https: require.resolve('https-browserify'),
                util: require.resolve('util'),
            }),
            buildTimerPlugin,
        ],
        assetNames: '[name]',
        ignoreAnnotations: true,
        treeShaking: true,
        watch: !!process.env.WATCH,
        minify: true,
        sourcemap: true,
        outdir: distributionPath,
    })
}
