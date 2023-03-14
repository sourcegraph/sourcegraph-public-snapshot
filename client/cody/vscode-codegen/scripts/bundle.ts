import childProcess from 'child_process'
import { existsSync } from 'fs'
import path from 'path'

import * as esbuild from 'esbuild'

const minify = process.env.NODE_ENV === 'production'
const outdir = path.join(__dirname, '../dist')

const SHARED_CONFIG: Required<Pick<esbuild.BuildOptions, 'outdir' | 'minify' | 'sourcemap'>> = {
    outdir,
    minify,
    sourcemap: true,
}

async function build(): Promise<void> {
    if (existsSync(outdir)) {
        // eslint-disable-next-line no-sync
        childProcess.execFileSync('rm', ['-rf', outdir], { stdio: 'inherit' })
    }

    const ctx = await esbuild.context({
        entryPoints: { extension: path.join(__dirname, '/../src/extension.ts') },
        bundle: true,
        format: 'cjs',
        platform: 'node',
        external: ['vscode'],
        /// / TODO(sqs) banner: { js: 'global.Buffer = require("buffer").Buffer' },
        ...SHARED_CONFIG,
        outdir: path.join(SHARED_CONFIG.outdir, 'node'),
    })

    await ctx.rebuild()

    if (process.env.WATCH) {
        await ctx.watch()
    }

    await ctx.dispose()
}

if (require.main === module) {
    build()
        .catch(error => {
            console.error('Error:', error)
            process.exit(1)
        })
        .finally(() => process.exit(0))
}
