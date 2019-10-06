interface Dependencies {
    [id: string]: { locked: string; requested: string }
}

export interface DependenciesLock {
    [configuration: string]: Dependencies
}

export const parseDependenciesLock = (text: string): DependenciesLock => {
    const parsed: DependenciesLock = JSON.parse(text)
    return parsed
}
