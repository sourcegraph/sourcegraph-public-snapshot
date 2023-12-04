import { promises as fs } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'

import { stylePlugin, packageResolutionPlugin, buildTimerPlugin } from '@sourcegraph/build-config'
import { isDefined } from '@sourcegraph/common'

async function copyStaticFiles(sourceDir: string, destinationDir: string): Promise<void> {
    await fs.mkdir(destinationDir, { recursive: true })

    const files = await fs.readdir(sourceDir)

    for (const file of files) {
        const sourceFile = path.join(sourceDir, file)
        const destinationFile = path.join(destinationDir, file)
        await fs.copyFile(sourceFile, destinationFile)
        console.info(`${sourceFile} written to ${destinationFile}`)
    }
}

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
        await fs.writeFile(metafile, JSON.stringify(result.metafile), 'utf-8')
    }

    if (result.errors.length === 0) {
        await copyStaticFiles('static', 'dist')
    }
}

if (require.main === module) {
    build()
        .catch(error => console.error('Error:', error))
        .finally(() => process.exit(0))
}
