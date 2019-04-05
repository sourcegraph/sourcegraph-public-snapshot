export function illegalArgument(name?: string): Error {
    if (name) {
        return new Error(`Illegal argument: ${name}`)
    }
    return new Error('Illegal argument')
}
