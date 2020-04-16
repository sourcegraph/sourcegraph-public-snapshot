/**
 * Identity-function helper to ensure a value `T` is a subtype of `U`.
 *
 * @template U The type to check for (explicitly specify this)
 * @template T The actual type (inferred, don't specify this)
 */
export const subTypeOf = <U>() => <T extends U>(value: T): T => value

/**
 * Returns true if `val` is not `null` or `undefined`
 */
export const isDefined = <T>(val: T): val is NonNullable<T> => val !== undefined && val !== null

/**
 * Returns a type guard that checks whether the given value is strictly equal to a specific value.
 * This can for example be used with `isNot()` to exclude string literals like `"loading"`.
 *
 * @param constant The value to compare to.
 */
export const isExactly = <T, C extends T>(constant: C) => (value: T): value is C => value === constant

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
 * Returns a function that returns `true` if the given `key` of the object passes the given type guard.
 *
 * @param key The key of the property to check.
 */
export const hasProperty = <O extends object, K extends string | number | symbol>(key: K) => (
    object: O
): object is O & { [k in K]: unknown } => key in object

/**
 * Returns a function that returns `true` if the given `key` of the object passes the given type guard.
 *
 * @param key The key of the property to check.
 * @param isType The type guard to evalute on the property value.
 */
export const property = <O extends object, K extends keyof O, T extends O[K]>(
    key: K,
    isType: (value: O[K]) => value is T
) => (object: O): object is O & { [k in K]: T } => isType(object[key])

/**
 * Returns a function that returns `true` if the given value is an instance of the given class.
 *
 * @param of A reference to a class, e.g. `HTMLElement`
 */
export const isInstanceOf = <C extends new () => object>(of: C) => (val: unknown): val is InstanceType<C> =>
    val instanceof of
