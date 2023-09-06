import type { TreeEntryFields } from './api/tree'

export function findReadme(entries: TreeEntryFields[]): TreeEntryFields | null {
    return entries.find(entry => !entry.isDirectory && /^readme($|\.)/i.test(entry.name)) ?? null
}
