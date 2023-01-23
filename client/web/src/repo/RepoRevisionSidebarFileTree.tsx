import { useCallback, useRef, useState } from 'react'

import { mdiFileDocumentOutline, mdiFolderOutline, mdiFolderOpenOutline } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom-v5-compat'

import { gql, useQuery } from '@sourcegraph/http-client'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Alert,
    ErrorMessage,
    Icon,
    TreeNode as WildcardTreeNode,
    Link,
    LoadingSpinner,
    Tree,
} from '@sourcegraph/wildcard'

import { FileTreeEntriesResult, FileTreeEntriesVariables } from '../graphql-operations'
import { MAX_TREE_ENTRIES } from '../tree/constants'
import { dirname } from '../util/path'

import styles from './RepoRevisionSidebarFileTree.module.scss'

const QUERY = gql`
    query FileTreeEntries(
        $repoName: String!
        $revision: String!
        $commitID: String!
        $filePath: String!
        $ancestors: Boolean!
        $first: Int
    ) {
        repository(name: $repoName) {
            commit(rev: $commitID, inputRevspec: $revision) {
                tree(path: $filePath) {
                    isRoot
                    url
                    entries(first: $first, ancestors: $ancestors) {
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

// TODO(philipp-spiess): Figure out why we can't use a GraphQL fragment here.
// The issue is that on `TreeEntry` does not match because the type is either
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
    telemetryService: TelemetryService
}
export function RepoRevisionSidebarFileTree(props: Props): JSX.Element {
    // Ensure that the initial file path does not update when the props change
    const [initialFilePath] = useState(
        props.initialFilePathIsDirectory ? props.initialFilePath : dirname(props.initialFilePath)
    )
    const [treeData, setTreeData] = useState<TreeData | null>(null)

    const navigate = useNavigate()

    const [defaultVariables] = useState({
        repoName: props.repoName,
        revision: props.revision,
        commitID: props.commitID,
        first: MAX_TREE_ENTRIES,
    })

    const { error, loading, refetch } = useQuery<FileTreeEntriesResult, FileTreeEntriesVariables>(QUERY, {
        variables: {
            ...defaultVariables,
            filePath: initialFilePath,
            // The initial search should include all parent directories and
            // their entries, so we can render a collapsed tree view for the
            // initial file path.
            ancestors: true,
        },
        fetchPolicy: 'cache-and-network',
        onCompleted(data) {
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
        treeData && defaultNode && defaultNodeId
            ? [defaultNodeId, ...getAllParentsOfNode(treeData, defaultNode).map(node => node.id)]
            : []

    const onLoadData = useCallback(
        async ({ element }: { element: TreeNode }) => {
            const fullPath = element.entry?.path ?? ''
            const alreadyLoaded = element.children?.length > 0 || treeData?.loadedPaths.has(fullPath)
            if (alreadyLoaded || !element.isBranch) {
                return
            }

            props.telemetryService.log('FileTreeLoadDirectory')
            await refetch({
                ...defaultVariables,
                filePath: fullPath,
                ancestors: false,
            })

            setTreeData(treeData => setLoadedPath(treeData!, fullPath))
        },
        [defaultVariables, refetch, treeData?.loadedPaths, props.telemetryService]
    )

    const defaultSelectFiredRef = useRef<boolean>(false)
    const onSelect = useCallback(
        ({ element, isSelected }: { element: TreeNode; isSelected: boolean }) => {
            if (!isSelected) {
                return
            }

            // On the initial rendering, an onSelect event is fired for the
            // default node. We don't want to navigate to that node though.
            if (defaultSelectFiredRef.current === false && element.id === defaultNodeId) {
                defaultSelectFiredRef.current = true
                return
            }

            if (element.entry) {
                props.telemetryService.log('FileTreeClick')
                navigate(element.entry.url)
            }
        },
        [defaultNodeId, navigate, props.telemetryService]
    )

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
        <Tree<TreeNode>
            data={treeData.nodes}
            aria-label="file tree"
            defaultSelectedIds={defaultSelectedIds}
            defaultExpandedIds={defaultExpandedIds}
            onSelect={onSelect}
            onLoadData={onLoadData}
            renderNode={renderNode}
        />
    )
}

function renderNode({
    element,
    isBranch,
    isExpanded,
    handleSelect,
}: {
    element: TreeNode
    isBranch: boolean
    isExpanded: boolean
    handleSelect: (event: React.MouseEvent) => {}
}): React.ReactNode {
    return (
        <>
            <Icon
                svgPath={isBranch ? (isExpanded ? mdiFolderOpenOutline : mdiFolderOutline) : mdiFileDocumentOutline}
                className={classNames('mr-1', styles.icon)}
                aria-hidden={true}
            />
            <Link
                to={element.entry?.url ?? '#'}
                tabIndex={-1}
                onClick={event => {
                    event.preventDefault()
                    handleSelect(event)
                }}
            >
                {element.entry?.name}
            </Link>
        </>
    )
}

type TreeNode = WildcardTreeNode & { entry: FileTreeEntry | null }
interface TreeData {
    // The flat nodes list used by react-accessible-treeview
    nodes: TreeNode[]

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

        const node: TreeNode = {
            name: entry.name,
            id,
            isBranch: entry.isDirectory,
            parent: 0,
            children: [],
            entry,
        }

        const children = tree.nodes.filter(potentialChild => isAParentOfB(node, potentialChild))
        const parent = tree.nodes.find(potentialParent => isAParentOfB(potentialParent, node))

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
            entry: null,
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

function isAParentOfB(a: TreeNode, b: TreeNode): boolean {
    // B is the root so it can not have a parent
    const aFullPath = a.entry?.path || ''
    const bFullPath = b.entry?.path || ''
    if (bFullPath === '') {
        return false
    }
    const bParentPath = bFullPath.split('/').slice(0, -1).join('/') || ''
    return aFullPath === bParentPath
}

function getAllParentsOfNode(tree: TreeData, node: TreeNode): TreeNode[] {
    const parents = []

    let parent = tree.nodes.find(potentialParent => isAParentOfB(potentialParent, node))
    while (parent) {
        parents.push(parent)

        parent = tree.nodes.find(potentialParent => isAParentOfB(potentialParent, parent!))
    }
    return parents
}
