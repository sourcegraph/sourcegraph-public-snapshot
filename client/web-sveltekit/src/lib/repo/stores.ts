import { memoize } from 'lodash'
import { writable, type Writable } from 'svelte/store'

import { createEmptySingleSelectTreeState, type TreeState } from '$lib/TreeView'

/**
 * Persistent, global state for the file sidebar. By keeping the state in memory we can
 * properly restore the UI when the user closes/opens the sidebar or navigates up the repo.
 */
export const getSidebarFileTreeStateForRepo = memoize(
    (_repoName: string): Writable<TreeState> => writable<TreeState>(createEmptySingleSelectTreeState()),
    repoName => repoName
)
