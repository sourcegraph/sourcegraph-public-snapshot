/* eslint-disable @typescript-eslint/explicit-module-boundary-types */

const createLogger = (prefix: string, level: 'info' | 'debug' | 'log' | 'error') => (...args: any): void =>
    console[level](prefix, ...args)

export const logger = {
    info: createLogger('[Sourcegraph]', 'log'),
    debug: createLogger('[Sourcegraph Debug]', 'log'),
    error: createLogger('[Sourcegraph]', 'error'),
}
