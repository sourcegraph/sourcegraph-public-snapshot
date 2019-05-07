/**
 * Returns true if `val` is not `null` or `undefined`
 */
export const isDefined = <T>(val: T): val is NonNullable<T> => val !== undefined && val !== null

/**
 * Returns a function that returns `true` if the given `key` of the object is not `null` or `undefined`.
 *
 * I ❤️ TypeScript.
 */
export const propertyIsDefined = <T extends object, K extends keyof T>(key: K) => (
    val: T
): val is T & { [k in K]-?: NonNullable<T[k]> } => isDefined(val[key])

/**
 * Returns a function that returns `true` if the given value is an instance of the given class.
 *
 * @param of A reference to a class, e.g. `HTMLElement`
 */
export const isInstanceOf = <C extends new () => object>(of: C) => (val: unknown): val is InstanceType<C> =>
    val instanceof of
