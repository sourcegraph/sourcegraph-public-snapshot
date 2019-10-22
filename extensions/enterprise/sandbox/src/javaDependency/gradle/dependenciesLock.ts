interface Dependencies {
    [id: string]: { locked: string; requested: string }
}

export interface DependenciesLock {
    [configuration: string]: Dependencies
}

export const parseDependenciesLock = (text: string, logLabel: string): DependenciesLock => {
    try {
        const parsed: DependenciesLock = JSON.parse(text)
        return parsed
    } catch (err) {
        console.error(`Error parsing dependencies.lock (${logLabel})`)
        return {}
    }
}
