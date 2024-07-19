/**
 * Interface for providing the tree data.
 */
export interface TreeProvider<T> {
    /**
     * Returns the node values (e.g. the initial set of nodes for the root or child nodes for a node).
     */
    getEntries(): T[]
    /**
     * Whether or not the provided entry has (possibly) children or not.
     */
    isExpandable(entry: T): boolean
    /**
     * Whether or not the provided entry can be selected.
     */
    isSelectable(entry: T): boolean
    /**
     * Called when the corresponding node id expanded.
     */
    fetchChildren(entry: T): Promise<TreeProvider<T>>
    /**
     * Returns a (tree-wide) unique ID for the provided value. The tree view uses this
     * to track various state (open/closed, selected, ...).
     */
    getNodeID(entry: T): string
}

export interface SingleSelectTreeState {
    focused: string
    selected: string
    expandedNodes: Set<string>
    disableScope: boolean
}

export type TreeState = SingleSelectTreeState

export function createEmptySingleSelectTreeState(): SingleSelectTreeState {
    return {
        focused: '',
        selected: '',
        expandedNodes: new Set(),
        disableScope: false,
    }
}

export enum TreeStateUpdate {
    EXPAND = 2 ** 0,
    FOCUS = 2 ** 1,
    COLLAPSE = 2 ** 2,
    SELECT = 2 ** 3,
    COLLAPSEANDFOCUS = COLLAPSE | FOCUS,
    EXPANDANDFOCUS = EXPAND | FOCUS,
}

export function updateTreeState<T extends TreeState>(state: T, nodeID: string, update: TreeStateUpdate): T {
    if (update & TreeStateUpdate.FOCUS) {
        if (state.focused !== nodeID) {
            state = { ...state, focused: nodeID }
        }
    }
    if (update & TreeStateUpdate.SELECT) {
        if (state.selected !== nodeID) {
            state = { ...state, selected: nodeID }
        }
    }
    if (update & TreeStateUpdate.EXPAND) {
        if (!state.expandedNodes.has(nodeID)) {
            const expandedNodes = new Set(state.expandedNodes)
            expandedNodes.add(nodeID)
            state = { ...state, expandedNodes }
        }
    }
    if (update & TreeStateUpdate.COLLAPSE) {
        if (state.expandedNodes.has(nodeID)) {
            const expandedNodes = new Set(state.expandedNodes)
            expandedNodes.delete(nodeID)
            state = { ...state, expandedNodes }
        }
    }

    return state
}
