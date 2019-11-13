/**
 * Identity-function helper to ensure a value `T` is a subtype of `U`.
 *
 * @template U The type to check for (explicitely specify this)
 * @template T The actual type (inferred, don't specify this)
 */
export const subTypeOf = <U>() => <T extends U>(value: T): T => value

/**
 * Returns true if `val` is not `null` or `undefined`
 */
export const isDefined = <T>(val: T): val is NonNullable<T> => val !== undefined && val !== null

/**
 * Negates a type guard.
 * Returns a function that returns `true` when the input is **not** the type checked for by the given type guard.
 * It therefor excludes a type from a union type.
 *
 * @param isType The type guard that checks whether the given input value should be excluded.
 */
export const isNot = <TInput, TExclude extends TInput>(isType: (val: TInput) => val is TExclude) => (
    value: TInput
): value is Exclude<TInput, TExclude> => !isType(value)

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
