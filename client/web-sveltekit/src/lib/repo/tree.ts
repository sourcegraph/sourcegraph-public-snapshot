import type { TreeEntryFields } from '$lib/graphql-operations'

export function findReadme(entries: TreeEntryFields[]): TreeEntryFields | null {
    return entries.find(entry => !entry.isDirectory && /^readme($|\.)/i.test(entry.name)) ?? null
}
