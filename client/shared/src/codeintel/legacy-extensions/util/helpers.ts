/**
 * Returns true if the given value is not undefined.
 *
 * @param value The value to test.
 */
export function isDefined<T>(value: T | undefined): value is T {
    return value !== undefined
}

/**
 * Returns true if the value is defined and, if an array, contains at least
 * one element.
 *
 * @param value The value to test.
 */
export function nonEmpty<T>(value: T | T[] | null | undefined): value is T | T[] {
    return !!value && !(Array.isArray(value) && value.length === 0)
}

/**
 * Ensure that the given value is an array.
 *
 * @param value The list of values, a single value, or null.
 */
export function asArray<T>(value: T | T[] | null): T[] {
    return Array.isArray(value) ? value : value ? [value] : []
}

/**
 * Apply a map function on a single value or over a list of values. Returns the
 * modified result in the same shape as the input.
 *
 * @param value The list of values, a single value, or null.
 * @param func The map function.
 */
export function mapArrayish<T, R>(value: T | T[] | null, func: (value: T) => R): R | R[] | null {
    return Array.isArray(value) ? value.map(func) : value ? func(value) : null
}

/**
 * Removes duplicates and sorts the given list.
 *
 * @param values The input values.
 */
export function sortUnique<T>(values: T[]): T[] {
    const sorted = Array.from(new Set(values))
    sorted.sort()
    return sorted
}

/**
 * Constructs a function that returns true if the input is not in the excludelist.
 *
 * @param excludelist The excludelist.
 */
export function notIn<T>(excludelist: T[]): (value: T) => boolean {
    return (value: T): boolean => !excludelist.includes(value)
}

/**
 * Converts a promise returning a value into a promise returning a value or undefined.
 * Catches any errors that occur during invocation and returns undefined instead of
 * rejecting the promise.
 *
 * @param promise The promise.
 */
export function safePromise<P, R>(promise: (argument: P) => Promise<R>): (argument: P) => Promise<R | undefined> {
    return async (argument: P): Promise<R | undefined> => {
        try {
            return await promise(argument)
        } catch {
            return undefined
        }
    }
}
