export function assertNonNullable<T>(value: T | null | undefined, message: string): asserts value is T {
    if (value === null || value === undefined) {
        throw new Error(message)
    }
}
