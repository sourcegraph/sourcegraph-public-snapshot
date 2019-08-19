/**
 * Reads an integer from an environment variable or defaults to the given value.
 *
 * @param key The environment variable name.
 * @param defaultValue The default value.
 */
export function readEnvInt(key: string, defaultValue: number): number {
    return (process.env[key] && parseInt(process.env[key] || '', 10)) || defaultValue
}

/**
 * Reads a string from an environment variable or defaults to the given value.
 *
 * @param key The environment variable name.
 * @param defaultValue The default value.
 */
export function readEnv(key: string, defaultValue: string): string {
    return process.env[key] || defaultValue
}

/**
 * Determine if an exception value has the given error code.
 *
 * @param e The exception value.
 * @param expectedCode The expected error code.
 */
export function hasErrorCode(e: any, expectedCode: string): boolean {
    return 'code' in e && e.code === expectedCode
}
