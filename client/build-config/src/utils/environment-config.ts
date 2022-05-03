// Read the environment variable from `process.env` and cast it to `Boolean`.
export function getEnvironmentBoolean(name: string): boolean {
    const variable = process.env[name]

    return Boolean(variable && JSON.parse(variable))
}
