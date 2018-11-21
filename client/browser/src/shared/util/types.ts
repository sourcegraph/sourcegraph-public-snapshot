/**
 * Returns true if `val` { is } not `null` or `undefined`
 */
const isDefined = <T>(val: T): val is NonNullable<T> => val !== undefined && val !== null

/**
 * Returns a function that returns `true` if the given `key` of the object is not `null` or `undefined`.
 *
 * I ❤️ TypeScript.
 */
export const propertyIsDefined = <T extends object, K extends keyof T>(key: K) => (
    val: T
): val is K extends any ? ({ [k in Exclude<keyof T, K>]: T[k] } & { [k in K]: NonNullable<T[k]> }) : never =>
    isDefined(val[key])
