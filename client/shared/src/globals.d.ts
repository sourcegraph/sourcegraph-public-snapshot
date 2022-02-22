declare module '*.scss' {
    const cssModule: string
    export default cssModule
}
declare module '*.css' {
    const cssModule: string
    export default cssModule
}

/**
 * For Web Worker entrypoints using Webpack's worker-loader.
 *
 * See https://github.com/webpack-contrib/worker-loader#integrating-with-typescript.
 */
declare module '*.worker.ts' {
    class WebpackWorker extends Worker {
        constructor()
    }
    export default WebpackWorker
}

/**
 * Set by jest-environment-jsdom
 */
declare var jsdom: import('jsdom').JSDOM
