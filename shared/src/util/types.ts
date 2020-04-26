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

/**
 * Combines multiple type guards into one type guard that checks if the value passes any of the provided type guards.
 */
export function anyOf<T0, T1 extends T0, T2 extends Exclude<T0, T1>>(
    t1: (value: T0) => value is T1,
    t2: (value: Exclude<T0, T1>) => value is T2
): (value: T0) => value is T1 | T2
export function anyOf<T0, T1 extends T0, T2 extends Exclude<T0, T1>, T3 extends Exclude<T1, T2>>(
    t1: (value: T0) => value is T1,
    t2: (value: Exclude<T0, T1>) => value is T2,
    t3: (value: Exclude<T1, T2>) => value is T3
): (value: T0) => value is T1 | T2 | T3
export function anyOf<
    T0,
    T1 extends T0,
    T2 extends Exclude<T0, T1>,
    T3 extends Exclude<T1, T2>,
    T4 extends Exclude<T2, T3>
>(
    t1: (value: T0) => value is T1,
    t2: (value: Exclude<T0, T1>) => value is T2,
    t3: (value: Exclude<T1, T2>) => value is T3,
    t4: (value: Exclude<T2, T3>) => value is T4
): (value: T0) => value is T1 | T2 | T3 | T4
export function anyOf(...typeGuards: any[]): any {
    return (value: unknown) => typeGuards.some((guard: (value: unknown) => boolean) => guard(value))
}
