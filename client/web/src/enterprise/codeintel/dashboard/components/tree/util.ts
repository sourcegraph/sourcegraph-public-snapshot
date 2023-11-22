import { isDefined } from '@sourcegraph/common'
import type { TreeNode as WTreeNode } from '@sourcegraph/wildcard'

import type { PreciseIndexFields } from '../../../../../graphql-operations'

// Strip leading/trailing slashes and add a single leading slash
export function sanitizePath(root: string): string {
    return `/${root.replaceAll(/(^\/+)|(\/+$)/g, '')}`
}

// Group values together based on the given function
export function groupBy<V, K>(values: V[], keyFn: (value: V) => K): Map<K, V[]> {
    return values.reduce(
        (acc, val) => acc.set(keyFn(val), (acc.get(keyFn(val)) || []).concat([val])),
        new Map<K, V[]>()
    )
}

// Compare two flattened Map<string, T> entries by key.
export function byKey<T>(tup1: [string, T], tup2: [string, T]): number {
    return tup1[0].localeCompare(tup2[0])
}

// Return the list of keys for the associated values for which the given predicate returned true.
export function keysMatchingPredicate<K, V>(map: Map<K, V>, predFn: (value: V) => boolean): K[] {
    return [...map.entries()].map(([key, value]) => (predFn(value) ? key : undefined)).filter(isDefined)
}

// Return true if the given slices for a proper (pairwise) subset < superset relation
export function checkSubset(subset: string[], superset: string[]): boolean {
    return subset.length < superset.length && subset.filter((value, index) => value !== superset[index]).length === 0
}

export function getIndexerKey(index: PreciseIndexFields): string {
    return index.indexer?.key || index.inputIndexer
}

export function getIndexRoot(index: PreciseIndexFields): string {
    return sanitizePath(index.projectRoot?.path || index.inputRoot)
}

export interface TreeNodeWithDisplayName extends WTreeNode {
    displayName: string
}

// Constructs an outline suitable for use with the wildcard <Tree /> component. This function constructs
// a file tree outline with a dummy root node (un-rendered) so that we can display explicit data for the
// root directory. We also attempt to collapse any runs of directories that have no data of its own to
// display and only one child.
export function buildTreeData(dataPaths: Set<string>): TreeNodeWithDisplayName[] {
    // Construct a list of paths reachable from the given input paths by sanitizing the input path,
    // exploding the resulting path list into directory segments, constructing all prefixes of the
    // resulting segments, and deduplicating and sorting the result. This gives all ancestor paths
    // of the original input paths in lexicographic (NOTE: topological) order.
    const allPaths = [
        ...new Set(
            [...dataPaths]
                .map(root => sanitizePath(root).split('/'))
                .flatMap((segments: string[]): string[] =>
                    segments.map((_value, index) => sanitizePath(segments.slice(0, index + 1).join('/')))
                )
        ),
    ].sort()

    // Assign a stable set of identifiers for each of these paths. We start counting at one here due
    // to needing to have our indexes align with. See inline comments below for more detail.
    const treeIdsByPath = new Map(allPaths.map((name, index) => [name, index + 1]))

    // Build functions we can use to query which paths are direct parents and children of one another
    const { parentOf, childrenOf } = buildTreeQuerier(treeIdsByPath)

    // Build our list of tree nodes
    const nodes = [
        // The first is a synthetic fake node that isn't rendered
        buildNode(0, '', null, childrenOf(undefined)),
        // The remainder of the nodes come from our treeIds (which we started counting at one)
        ...[...treeIdsByPath.entries()]
            .sort(byKey)
            .map(([root, id]) => buildNode(id, root, parentOf(id), childrenOf(id))),
    ]

    // tryUnlink will attempt to unlink the give node from the list of nodes forming a tree.
    // Returns true if a node re-link occurred.
    const tryUnlink = (nodes: TreeNodeWithDisplayName[], nodeId: number): boolean => {
        const node = nodes[nodeId]
        if (nodeId === 0 || node.parent === null || node.children.length !== 1) {
            // Not a candidate - no  unique parent/child to re-link
            return false
        }
        if (node.displayName === '/') {
            // usability :comfy:
            return false
        }
        const parentId = node.parent
        const childId = node.children[0]

        // Link parent to child
        nodes[childId].parent = parentId
        // Replace replace node by child in parent
        nodes[parentId].children = nodes[parentId].children.map(id => (id === nodeId ? childId : id))
        // Move (prepend) text from node to child
        nodes[childId].displayName = nodes[nodeId].displayName + nodes[childId].displayName

        return true
    }

    const unlinkedIds = new Set(
        nodes
            // Attempt to unlink/collapse all paths that do not have data
            .filter(node => !dataPaths.has(node.name) && tryUnlink(nodes, node.id))
            // Return node for organ harvesting :screamcat:
            .map(node => node.id)
    )

    return (
        nodes
            // Remove each of the roots we've marked for skipping in the loop above
            .filter((_value, index) => !unlinkedIds.has(index))
            // Remap each of the identifiers. We just collapse numbers so the sequence remains gap-less.
            // For some reason the wildcard <Tree /> component is a big stickler for having id and indexes align.
            .map(node => rewriteNodeIds(node, id => id - [...unlinkedIds].filter(unlinkedId => unlinkedId < id).length))
    )
}

export function descendentNames(treeData: TreeNodeWithDisplayName[], id: number): string[] {
    return [
        // children names
        ...treeData[id].children.map(id => treeData[id].name),
        // descendent names
        ...treeData[id].children.flatMap(id => descendentNames(treeData, id)),
    ]
}

interface TreeQuerier {
    parentOf: (id: number) => number
    childrenOf: (id: number | undefined) => number[]
}

// Return a pair of functions that can return the immediate parents and children of paths given tree identifiers.
function buildTreeQuerier(idsByPath: Map<string, number>): TreeQuerier {
    // Construct a map from identifiers of paths to the identifier of their immediate parent path
    const parentTreeIdByTreeId = new Map(
        [...idsByPath.entries()].map(([path, id]) => [
            id,
            [...idsByPath.keys()]
                // Filter out any non-ancestor directories
                // (NOTE: paths here guaranteed to start with slash)
                .filter(child =>
                    // Trim trailing slash and split each input (covers the `/` case)
                    checkSubset(child.replace(/(\/)$/, '').split('/'), path.replace(/(\/)$/, '').split('/'))
                )
                .sort((a, b) => b.length - a.length) // Sort by reverse length (most specific proper ancestor first)
                .map(key => idsByPath.get(key))[0], // Take the first element as its associated identifier
        ])
    )

    return {
        // Return parent identifier of entry (or zero if undefined)
        parentOf: id => parentTreeIdByTreeId.get(id) || 0,
        // Return identifiers of entries that declare their own parent as the target
        childrenOf: id => keysMatchingPredicate(parentTreeIdByTreeId, parentId => parentId === id),
    }
}

// Create a node with a default display name based on name (a filepath in this case)
function buildNode(id: number, name: string, parent: number | null, children: number[]): TreeNodeWithDisplayName {
    return { id, name, parent, children, displayName: `${name.split('/').reverse()[0]}/` }
}

// Rewrite the identifiers in each of the given tree node's fields.
function rewriteNodeIds(
    { id, parent, children, ...rest }: TreeNodeWithDisplayName,
    rewriteId: (id: number) => number
): TreeNodeWithDisplayName {
    return {
        id: rewriteId(id),
        parent: parent !== null ? rewriteId(parent) : null,
        children: children.map(rewriteId).sort(),
        ...rest,
    }
}
