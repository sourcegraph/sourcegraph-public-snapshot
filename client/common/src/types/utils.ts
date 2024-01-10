import type { Falsey } from 'utility-types'

/**
 * Returns true if `val` is not `null` or `undefined`
 */
export const isDefined = <T>(value: T): value is NonNullable<T> => value !== undefined && value !== null

/**
 * Returns true if `val` is truthy.
 */
export const isTruthy = <T>(value: T | Falsey): value is T => !!value

/**
 * Removes `null` and `undefined` values from a list of values
 */
export const defined = <T>(values: (T | null | undefined)[]): T[] => values.filter(isDefined)
