import child_process from 'child_process'
import fs from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'
import signale from 'signale'

import { BUILD_OPTIONS } from '@sourcegraph/web/dev/esbuild/build'
import { omitPackagePlugin } from '@sourcegraph/web/dev/esbuild/omitPackagePlugin'

import { omitNonModuleStylePlugin } from './dev/omitNonModuleStylePlugin'

const outputDirectory = path.join(__dirname, 'out')
const outputBundlePath = path.join(outputDirectory, 'prerender.js')

const PRERENDER_BUILD_OPTIONS: esbuild.BuildOptions = {
    ...BUILD_OPTIONS,
    entryPoints: ['src/serve.ts'],
    splitting: false,
    outdir: undefined,
    outfile: outputBundlePath,
    platform: 'node',
    external: ['node-fetch'],
    format: 'iife',

    // TODO(sqs): Tree shaking is disabled in the default build options due to
    // https://github.com/evanw/esbuild/pull/1458, but that bug doesn't affect the prerender
    // bundle since it only affects the CSS. So, we can enable (i.e., un-disable) tree shaking
    // in this build.
    treeShaking: undefined,

    // Don't alias `global` to `window`, because `global` is actually meaningful when running in
    // Node.
    define: Object.fromEntries(Object.entries(BUILD_OPTIONS.define || {}).filter(([name]) => name !== 'global')),

    // Omit monaco-editor from the bundle because it's never used on the server and is quite large.
    plugins: [
        omitNonModuleStylePlugin,
        omitPackagePlugin(['monaco-editor(/.*)?', 'recharts', 'visx/.*', 'cjs']),
        ...(BUILD_OPTIONS.plugins || []),
    ],
}

export const buildBundle = async (): Promise<void> => {
    await esbuild.build(PRERENDER_BUILD_OPTIONS)
    await logBuiltBundle()
}

export const watchBundleAndServe = async (): Promise<void> => {
    const server = newServer()
    const onBuildOrRebuild = async (error: esbuild.BuildFailure | null): Promise<void> => {
        if (error) {
            signale.error('Build error:', error)
        } else {
            await logBuiltBundle()
            await server.restart()

            // For convenience when running `jest --watch`, touch the test file so that the test reruns
            // after the server is updated.
            // await new Promise(resolve => setTimeout(resolve, 500)) // wait for server to restart
            const date = Date.now()
            await fs.promises.utimes(path.join(__dirname, 'src/handle.test.ts'), date, date)
        }
    }

    const result = await esbuild.build({
        ...PRERENDER_BUILD_OPTIONS,
        watch: {
            onRebuild: onBuildOrRebuild,
        },
    })
    await onBuildOrRebuild(
        result.errors.length > 0
            ? { ...new Error('esbuild build failure'), errors: result.errors, warnings: result.warnings }
            : null
    )
}

async function logBuiltBundle(): Promise<void> {
    const stat = await fs.promises.stat(outputBundlePath)
    signale.success(
        `Built bundle: ${path.relative(__dirname, outputBundlePath)} [${(stat.size / (1024 * 1024)).toFixed(1)}MB]`
    )
}

function newServer(): {
    restart: () => Promise<void>
    kill: () => Promise<void>
} {
    let current: child_process.ChildProcess | undefined

    const spawn = (): Promise<child_process.ChildProcess> => {
        const childProcess = child_process.spawn('node', ['-r', 'source-map-support/register', outputBundlePath], {
            stdio: 'inherit',
        })
        childProcess.once('close', exitCode => {
            current = undefined
            if (exitCode) {
                signale.error(`Server exited with error (${exitCode})`)
            }
        })
        return new Promise<child_process.ChildProcess>((resolve, reject) => {
            childProcess.once('error', error => reject(error))
            childProcess.once('spawn', () => resolve(childProcess))
        })
    }

    const kill = (): Promise<void> =>
        new Promise<void>((resolve, reject) => {
            if (current) {
                current.once('exit', () => resolve())
                current.once('error', error => reject(error))
                if (!current.kill()) {
                    reject(new Error('kill() failed'))
                }
                current = undefined
            } else {
                resolve()
            }
        })

    return {
        restart: async (): Promise<void> => {
            await kill()
            current = await spawn()
        },
        kill,
    }
}
