/**
 * Reads an integer from an environment variable or defaults to the given value.
 */
export function readEnvInt({ key, defaultValue }: { key: string; defaultValue: number }): number {
    const value = process.env[key]
    if (!value) {
        return defaultValue
    }

    const n = parseInt(value, 10)
    if (isNaN(n)) {
        return defaultValue
    }

    return n
}
