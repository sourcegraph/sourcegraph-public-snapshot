import { existsSync } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import { rm } from 'shelljs'

import { WORKSPACES_PATH, buildTimerPlugin } from '@sourcegraph/build-config'

const PACKAGE_ROOT_PATH = path.resolve(WORKSPACES_PATH, 'backstage-frontend')
const DIST_PATH = path.resolve(PACKAGE_ROOT_PATH, 'dist')

async function build(): Promise<void> {
    if (existsSync(DIST_PATH)) {
        rm('-rf', DIST_PATH)
    }

    await esbuild.build({
        entryPoints: [path.resolve(PACKAGE_ROOT_PATH, 'src', 'index.ts')],
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
        // Most likely due to https://github.com/evanw/esbuild/issues/1619#issuecomment-922787629
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
        outdir: DIST_PATH,
    })
}

if (require.main == module) {
    build().catch(error => {
        console.error('Error:', error)
        process.exit(1)
    })
}
