import { Falsey } from 'utility-types'

/**
 * Returns true if `val` is not `null` or `undefined`
 */
export const isDefined = <T>(value: T): value is NonNullable<T> => value !== undefined && value !== null

/**
 * Returns true if `val` is truthy.
 */
export const isTruthy = <T>(value: T | Falsey): value is T => !!value
