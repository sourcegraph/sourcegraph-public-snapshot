import { gql, useQuery } from '@sourcegraph/http-client'
import { Alert, ErrorMessage, Icon, LoadingSpinner } from '@sourcegraph/wildcard'
import { useCallback, useState } from 'react'
import { FileTreeEntriesResult, FileTreeEntriesVariables } from '../graphql-operations'
import { MAX_TREE_ENTRIES } from '../tree/constants'
import { dirname } from '../util/path'
import TreeView, { INode } from 'react-accessible-treeview'
import { mdiFileDocumentOutline, mdiFolderOutline, mdiMenuRight, mdiMenuDown } from '@mdi/js'

import styles from './RepoRevisionSidebarFileTree.module.scss'

const QUERY = gql`
    query FileTreeEntries(
        $repoName: String!
        $revision: String!
        $commitID: String!
        $filePath: String!
        $recursiveParents: Boolean!
        $first: Int
    ) {
        repository(name: $repoName) {
            commit(rev: $commitID, inputRevspec: $revision) {
                tree(path: $filePath) {
                    isRoot
                    url
                    entries(first: $first, recursiveParents: $recursiveParents) {
                        name
                        path
                        isDirectory
                        url
                        submodule {
                            url
                            commit
                        }
                        isSingleChild
                    }
                }
            }
        }
    }
`

// TODO(philipp-spiess): Figure out why we can't use a real fragment here. The
// issue is that on `TreeEntry` does not match because the type is either
// `GitTree` or `GitBlob` and it needs to match?
interface FileTreeEntry {
    name: string
    path: string
    isDirectory: boolean
    url: string
    isSingleChild: boolean
    submodule: { __typename?: 'Submodule'; url: string; commit: string } | null
}

interface Props {
    repoName: string
    revision: string
    commitID: string
    initialFilePath: string
    initialFilePathIsDirectory: boolean
}
export function RepoRevisionSidebarFileTree(props: Props) {
    // Ensure that the initial file path does not update when the props change
    const [initialFilePath] = useState(
        props.initialFilePathIsDirectory ? props.initialFilePath : dirname(props.initialFilePath)
    )

    const [treeData, setTreeData] = useState<TreeData | null>(null)

    const defaultVariables = {
        repoName: props.repoName,
        revision: props.revision,
        commitID: props.commitID,
        first: MAX_TREE_ENTRIES,
    }

    const { error, loading, refetch } = useQuery<FileTreeEntriesResult, FileTreeEntriesVariables>(QUERY, {
        variables: {
            ...defaultVariables,
            filePath: initialFilePath,
            // The initial load should include parents
            recursiveParents: true,
        },
        fetchPolicy: 'cache-and-network',
        onCompleted(data) {
            console.log('onCompleted', data)
            const tree = data?.repository?.commit?.tree?.entries
            if (!tree) {
                throw new Error('No tree data')
            }
            if (treeData === null) {
                setTreeData(appendTreeData(createTreeData(), tree))
            } else {
                setTreeData(treeData => appendTreeData(treeData!, tree))
            }
        },
    })

    const defaultNodeId = treeData?.pathToId.get(props.initialFilePath)
    const defaultNode = defaultNodeId ? treeData?.nodes[defaultNodeId] : undefined
    const defaultSelectedIds = defaultNodeId ? [defaultNodeId] : []
    const defaultExpandedIds =
        treeData && defaultNode
            ? [defaultNodeId, ...getAllParentsOfNode(treeData, defaultNode).map(node => node.id)]
            : []

    const onLoadData = useCallback(async ({ element }: { element: INode }) => {
        console.log('onLoadData', element)
        const alreadyLoaded = element.children?.length > 0 || treeData?.loadedPaths.has(element.name)

        if (alreadyLoaded) {
            return
        }

        await refetch({
            ...defaultVariables,
            filePath: element.name,
            recursiveParents: false,
        })

        setTreeData(treeData => setLoadedPath(treeData!, element.name))
    }, [])

    if (error) {
        return (
            <Alert variant="danger">
                <ErrorMessage error={error.message} />
            </Alert>
        )
    }
    if (loading || treeData === null) {
        return <LoadingSpinner />
    }

    return (
        <TreeView
            data={treeData.nodes}
            aria-label="file tree"
            defaultSelectedIds={defaultSelectedIds}
            defaultExpandedIds={defaultExpandedIds}
            onLoadData={onLoadData}
            className={styles.fileTree}
            nodeRenderer={({
                element,
                isBranch,
                isExpanded,
                isSelected,
                getNodeProps,
                level,
                handleSelect,
                handleExpand,
            }) => {
                const branchNode = (isExpanded: boolean, element: INode) =>
                    isExpanded && element.children.length === 0 ? (
                        <LoadingSpinner className="mr-1" />
                    ) : (
                        <Icon aria-hidden={true} svgPath={isExpanded ? mdiMenuRight : mdiMenuDown} className="mr-1" />
                    )
                return (
                    <div
                        {...getNodeProps({ onClick: handleExpand })}
                        style={{ marginLeft: `${1 * (level - 1)}rem`, background: isSelected ? 'red' : undefined }}
                        data-tree-node-id={element.id}
                    >
                        {isBranch && branchNode(isExpanded, element)}
                        <Icon
                            svgPath={isBranch ? mdiFolderOutline : mdiFileDocumentOutline}
                            className="mr-1"
                            onClick={e => {
                                handleSelect(e)
                                e.stopPropagation()
                            }}
                            aria-hidden={true}
                        />
                        <span
                            className="name"
                            onClick={e => {
                                handleSelect(e)
                                e.stopPropagation()
                            }}
                        >
                            {element.name}
                        </span>
                    </div>
                )
            }}
        />
    )
}

