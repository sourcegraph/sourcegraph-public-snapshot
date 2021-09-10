import * as esbuild from 'esbuild'

/**
 * An esbuild plugin to omit a package from the bundle (for example, to omit 'monaco-editor' by
 * treating it as an empty file).
 */
export const omitPackagePlugin = (modules: string[]): esbuild.Plugin => ({
    name: 'omitPackage',
    setup: build => {
        const filter = new RegExp(`^(${modules.join('|')})$`)
        build.onResolve({ filter, namespace: 'file' }, () => ({ path: '-', namespace: 'omit' }))
        build.onLoad({ filter: /./, namespace: 'omit' }, () => ({ contents: '', loader: 'js' }))
    },
})
