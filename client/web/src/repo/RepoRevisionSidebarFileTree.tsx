import { useCallback, useRef, useState } from 'react'

import { mdiFileDocumentOutline, mdiSourceRepository, mdiFolderOutline, mdiFolderOpenOutline } from '@mdi/js'
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
    Tooltip,
    ErrorAlert,
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
type FileTreeEntry = Extract<
    Extract<
        Extract<FileTreeEntriesResult['repository'], { __typename?: 'Repository' }>['commit'],
        { __typename?: 'GitCommit' }
    >['tree'],
    { __typename?: 'GitTree' }
>['entries'][number]

interface Props {
    repoName: string
    revision: string
    commitID: string
    initialFilePath: string
    initialFilePathIsDirectory: boolean
    telemetryService: TelemetryService
}
export const RepoRevisionSidebarFileTree: React.FunctionComponent<Props> = props => {
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
            ? [defaultNodeId, ...getAllParentsOfPath(treeData, defaultNode?.entry?.path ?? '').map(node => node.id)]
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
    const { entry, error } = element
    const submodule = entry?.submodule
    const name = entry?.name
    const url = entry?.url

    if (error) {
        return <ErrorAlert className="m-0" variant="note" error={error} />
    }

    if (submodule) {
        const rev = submodule.commit.slice(0, 7)
        const title = `${name} @ ${rev}`

        let content = <span>{title}</span>
        if (url) {
            content = (
                <Link
                    to={url ?? '#'}
                    tabIndex={-1}
                    onClick={event => {
                        event.preventDefault()
                        handleSelect(event)
                    }}
                >
                    {title}
                </Link>
            )
        }

        return (
            <>
                <Tooltip content={'Submodule: ' + submodule.url}>
                    <span>
                        <Icon
                            svgPath={mdiSourceRepository}
                            className={classNames('mr-1', styles.icon)}
                            aria-label={'Submodule: ' + submodule.url}
                        />
                        {content}
                    </span>
                </Tooltip>
            </>
        )
    }

    return (
        <>
            <Icon
                svgPath={isBranch ? (isExpanded ? mdiFolderOpenOutline : mdiFolderOutline) : mdiFileDocumentOutline}
                className={classNames('mr-1', styles.icon)}
                aria-hidden={true}
            />
            <Link
                to={url ?? '#'}
                tabIndex={-1}
                onClick={event => {
                    event.preventDefault()
                    handleSelect(event)
                }}
            >
                {name}
            </Link>
        </>
    )
}

type TreeNode = WildcardTreeNode & { entry: FileTreeEntry | null } & { error?: string }
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

    function appendNode(node: TreeNode, path: string): void {
        const children = tree.nodes.filter(potentialChild => isAParentOfB(path, potentialChild.entry?.path ?? ''))
        const parent = tree.nodes.find(potentialParent => isAParentOfB(potentialParent.entry?.path ?? '', path))

        // Fix all children references
        node.children = children.map(child => child.id)
        for (const child of children) {
            child.parent = node.id
        }

        // Fix all parent references
        if (parent) {
            node.parent = parent.id
            parent.children.push(node.id)
        }

        if (node.entry) {
            tree.pathToId.set(node.entry.path, node.id)
        }
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

    let siblingCount = 0
    for (let idx = 0; idx < entries.length; idx++) {
        const entry = entries[idx]

        if (tree.pathToId.has(entry.path)) {
            continue
        }

        const isSiblingOfPreviousNode = idx > 0 && dirname(entries[idx - 1].path) === dirname(entry.path)

        if (isSiblingOfPreviousNode) {
            siblingCount++
        } else {
            siblingCount = 0
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
        appendNode(node, entry.path)

        // We have reached the maximum number of entries for a directory. Add a
        // dummy node to indicate that there are more entries.
        if (siblingCount === MAX_TREE_ENTRIES - 1) {
            siblingCount = 0
            const errorMessage = 'Full list of files is too long to display. Use search to find a specific file.'
            const id = tree.nodes.length
            const node: TreeNode = {
                name: errorMessage,
                id,
                isBranch: false,
                parent: 0,
                children: [],
                entry: null,
                error: errorMessage,
            }
            appendNode(node, entry.path)
        }
    }

    return tree
}

function setLoadedPath(tree: TreeData, path: string): TreeData {
    tree = { ...tree, loadedPaths: new Set(tree.loadedPaths) }
    tree.loadedPaths.add(path)
    return tree
}

function isAParentOfB(fullPathA: string, fullPathB: string): boolean {
    // B is the root so it can not have a parent
    if (fullPathB === '') {
        return false
    }
    const bParentPath = fullPathB.split('/').slice(0, -1).join('/') || ''
    return fullPathA === bParentPath
}

function getAllParentsOfPath(tree: TreeData, fullPath: string): TreeNode[] {
    const parents = []

    let parent = tree.nodes.find(potentialParent => isAParentOfB(potentialParent.entry?.path ?? '', fullPath))
    while (parent) {
        parents.push(parent)

        parent = tree.nodes.find(potentialParent =>
            isAParentOfB(potentialParent.entry?.path ?? '', parent!.entry?.path ?? '')
        )
    }
    return parents
}
