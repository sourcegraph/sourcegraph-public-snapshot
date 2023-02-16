import { existsSync } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import { rm } from 'shelljs'

import { buildTimerPlugin } from '@sourcegraph/build-config'

const distributionPath = path.resolve(__dirname, '..', 'dist')

;(async function build(): Promise<void> {
    if (existsSync(distributionPath)) {
        rm('-rf', distributionPath)
    }

    await esbuild.build({
        entryPoints: [path.resolve(__dirname, '..', 'src', 'index.ts')],
        bundle: true,
        external: [
            '@backstage/cli',
            '@backstage/catalog-model',
            '@backstage/config',
            '@backstage/backend-common',
            '@backstage/plugin-catalog-backend',
            'graphql-request',
            'react',
            'react-dom',
            'lodash',
            'apollo',
            'winston',
        ],
        // If we don't specify module first, esbuild bundles somethings "incorrectly" and you'll get a error with './impl/format' error
        mainFields: ['module', 'main'],
        format: 'cjs',
        platform: 'node',
        define: {
            'process.env.IS_TEST': 'false',
            global: 'globalThis',
        },
        plugins: [buildTimerPlugin],
        ignoreAnnotations: true,
        treeShaking: true,
        sourcemap: true,
        outdir: distributionPath,
    })
})()
