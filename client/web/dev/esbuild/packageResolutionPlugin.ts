import * as esbuild from 'esbuild'

/**
 * An esbuild plugin to redirect imports from one package to another (for example, from 'path' to
 * 'path-browserify' to run in the browser).
 */
export const packageResolutionPlugin = (resolutions: { [fromModule: string]: string }): esbuild.Plugin => ({
    name: 'packageResolution',
    setup: build => {
        const filter = new RegExp(`^(${Object.keys(resolutions).join('|')})$`)
        build.onResolve({ filter, namespace: 'file' }, args =>
            (args.kind === 'import-statement' || args.kind === 'require-call') && resolutions[args.path]
                ? { path: resolutions[args.path] }
                : undefined
        )
    },
})
