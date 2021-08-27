import path from 'path'

import * as esbuild from 'esbuild'

const nodeModule = (name: string): string =>
    require.resolve(name, { paths: [path.join(__dirname, '../../../../node_modules')] })

const RESOLUTIONS: Record<string, string> = {
    path: nodeModule('path-browserify'),
}

/**
 * An esbuild plugin to resolve imports of 'path' to 'path-browserify' to run in the browser.
 */
export const packageResolutionPlugin: esbuild.Plugin = {
    name: 'packageResolution',
    setup: build => {
        const filter = new RegExp(`^(${Object.keys(RESOLUTIONS).join('|')})$`)
        build.onResolve({ filter, namespace: 'file' }, args => {
            if (args.kind === 'import-statement' || args.kind === 'require-call') {
                const resolution = RESOLUTIONS[args.path]
                if (resolution) {
                    return {
                        path: resolution,
                    }
                }
            }
        })
    },
}
