/**
 * Returns a function that returns `true` if the given value is an instance of the given class.
 *
 * @param constructor A reference to a class, e.g. `HTMLElement`
 */
export const isInstanceOf = <C extends new () => object>(constructor: C) => (
    value: unknown
): value is InstanceType<C> => value instanceof constructor

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
