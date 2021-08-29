import * as path from 'path'

import * as esbuild from 'esbuild'

import { MONACO_LANGUAGES_AND_FEATURES } from '../webpack/monacoWebpack'

import { manifestPlugin } from './manifestPlugin'
import { monacoPlugin } from './monacoPlugin'
import { packageResolutionPlugin } from './packageResolutionPlugin'
import { sassPlugin } from './sassPlugin'
import { workerPlugin } from './workerPlugin'

const rootPath = path.resolve(__dirname, '..', '..', '..', '..')
export const uiAssetsPath = path.join(rootPath, 'ui', 'assets')
const isEnterpriseBuild = process.env.ENTERPRISE && Boolean(JSON.parse(process.env.ENTERPRISE))
const enterpriseDirectory = path.resolve(__dirname, '..', '..', 'src', 'enterprise')

export const esbuildOutDirectory = path.join(uiAssetsPath, 'esbuild')

// TODO(sqs): look into speeding this up by ignoring node_modules/monaco-editor/... entrypoints
export const BUILD_OPTIONS: esbuild.BuildOptions = {
    entryPoints: {
        // Enterprise vs. OSS builds use different entrypoints. The enterprise entrypoint imports a
        // strict superset of the OSS entrypoint.
        app: isEnterpriseBuild
            ? path.join(enterpriseDirectory, 'main.tsx')
            : path.join(__dirname, '..', '..', 'src', 'main.tsx'),
    },
    bundle: true,
    format: 'esm',
    logLevel: 'error',
    splitting: true,
    chunkNames: 'chunk-[name]-[hash]',
    plugins: [
        sassPlugin,
        workerPlugin,
        manifestPlugin,
        packageResolutionPlugin,
        monacoPlugin(MONACO_LANGUAGES_AND_FEATURES),
    ],
    define: {
        'process.env.NODE_ENV': '"development"',
        'process.env.PERCY_ON': JSON.stringify(process.env.PERCY_ON),
        'process.env.SOURCEGRAPH_API_URL': JSON.stringify(process.env.SOURCEGRAPH_API_URL),
    },
    loader: {
        '.yaml': 'text',
        '.ttf': 'file',
        '.png': 'file',
    },
    target: 'es2021',
    sourcemap: true,
    // incremental: true,
}

export const buildMonaco = async (): Promise<void> => {
    await esbuild.build({
        entryPoints: {
            'scripts/editor.worker.bundle': 'monaco-editor/esm/vs/editor/editor.worker.js',
            'scripts/json.worker.bundle': 'monaco-editor/esm/vs/language/json/json.worker.js',
        },
        format: 'iife',
        target: 'es2021',
        bundle: true,
        outdir: esbuildOutDirectory,
    })
}

export const build = async (): Promise<void> => {
    await esbuild.build({
        ...BUILD_OPTIONS,
        outdir: esbuildOutDirectory,
        incremental: false,
    })
    if (process.env.FOO) {
        // TODO(sqs): reenable
        await buildMonaco()
    }
}

if (require.main === module) {
    build()
        .catch(error => console.error('Error:', error))
        .finally(() => process.exit(0))
}
