/**
 * Interface for providing the tree data.
 */
export interface TreeProvider<T> {
    /**
     * Returns the node values (e.g. the initial set of nodes for the root or child nodes for a node).
     */
    getEntries(): T[]
    /**
     * Whether or not the provided entrty is has (possibly) children or not.
     */
    isExpandable(entry: T): boolean
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

export interface NodeState {
    expanded: boolean
    selected: boolean
}

export interface TreeState {
    focused: string
    nodes: Record<string, NodeState>
}

/**
 * Helper function for augmenting a node's state.
 */
export function updateNodeState(
    treeState: TreeState,
    nodeID: string,
    state: Partial<NodeState>
): Record<string, NodeState> {
    return {
        ...treeState.nodes,
        [nodeID]: { ...treeState.nodes[nodeID], ...state },
    }
}

export class DummyTreeProvider implements TreeProvider<any> {
    isExpandable(_entry: any): boolean {
        return false
    }
    getEntries(): any[] {
        return []
    }
    fetchChildren(_entry: any): Promise<TreeProvider<any>> {
        return Promise.resolve(new DummyTreeProvider())
    }
    getNodeID(_entry: any): string {
        return ''
    }
}
