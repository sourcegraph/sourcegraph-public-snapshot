/**
 * Asserts that the given condition is trueish.
 *
 * @param expectedCondition The condition to check.
 * @param message The error message to use if the condition is `false`.
 * @throws An error with the given message if the condition is `false`.
 */
export function assert(expectedCondition: any, message: string): asserts expectedCondition {
    if (!expectedCondition) {
        console.error(message)

        throw new Error(message)
    }
}

/**
 * Asserts that the given value is not `null` or `undefined`.
 *
 * @param value The value to check.
 * @param message The error message to use if the value is `null` or `undefined`.
 * @throws An error with the given message if the value is `null` or `undefined`.
 **/
export function assertNonNullable<T>(value: T | null | undefined, message: string): asserts value is T {
    if (value === null || value === undefined) {
        throw new Error(message)
    }
}
