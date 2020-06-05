/**
 * For Web Worker entrypoints using Webpack's worker-loader.
 *
 * See https://github.com/webpack-contrib/worker-loader#integrating-with-typescript.
 */
declare module 'worker-loader?*' {
    class WebpackWorker extends Worker {
        constructor()
    }
    export default WebpackWorker
}

declare module '*.scss' {
    const cssModule: string
    export default cssModule
}
declare module '*.css' {
    const cssModule: string
    export default cssModule
}

/**
 * Set by shared/dev/jest-environment.js
 */
declare var jsdom: import('jsdom').JSDOM
