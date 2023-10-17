export function isDevEnvironment(): boolean {
    return process.env.NODE_ENV === 'development'
}
