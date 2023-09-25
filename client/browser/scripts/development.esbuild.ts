import * as esbuild from 'esbuild'
import signale from 'signale'

import { esbuildBuildOptions } from '../config/esbuild'

async function watch(): Promise<void> {
    const ctx = await esbuild.context(esbuildBuildOptions('dev'))
    await ctx.watch()
    signale.info('Watching...')
    await new Promise(() => {}) // wait forever
}

if (require.main === module) {
    watch().catch(error => {
        console.error('Error:', error)
        process.exit(1)
    })
}
