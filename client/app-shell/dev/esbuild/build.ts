import * as fs from 'fs'

import * as esbuild from 'esbuild'

import {
    stylePlugin,
    packageResolutionPlugin,
    experimentalNoticePlugin,
    buildTimerPlugin,
} from '@sourcegraph/build-config'
import { isDefined } from '@sourcegraph/common'

export const BUILD_OPTIONS: esbuild.BuildOptions = {
    entryPoints: ['src/app-shell.tsx'],
    bundle: true,
    format: 'esm',
    logLevel: 'debug',
    jsx: 'automatic',
    splitting: false,
    plugins: [
        stylePlugin,
        packageResolutionPlugin({
            path: require.resolve('path-browserify'),
        }),
        buildTimerPlugin,
        experimentalNoticePlugin,
    ].filter(isDefined),
    define: {
        global: 'window',
    },
    loader: {
        '.yaml': 'text',
        '.ttf': 'file',
        '.png': 'file',
    },
    target: 'esnext',
    sourcemap: true,
}

export const build = async (): Promise<void> => {
    const metafile = process.env.ESBUILD_METAFILE
    const result = await esbuild.build({
        ...BUILD_OPTIONS,
        outdir: 'dist/scripts/',
        metafile: Boolean(metafile),
    })
    if (metafile) {
        fs.writeFileSync(metafile, JSON.stringify(result.metafile), 'utf-8')
    }

    if (result.errors.length === 0) {
        const content = await fs.promises.readFile('index.html', 'utf8')
        fs.writeFileSync('dist/index.html', content)
        console.info('index.html written to dist/index.html')
    }
}

if (require.main === module) {
    build()
        .catch(error => console.error('Error:', error))
        .finally(() => process.exit(0))
}
