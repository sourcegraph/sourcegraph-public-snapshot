declare module '*.scss' {
    const cssModule: string
    export default cssModule
}
declare module '*.css' {
    const cssModule: string
    export default cssModule
}

/**
 * Set by jest-environment-jsdom
 */
declare var jsdom: import('jsdom').JSDOM
