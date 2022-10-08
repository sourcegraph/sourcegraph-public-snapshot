import { existsSync } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import { rm } from 'shelljs'

import { stylePlugin, buildTimerPlugin } from '@sourcegraph/build-config'

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
        plugins: [stylePlugin, buildTimerPlugin],
        assetNames: '[name]',
        watch: !!process.env.WATCH,
        sourcemap: true,
        outdir: distributionPath,
    })
}
