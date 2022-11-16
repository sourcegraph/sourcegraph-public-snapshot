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

declare global {
    namespace jest {
        interface Matchers<R, T> {
            toBeAriaEnabled(): R
            toBeAriaDisabled(): R
        }
    }
}

export {}
