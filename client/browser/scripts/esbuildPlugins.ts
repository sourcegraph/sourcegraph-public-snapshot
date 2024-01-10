import type * as esbuild from 'esbuild'
import shelljs from 'shelljs'
import signale from 'signale'

import { buildNpm } from './build-npm'
import * as tasks from './tasks'

export const copyAssetsPlugin: esbuild.Plugin = {
    name: 'copyAssets',
    setup: (): void => {
        tasks.copyAssets()
    },
}

export function browserBuildStepsPlugin(mode: 'dev' | 'prod'): esbuild.Plugin {
    return {
        name: 'browserBuildSteps',
        setup: (build: esbuild.PluginBuild): void => {
            build.onEnd(async result => {
                if (result.errors.length === 0) {
                    signale.info('Running browser build steps...')
                    tasks.buildChrome(mode)()
                    tasks.buildFirefox(mode)()
                    tasks.buildEdge(mode)()
                    if (mode === 'prod') {
                        if (isXcodeAvailable()) {
                            tasks.buildSafari(mode)()
                        } else {
                            signale.debug(
                                'Skipping Safari build because Xcode tools were not found (xcrun, xcodebuild)'
                            )
                        }
                    }
                    tasks.copyIntegrationAssets()
                    if (mode === 'prod') {
                        await buildNpm()
                    }
                }
            })
        },
    }
}

function isXcodeAvailable(): boolean {
    return !!shelljs.which('xcrun') && !!shelljs.which('xcodebuild')
}
