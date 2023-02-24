import type { SidebarFilter } from './utils'

export const searchTypes: SidebarFilter[] = [
    {
        label: 'Search repos by org or name',
        value: 'repo:',
        kind: 'utility',
    },
    {
        label: 'Find a symbol',
        value: 'type:symbol',
        kind: 'utility',
        runImmediately: true,
    },
    {
        label: 'Search diffs',
        value: 'type:diff',
        kind: 'utility',
        runImmediately: true,
    },
    {
        label: 'Search commit message',
        value: 'type:commit',
        kind: 'utility',
    },
]
