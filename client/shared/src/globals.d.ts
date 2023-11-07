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

/**
 * Set by shared/dev/jest-environment.js
 */
declare var jsdom: import('jsdom').JSDOM

interface Window {
    context?: any
}
