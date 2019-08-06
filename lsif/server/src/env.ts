/**
 * Reads an integer from an environment variable or defaults to the given value.
 */
export function readEnvInt({ key, defaultValue }: { key: string; defaultValue: number }): number {
    return (process.env[key] && parseInt(process.env[key] || '', 10)) || defaultValue
}
