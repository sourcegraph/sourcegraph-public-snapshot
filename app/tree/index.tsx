import flatten from 'lodash/flatten'
import groupBy from 'lodash/groupBy'
import sortBy from 'lodash/sortBy'
import { BehaviorSubject } from 'rxjs'

export interface TreeNode {
    text: string
    children: TreeNode[]
    state: BehaviorSubject<TreeNodeState>
    id?: string
    a_attr?: {
        href: string
    }
    filePath: string
}

interface TreeNodeState {
    collapsed: boolean
    selected: boolean
}

export function parseNodes(paths: string[], selectedPath: string): any {
    const getFilePath = (prefix: string, restParts: string[]) => {
        if (prefix === '') {
            return restParts.join('/')
        }
        return prefix + '/' + restParts.join('/')
    }

    const parseHelper = (
        splits: string[][],
        subpath = '',
        nodeMap = new Map<string, TreeNode>()
    ): { nodes: TreeNode[]; nodeMap: Map<string, TreeNode> } => {
        const splitsByDir = groupBy(splits, split => {
            if (split.length === 1) {
                return ''
            }
            return split[0]
        })

        const entries = flatten<TreeNode>(
            Object.entries(splitsByDir).map(([dir, pathSplits]) => {
                if (dir === '') {
                    return pathSplits.map(split => {
                        const filePath = getFilePath(subpath, split)
                        const node: TreeNode = {
                            text: filePath.split('/').pop()!,
                            children: [],
                            filePath,
                            state: new BehaviorSubject<TreeNodeState>({
                                selected: filePath === selectedPath,
                                collapsed: true,
                            }),
                            id: filePath,
                        }
                        nodeMap.set(filePath, node)
                        return node
                    })
                }

                const dirPath = getFilePath(subpath, [dir])
                const dirNode: TreeNode = {
                    text: dirPath.split('/').pop()!,
                    children: parseHelper(
                        pathSplits.map(split => split.slice(1)),
                        subpath ? subpath + '/' + dir : dir,
                        nodeMap
                    ).nodes,
                    filePath: dirPath,
                    state: new BehaviorSubject<TreeNodeState>({
                        selected: dirPath === selectedPath,
                        collapsed: true,
                    }),
                }
                nodeMap.set(dirPath, dirNode)
                return [dirNode]
            })
        )

        // filter entries to those that are siblings to the selectedPath
        let filter = entries
        const selectedPathParts = selectedPath.split('/')
        let part = 0
        while (true) {
            if (part >= selectedPathParts.length) {
                break
            }
            let matchedDir: TreeNode | undefined
            // let matchedFile: TreeNode | undefined
            for (const entry of filter) {
                if (entry.filePath.split('/').pop() === selectedPathParts[part] && entry.children.length > 0) {
                    matchedDir = entry
                    break
                }
            }
            if (matchedDir) {
                filter = matchedDir.children
            }
            if (part === selectedPathParts.length - 1) {
                // on the last part, filter either contains the matched file + siblings, or the matched directories children
                break
            }
            ++part
        }

        // directories first (nodes w/ children), then sort lexicographically
        return { nodes: sortBy(filter, [(e: TreeNode) => (e.children.length > 0 ? 0 : 1), 'text']), nodeMap }
    }

    return parseHelper(paths.map(path => path.split('/')))
}
