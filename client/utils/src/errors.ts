/**
 * Run the passed function and return `undefined` if it throws an error.
 */
export function tryCatch<T>(function_: () => T): T | undefined {
    try {
        return function_()
    } catch {
        return undefined
    }
}
