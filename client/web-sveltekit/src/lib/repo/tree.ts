interface Entry {
    isDirectory: boolean
    name: string
}

export function findReadme<T extends Entry>(entries: T[]): T | null {
    return entries.find(entry => !entry.isDirectory && /^readme($|\.)/i.test(entry.name)) ?? null
}
