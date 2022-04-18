import * as esbuild from 'esbuild'

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
        build.onResolve({ filter, namespace: 'file' }, args =>
            (args.kind === 'import-statement' || args.kind === 'require-call') && resolutions[args.path]
                ? { path: resolutions[args.path] }
                : undefined
        )
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
