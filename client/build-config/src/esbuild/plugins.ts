import type * as esbuild from 'esbuild'

export * from './packageResolutionPlugin'
export * from './stylePlugin'
export * from './workerPlugin'

export const buildTimerPlugin: esbuild.Plugin = {
    name: 'buildTimer',
    setup: (build: esbuild.PluginBuild): void => {
        let buildStarted: number
        build.onStart(() => {
            buildStarted = Date.now()
        })
        build.onEnd(() => console.log(`# esbuild: build took ${Date.now() - buildStarted}ms`))
    },
}
