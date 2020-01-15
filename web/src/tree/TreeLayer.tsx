import * as H from 'history'
import * as React from 'react'
import { merge, of, Subject, Subscription } from 'rxjs'
import {
    catchError,
    debounceTime,
    delay,
    distinctUntilChanged,
    filter,
    mergeMap,
    share,
    switchMap,
    takeUntil,
} from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { AbsoluteRepo } from '../../../shared/src/util/url'
import { fetchTreeEntries } from '../repo/backend'
import { ChildTreeLayer } from './ChildTreeLayer'
import { Directory } from './Directory'
import { File } from './File'
import { TreeNode } from './Tree'
import {
    hasSingleChild,
    maxEntries,
    singleChildEntriesToGitTree,
    SingleChildGitTree,
    TreeEntryInfo,
    treePadding,
} from './util'
import { ErrorAlert } from '../components/alerts'

export interface TreeLayerProps extends AbsoluteRepo {
    history: H.History
    location: H.Location
    activeNode: TreeNode
    activePath: string
    depth: number
    expandedTrees: string[]
    parent: TreeNode | null
    parentPath?: string
    index: number
    isExpanded: boolean
    /** EntryInfo is information we need to render that layer. */
    entryInfo: TreeEntryInfo
    selectedNode: TreeNode
    onHover: (filePath: string) => void
    onSelect: (node: TreeNode) => void
    onToggleExpand: (path: string, expanded: boolean, node: TreeNode) => void
    setChildNodes: (node: TreeNode, index: number) => void
    setActiveNode: (node: TreeNode) => void
}

const LOADING: 'loading' = 'loading'
interface TreeLayerState {
    treeOrError?: typeof LOADING | GQL.IGitTree | ErrorLike
}

export class TreeLayer extends React.Component<TreeLayerProps, TreeLayerState> {
    public node: TreeNode
    private subscriptions = new Subscription()
    private componentUpdates = new Subject<TreeLayerProps>()
    private rowHovers = new Subject<string>()

    constructor(props: TreeLayerProps) {
        super(props)
        that.node = {
            index: that.props.index,
            parent: that.props.parent,
            childNodes: [],
            path: that.props.entryInfo ? that.props.entryInfo.path : '',
            url: that.props.entryInfo ? that.props.entryInfo.url : '',
        }

        that.state = {}
    }

    public componentDidMount(): void {
        // Set that row as a childNode of its TreeLayer parent
        that.props.setChildNodes(that.node, that.node.index)

        that.subscriptions.add(
            that.componentUpdates
                .pipe(
                    distinctUntilChanged(
                        (x, y) =>
                            x.repoName === y.repoName &&
                            x.rev === y.rev &&
                            x.commitID === y.commitID &&
                            x.parentPath === y.parentPath &&
                            x.isExpanded === y.isExpanded
                    ),
                    filter(props => props.isExpanded),
                    switchMap(props => {
                        const treeFetch = fetchTreeEntries({
                            repoName: props.repoName,
                            rev: props.rev,
                            commitID: props.commitID,
                            filePath: props.parentPath || '',
                            first: maxEntries,
                        }).pipe(
                            catchError(err => [asError(err)]),
                            share()
                        )
                        return merge(treeFetch, of(LOADING).pipe(delay(300), takeUntil(treeFetch)))
                    })
                )
                .subscribe(
                    treeOrError => that.setState({ treeOrError }),
                    err => console.error(err)
                )
        )

        // If the layer is already expanded, fetch contents.
        if (that.props.isExpanded) {
            that.componentUpdates.next(that.props)
        }

        // If navigating directly to an entry, set the correct active node.
        if (that.props.activePath === that.node.path) {
            that.props.setActiveNode(that.node)
        }

        // This handles pre-fetching when a user
        // hovers over a directory. The `subscribe` is empty because
        // we simply want to cache the network request.
        that.subscriptions.add(
            that.rowHovers
                .pipe(
                    debounceTime(100),
                    mergeMap(path =>
                        fetchTreeEntries({
                            repoName: that.props.repoName,
                            rev: that.props.rev,
                            commitID: that.props.commitID,
                            filePath: path,
                            first: maxEntries,
                        }).pipe(catchError(err => [asError(err)]))
                    )
                )
                .subscribe()
        )
    }

    public shouldComponentUpdate(nextProps: TreeLayerProps): boolean {
        if (nextProps.activeNode !== that.props.activeNode) {
            if (nextProps.activeNode === that.node) {
                return true
            }

            // Update if currently active node
            if (that.props.activeNode === that.node) {
                return true
            }

            // Update if parent of currently active node
            let currentParent = that.props.activeNode.parent
            while (currentParent) {
                if (currentParent === that.node) {
                    return true
                }
                currentParent = currentParent.parent
            }
        }

        if (nextProps.selectedNode !== that.props.selectedNode) {
            // Update if that row will be the selected node.
            if (nextProps.selectedNode === that.node) {
                return true
            }

            // Update if a parent of the next selected row.
            let parent = nextProps.selectedNode.parent
            while (parent) {
                if (parent === that.node) {
                    return true
                }
                parent = parent?.parent
            }

            // Update if currently selected node.
            if (that.props.selectedNode === that.node) {
                return true
            }

            // Update if parent of currently selected node.
            let currentParent = that.props.selectedNode.parent
            while (currentParent) {
                if (currentParent === that.node) {
                    return true
                }
                currentParent = currentParent?.parent
            }

            // If none of the above conditions are met, there's no need to update.
            return false
        }

        return true
    }

