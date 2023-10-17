import * as esbuild from 'esbuild'

import { esbuildBuildOptions } from '../config/esbuild'

import { browserBuildStepsPlugin, copyAssetsPlugin } from './esbuildPlugins'

async function build(): Promise<void> {
    await esbuild.build(esbuildBuildOptions('prod', [copyAssetsPlugin, browserBuildStepsPlugin('prod')]))
}

if (require.main === module) {
    build().catch(error => {
        console.error('Error:', error)
        process.exit(1)
    })
}
