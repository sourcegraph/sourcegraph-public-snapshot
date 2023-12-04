declare module '*.scss' {
    const cssModule: string
    export default cssModule
}
declare module '*.css' {
    const cssModule: string
    export default cssModule
}

/**
 * For Web Worker entrypoints using workerPlugin.
 */
declare module '*.worker.ts' {
    class _Worker extends Worker {
        constructor()
    }
    export default _Worker
}

interface Window {
    context?: any
}
