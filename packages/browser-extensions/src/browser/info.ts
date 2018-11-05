const GLOBAL = global as any

export const isFirefox = typeof GLOBAL.browser !== 'undefined'
