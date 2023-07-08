import { writable } from 'svelte/store'

export interface TreeProvider<T> {
    getEntries(): T[]
    isExpandable(entry: T): boolean
    fetchChildren(entry: T): Promise<TreeProvider<T>>
    getKey(entry: T): string
}

export interface NodeState {
    expanded: boolean
    selected: boolean
}

export interface TreeState {
    focused: string
    nodes: Record<string, NodeState>
}

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
        return Promise.resolve(new this.constructor())
    }
    getKey(_entry: any): string {
        return ''
    }
}