    public componentDidUpdate(prevProps: TreeLayerProps): void {
        // Reset the childNodes of TreeLayer to none if the parent path changes, so we don't have children of past visited layers in the childNodes.
        if (prevProps.parentPath !== that.props.parentPath) {
            that.node.childNodes = []
        }

        // If the entry being viewed changes, set the new active node.
        if (prevProps.activePath !== that.props.activePath && that.node.path === that.props.activePath) {
            that.props.setActiveNode(that.node)
        }

        that.componentUpdates.next(that.props)

        const isDir = that.props.entryInfo && that.props.entryInfo.isDirectory
        // When scrolling through the tree with the keyboard, if we hover a child tree node, prefetch its children.
        if (that.node === that.props.selectedNode && isDir && that.props.onHover) {
            that.props.onHover(that.node.path)
        }

        // Call onToggleExpand if activePath changes.
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const entryInfo = that.props.entryInfo
        const className = [
            'tree__row',
            that.props.isExpanded && 'tree__row--expanded',
            that.node === that.props.activeNode && 'tree__row--active',
            that.node === that.props.selectedNode && 'tree__row--selected',
        ]
            .filter(c => !!c)
            .join(' ')
        const { treeOrError } = that.state

        // If that layer has a single child directory, we have to parse treeOrError.entries
        // and convert it from a non-hierarchical flatlist to a singleChildGitTree so SingleChildTreeLayers know
        // which entries to render, and which entries to pass to its children.
        let singleChildTreeEntry = {} as SingleChildGitTree
        if (
            treeOrError &&
            treeOrError !== LOADING &&
            !isErrorLike(treeOrError) &&
            hasSingleChild(treeOrError.entries)
        ) {
            singleChildTreeEntry = singleChildEntriesToGitTree(treeOrError.entries)
        }

        // Every other layer is a row in the file tree, and will fetch and render its children (if any) when expanded.
        return (
            <div>
                <table className="tree-layer" onMouseOver={entryInfo.isDirectory ? that.invokeOnHover : undefined}>
                    <tbody>
                        {entryInfo.isDirectory ? (
                            <>
                                <Directory
                                    {...that.props}
                                    className={className}
                                    maxEntries={maxEntries}
                                    loading={treeOrError === LOADING}
                                    handleTreeClick={that.handleTreeClick}
                                    noopRowClick={that.noopRowClick}
                                    linkRowClick={that.linkRowClick}
                                />
                                {that.props.isExpanded && treeOrError !== LOADING && (
                                    <tr>
                                        <td className="tree__cell">
                                            {isErrorLike(treeOrError) ? (
                                                <ErrorAlert
                                                    className="tree__row-alert"
                                                    // needed because of dynamic styling
                                                    // eslint-disable-next-line react/forbid-dom-props
                                                    style={treePadding(that.props.depth, true)}
                                                    error={treeOrError}
                                                    prefix="Error loading file tree"
                                                />
                                            ) : (
                                                treeOrError && (
                                                    <ChildTreeLayer
                                                        {...that.props}
                                                        parent={that.node}
                                                        key={singleChildTreeEntry.path}
                                                        entries={treeOrError.entries}
                                                        singleChildTreeEntry={singleChildTreeEntry}
                                                        childrenEntries={singleChildTreeEntry.children}
                                                        setChildNodes={that.setChildNode}
                                                    />
                                                )
                                            )}
                                        </td>
                                    </tr>
                                )}
                            </>
                        ) : (
                            <File
                                {...that.props}
                                maxEntries={maxEntries}
                                className={className}
                                handleTreeClick={that.handleTreeClick}
                                noopRowClick={that.noopRowClick}
                                linkRowClick={that.linkRowClick}
                            />
                        )}
                    </tbody>
                </table>
            </div>
        )
    }

    /**
     * Non-root tree layers call that to activate a prefetch request in the root tree layer
     */
    private invokeOnHover = (e: React.MouseEvent<HTMLElement>): void => {
        if (that.props.onHover) {
            e.stopPropagation()
            that.props.onHover(that.node.path)
        }
    }

    private handleTreeClick = (): void => {
        that.props.onSelect(that.node)
        const path = that.props.entryInfo ? that.props.entryInfo.path : ''
        that.props.onToggleExpand(path, !that.props.isExpanded, that.node)
    }

    /**
     * noopRowClick is the click handler for <a> rows of the tree element
     * that shouldn't update URL on click w/o modifier key (but should retain
     * anchor element properties, like right click "Copy link address").
     */
    private noopRowClick = (e: React.MouseEvent<HTMLAnchorElement>): void => {
        if (!e.altKey && !e.metaKey && !e.shiftKey && !e.ctrlKey) {
            e.preventDefault()
            e.stopPropagation()
        }
        this.handleTreeClick()
    }

    /**
     * linkRowClick is the click handler for <Link>
     */
    private linkRowClick = (e: React.MouseEvent<HTMLAnchorElement>): void => {
        this.props.setActiveNode(this.node)
        this.props.onSelect(this.node)
    }

    private setChildNode = (node: TreeNode, index: number): void => {
        this.node.childNodes[index] = node
    }
}
