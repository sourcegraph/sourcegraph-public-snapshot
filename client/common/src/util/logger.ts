/* eslint-disable no-console */
export interface Logger {
    log: Console['log']
    warn: Console['warn']
    error: Console['error']
    info: Console['info']
    debug: Console['debug']
}

type LoggerArgsType = Parameters<Logger['log']>

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
