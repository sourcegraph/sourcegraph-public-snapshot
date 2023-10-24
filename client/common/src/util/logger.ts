/* eslint-disable no-console */
export interface Logger {
    log: Console['log']
    warn: Console['warn']
    error: Console['error']
    info: Console['info']
    debug: Console['debug']
}

type LoggerArgsType = Parameters<Logger['log']>

/**
 * The logger utility to be used instead of `console.*` methods in browser environments to:
 * 1. Avoid leaving `console.*` added for debugging purposes during development.
 * 2. Have more control over logging pipelines. E.g.,
 *    - Forward errors to the error monitoring services.
 *    - Dynamically change logging level depending on the environment.
 *
 * Check out the Unified logger service RFC for more context:
 * https://docs.google.com/document/d/1PolGRDS9XfKj-IJEBi7BTbVZeUsQfM-3qpjLsLlB-yw/edit
 */
export const logger: Logger = {
    log: (...args: LoggerArgsType) => {
        console.log(...args)
    },
    error: (...args: LoggerArgsType) => {
        console.error(...args)
    },
    warn: (...args: LoggerArgsType) => {
        console.warn(...args)
    },
    info: (...args: LoggerArgsType) => {
        console.info(...args)
    },
    debug: (...args: LoggerArgsType) => {
        console.debug(...args)
    },
}
