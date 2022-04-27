/**
 * Returns true if the given value is not undefined.
 *
 * @param value The value to test.
 */
export function isDefined<T>(value: T | undefined): value is T {
    return value !== undefined
}
