import esbuild from 'esbuild'

/**
 * An esbuild plugin that omits non-module .css and .scss stylesheets. This is useful when building
 * for server-side prerendering, where CSS stylesheets won't be used in the output and can be
 * skipped during the build.
 */
export const omitNonModuleStylePlugin: esbuild.Plugin = {
    name: 'omitNonModuleStyle',
    setup: build => {
        build.onResolve({ filter: /\.s?css$/, namespace: 'file' }, args => {
            const isCSSModule = args.path.endsWith('.module.scss') || args.path.endsWith('.module.css')
            return isCSSModule ? undefined : { path: '-', namespace: 'omit-style', sideEffects: false }
        })
        build.onLoad({ filter: /./, namespace: 'omit-style' }, args => ({
            contents: '',
            loader: 'css',
        }))
    },
}
