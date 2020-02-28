import * as path from 'path'
import { dirnameWithoutDot } from './paths'

/**
 * Create a directory tree from the complete set of file paths in an LSIF dump.
 * Returns a generator that will perform a breadth-first traversal of the tree.
 *
 * @param root The LSIF dump root.
 * @param documentPaths The set of file paths in an LSIF dump.
 */
export function createBatcher(root: string, documentPaths: string[]): Generator<string[], void, string[]> {
    return traverse(createTree(root, documentPaths))
}

/** A node in a directory path tree. */
interface Node {
    /** The name of a directory in this level of the directory tree. */
    dirname: string

    /** The segments directly nested in this node. */
    children: Node[]
}

/**
 * Create a directory tree from the complete set of file paths in an LSIF dump.
 *
 * @param root The LSIF dump root.
 * @param documentPaths The set of file paths in an LSIF dump.
 */
function createTree(root: string, documentPaths: string[]): Node {
    // Construct the the set of root-relative parent directories for each path
    const documentDirs = documentPaths.map(documentPath => dirnameWithoutDot(path.join(root, documentPath)))
    // Deduplicate and throw out any directory that is outside of the dump root
    const dirs = Array.from(new Set(documentDirs)).filter(dirname => !dirname.startsWith('..'))

    // Construct the canned root node
    const rootNode: Node = { dirname: '', children: [] }

    for (const dir of dirs) {
        // Skip the dump root as the following loop body would make
        // it a child of itself. This is a non-obvious edge condition.
        if (dir === '') {
            continue
        }

        let node = rootNode

        // For each path segment in this directory, traverse down the
        // tree. Any node that doesn't exist is created with empty
        // children.

        for (const dirname of dir.split('/')) {
            let child = node.children.find(n => n.dirname === dirname)
            if (!child) {
                child = { dirname, children: [] }
                node.children.push(child)
            }

            node = child
        }
    }

    return rootNode
}

/**
 * Perform a breadth-first traversal of a directory tree. Returns a generator that
 * acts as a coroutine visitor. The generator first yields the root (empty) path,
 * then waits for the caller to return the set of paths from the last batch that
 * exist in git. The next batch returned will only include the paths that are
 * properly nested under a previously returned value.
 *
 * @param root The root node.
 */
function* traverse(root: Node): Generator<string[], void, string[]> {
    // The frontier is the candidate batch that will be returned in the next
    // call to the generator. This is a superset of paths that will actually
    // be returned after pruning non-existent paths.
    let frontier: [string, Node[]][] = [['', root.children]]

    while (frontier.length > 0) {
        // Yield our current batch and wait for the caller to return the set
        // of previously yielded paths that exist.
        const exists = yield frontier.map(([parent]) => parent)

        frontier = frontier
            // Remove any children from the frontier that are not a proper
            // descendant of some path that was confirmed to be exist. This
            // will stop us from iterating children that definitely don't
            // exist as the parent is already known to be an un-tracked path.
            .filter(([parent]) => exists.includes(parent))
            .flatMap(([parent, children]) =>
                // Join the current path to a node with the node's segment to
                // get the full path to that node. Create the new frontier from
                // the current frontier's children after pruning the non-existent
                // paths.
                children.map((child): [string, Node[]] => [path.join(parent, child.dirname), child.children])
            )
    }
}
