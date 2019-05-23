/**
 * For Web Worker entrypoints using Webpack's worker-loader.
 *
 * See https://github.com/webpack-contrib/worker-loader#integrating-with-typescript.
 */
declare module 'worker-loader?inline*' {
    class WebpackWorker extends Worker {
        constructor()
    }
    export default WebpackWorker
}
