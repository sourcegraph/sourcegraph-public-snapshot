import * as esbuild from 'esbuild'

import { esbuildBuildOptions } from '../config/esbuild'

async function build(): Promise<void> {
    await esbuild.build(esbuildBuildOptions('prod'))
}

if (require.main === module) {
    build().catch(error => {
        console.error('Error:', error)
        process.exit(1)
    })
}
