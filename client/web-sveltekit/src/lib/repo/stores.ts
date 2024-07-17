import { memoize } from 'lodash'
import { writable, type Writable } from 'svelte/store'

import { createLocalWritable } from '$lib/stores'
import { createEmptySingleSelectTreeState, type TreeState } from '$lib/TreeView'
import type { Panel } from '$lib/wildcard'

/**
 * Persistent, global state for the file sidebar. By keeping the state in memory we can
 * properly restore the UI when the user closes/opens the sidebar or navigates up the repo.
 */
export const getSidebarFileTreeStateForRepo = memoize(
    (_repoName: string): Writable<TreeState> => writable<TreeState>(createEmptySingleSelectTreeState()),
    repoName => repoName
)

export const rightSidePanelOpen = createLocalWritable<boolean>('repo.right-panel.open', false)
export const fileTreeSidePanel = writable<Panel | null>(null)