interface TreeData {
    // The flat nodes list used by react-accessible-treeview
    nodes: INode[]

    // A map to quickly find the number ID for a node by its path
    pathToId: Map<string, number>

    // A set for paths that have been loaded. We can not rely on childern.length
    // because a directory can have no children.
    loadedPaths: Set<string>
}

function createTreeData(): TreeData {
    return {
        nodes: [],
        pathToId: new Map(),
        loadedPaths: new Set(),
    }
}

function appendTreeData(tree: TreeData, entries: FileTreeEntry[]): TreeData {
    tree = { ...tree, nodes: [...tree.nodes], pathToId: new Map(tree.pathToId) }

    function addTreeEntry(entry: FileTreeEntry): void {
        if (tree.pathToId.has(entry.path)) {
            return
        }

        const id = tree.nodes.length

        if (entry.path === undefined) {
            debugger
        }

        const node: INode = {
            name: entry.path,
            id,
            isBranch: entry.isDirectory,
            parent: 0,
            children: [],
        }

        const children = tree.nodes.filter(potentialChild => isAParentOfB(node, potentialChild))
        const parent = tree.nodes.find(potentialParent => isAParentOfB(potentialParent, node))

        console.log({ node: node.name, children, parent: parent?.name })

        // Fix all children references
        node.children = children.map(child => child.id)
        for (const child of children) {
            child.parent = id
        }

        // Fix all parent references
        if (parent) {
            node.parent = parent.id
            parent.children.push(id)
        }

        tree.pathToId.set(entry.path, id)
        tree.nodes.push(node)
    }

    // Insert a root node
    if (tree.nodes.length === 0) {
        const id = 0
        tree.pathToId.set('', id)
        tree.nodes.push({
            name: '',
            id,
            isBranch: true,
            parent: null,
            children: [],
        })
    }

    for (const entry of entries) {
        addTreeEntry(entry)
    }

    return tree
}

function setLoadedPath(tree: TreeData, path: string): TreeData {
    tree = { ...tree, loadedPaths: new Set(tree.loadedPaths) }
    tree.loadedPaths.add(path)
    return tree
}

function isAParentOfB(a: INode, b: INode): boolean {
    // B is the root so it can not have a parent
    if (b.name === '') {
        return false
    }
    const bParentName = b.name.split('/').slice(0, -1).join('/') || ''
    return a.name === bParentName
}

function getAllParentsOfNode(tree: TreeData, node: INode): INode[] {
    const parents = []

    let parent = tree.nodes.find(potentialParent => isAParentOfB(potentialParent, node))
    while (parent) {
        parents.push(parent)
        parent = tree.nodes.find(potentialParent => isAParentOfB(potentialParent, parent!))
    }
    return parents
}
