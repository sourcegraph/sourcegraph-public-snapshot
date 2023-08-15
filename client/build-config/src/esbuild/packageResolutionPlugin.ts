import fs from 'fs'

import { CachedInputFileSystem, ResolverFactory } from 'enhanced-resolve'
import type * as esbuild from 'esbuild'

import { NODE_MODULES_PATH, WORKSPACE_NODE_MODULES_PATHS } from '../paths'

interface Resolutions {
    [fromModule: string]: string
}

/**
 * An esbuild plugin to redirect imports from one package to another (for example, from 'path' to
 * 'path-browserify' to run in the browser).
 */
export const packageResolutionPlugin = (resolutions: Resolutions): esbuild.Plugin => ({
    name: 'packageResolution',
    setup: build => {
        const filter = new RegExp(`^(${Object.keys(resolutions).join('|')})$`)

        const resolver = ResolverFactory.createResolver({
            fileSystem: new CachedInputFileSystem(fs, 4000),
            extensions: ['.ts', '.tsx', '.js', '.jsx', '.json'],
            symlinks: true, // Resolve workspace symlinks
            modules: [NODE_MODULES_PATH, ...WORKSPACE_NODE_MODULES_PATHS],
            unsafeCache: true,
        })

        build.onResolve({ filter, namespace: 'file' }, async args => {
            if (
                (args.kind === 'import-statement' || args.kind === 'require-call' || args.kind === 'dynamic-import') &&
                resolutions[args.path]
            ) {
                if (resolutions[args.path] === '/dev/null') {
                    return { namespace: 'devnull', path: '/dev/null' }
                }
                const resolvedPath = await new Promise<string>((resolve, reject) => {
                    resolver.resolve({}, args.resolveDir, resolutions[args.path], {}, (error, filepath) => {
                        if (filepath) {
                            resolve(filepath)
                        } else {
                            reject(error ?? new Error(`Could not resolve file path for ${resolutions[args.path]}`))
                        }
                    })
                })
                return { path: resolvedPath }
            }
            return undefined
        })

        build.onLoad({ filter: new RegExp(''), namespace: 'devnull' }, () => ({ contents: '' }))
    },
})

export const RXJS_RESOLUTIONS: Resolutions = {
    // Needed because imports of rxjs/internal/... actually import a different variant of
    // rxjs in the same package, which leads to observables from combineLatestOrDefault (and
    // other places that use rxjs/internal/...) not being cross-compatible. See
    // https://stackoverflow.com/questions/53758889/rxjs-subscribeto-js-observable-check-works-in-chrome-but-fails-in-chrome-incogn.
    'rxjs/internal/OuterSubscriber': require.resolve('rxjs/_esm5/internal/OuterSubscriber'),
    'rxjs/internal/util/subscribeToResult': require.resolve('rxjs/_esm5/internal/util/subscribeToResult'),
    'rxjs/internal/util/subscribeToArray': require.resolve('rxjs/_esm5/internal/util/subscribeToArray'),
    'rxjs/internal/Observable': require.resolve('rxjs/_esm5/internal/Observable'),
}
