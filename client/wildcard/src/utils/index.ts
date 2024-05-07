export { screenReaderAnnounce, screenReaderClearAnnouncements } from './screenReaderAnnounce'
export { createLinkClickHandler } from './linkClickHandler'
export { joinWithAnd } from './joinWithAnd'
export * from './markdown'
export * from './contributions'

/**
 * Returns true if `val` is not `null` or `undefined`
 */
export const isDefined = <T>(value: T): value is NonNullable<T> => value !== undefined && value !== null

const AGGREGATE_ERROR_NAME = 'AggregateError'

/**
 * DEPRECATED: use dataOrThrowErrors instead
 * Creates an aggregate error out of multiple provided error likes
 *
 * @param errors The errors or ErrorLikes to aggregate
 */
export const createAggregateError = (errors: readonly ErrorLike[] = []): Error =>
    errors.length === 1
        ? asError(errors[0])
        : Object.assign(new Error(errors.map(error => error.message).join('\n')), {
              name: AGGREGATE_ERROR_NAME,
              errors: errors.map(asError),
          })

/**
 * Converts an ErrorLike to a proper Error if needed, copying all properties
 *
 * @param value An Error, object with ErrorLike properties, or other value.
 */
export const asError = (value: unknown): Error => {
    if (value instanceof Error) {
        return value
    }
    if (isErrorLike(value)) {
        return Object.assign(new Error(value.message), value)
    }
    return new Error(String(value))
}

const isErrorLike = (value: unknown): value is ErrorLike =>
    typeof value === 'object' && value !== null && ('stack' in value || 'message' in value) && !('__typename' in value)

export function getBrowserName(): 'chrome' | 'safari' | 'firefox' | 'other' {
    return isChrome() ? 'chrome' : isSafari() ? 'safari' : isFirefox() ? 'firefox' : 'other'
}

function isChrome(): boolean {
    return typeof window !== 'undefined' && !!window.navigator.userAgent.match(/chrome|chromium|crios/i)
}

function isSafari(): boolean {
    return typeof window !== 'undefined' && !!window.navigator.userAgent.match(/safari/i) && !isChrome()
}

function isFirefox(): boolean {
    return typeof window !== 'undefined' && window.navigator.userAgent.includes('Firefox')
}

/* eslint-disable no-console */
interface Logger {
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

interface ErrorLike {
    message: string
    name?: string
}
