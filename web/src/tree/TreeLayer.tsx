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
    /** EntryInfo is information we need to render this layer. */
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
        this.node = {
            index: this.props.index,
            parent: this.props.parent,
            childNodes: [],
            path: this.props.entryInfo ? this.props.entryInfo.path : '',
            url: this.props.entryInfo ? this.props.entryInfo.url : '',
        }

        this.state = {}
    }

    public componentDidMount(): void {
        // Set this row as a childNode of its TreeLayer parent
        this.props.setChildNodes(this.node, this.node.index)

        this.subscriptions.add(
            this.componentUpdates
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
                    treeOrError => this.setState({ treeOrError }),
                    err => console.error(err)
                )
        )

        // If the layer is already expanded, fetch contents.
        if (this.props.isExpanded) {
            this.componentUpdates.next(this.props)
        }

        // If navigating directly to an entry, set the correct active node.
        if (this.props.activePath === this.node.path) {
            this.props.setActiveNode(this.node)
        }

        // This handles pre-fetching when a user
        // hovers over a directory. The `subscribe` is empty because
        // we simply want to cache the network request.
        this.subscriptions.add(
            this.rowHovers
                .pipe(
                    debounceTime(100),
                    mergeMap(path =>
                        fetchTreeEntries({
                            repoName: this.props.repoName,
                            rev: this.props.rev,
                            commitID: this.props.commitID,
                            filePath: path,
                            first: maxEntries,
                        }).pipe(catchError(err => [asError(err)]))
                    )
                )
                .subscribe()
        )
    }

    public shouldComponentUpdate(nextProps: TreeLayerProps): boolean {
        if (nextProps.activeNode !== this.props.activeNode) {
            if (nextProps.activeNode === this.node) {
                return true
            }

            // Update if currently active node
            if (this.props.activeNode === this.node) {
                return true
            }

            // Update if parent of currently active node
            let currentParent = this.props.activeNode.parent
            while (currentParent) {
                if (currentParent === this.node) {
                    return true
                }
                currentParent = currentParent.parent
            }
        }

        if (nextProps.selectedNode !== this.props.selectedNode) {
            // Update if this row will be the selected node.
            if (nextProps.selectedNode === this.node) {
                return true
            }

            // Update if a parent of the next selected row.
            let parent = nextProps.selectedNode.parent
            while (parent) {
                if (parent === this.node) {
                    return true
                }
                parent = parent?.parent
            }

            // Update if currently selected node.
            if (this.props.selectedNode === this.node) {
                return true
            }

            // Update if parent of currently selected node.
            let currentParent = this.props.selectedNode.parent
            while (currentParent) {
                if (currentParent === this.node) {
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
        if (prevProps.parentPath !== this.props.parentPath) {
            this.node.childNodes = []
        }

        // If the entry being viewed changes, set the new active node.
        if (prevProps.activePath !== this.props.activePath && this.node.path === this.props.activePath) {
            this.props.setActiveNode(this.node)
        }

        this.componentUpdates.next(this.props)

        const isDir = this.props.entryInfo && this.props.entryInfo.isDirectory
        // When scrolling through the tree with the keyboard, if we hover a child tree node, prefetch its children.
        if (this.node === this.props.selectedNode && isDir && this.props.onHover) {
            this.props.onHover(this.node.path)
        }

        // Call onToggleExpand if activePath changes.
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const entryInfo = this.props.entryInfo
        const className = [
            'tree__row',
            this.props.isExpanded && 'tree__row--expanded',
            this.node === this.props.activeNode && 'tree__row--active',
            this.node === this.props.selectedNode && 'tree__row--selected',
        ]
            .filter(c => !!c)
            .join(' ')
        const { treeOrError } = this.state

        // If this layer has a single child directory, we have to parse treeOrError.entries
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
                <table className="tree-layer" onMouseOver={entryInfo.isDirectory ? this.invokeOnHover : undefined}>
                    <tbody>
                        {entryInfo.isDirectory ? (
                            <>
                                <Directory
                                    {...this.props}
                                    className={className}
                                    maxEntries={maxEntries}
                                    loading={treeOrError === LOADING}
                                    handleTreeClick={this.handleTreeClick}
                                    noopRowClick={this.noopRowClick}
                                    linkRowClick={this.linkRowClick}
                                />
                                {this.props.isExpanded && treeOrError !== LOADING && (
                                    <tr>
                                        <td className="tree__cell">
                                            {isErrorLike(treeOrError) ? (
                                                <ErrorAlert
                                                    className="tree__row-alert"
                                                    // needed because of dynamic styling
                                                    // eslint-disable-next-line react/forbid-dom-props
                                                    style={treePadding(this.props.depth, true)}
                                                    error={treeOrError}
                                                    prefix="Error loading file tree"
                                                />
                                            ) : (
                                                treeOrError && (
                                                    <ChildTreeLayer
                                                        {...this.props}
                                                        parent={this.node}
                                                        key={singleChildTreeEntry.path}
                                                        entries={treeOrError.entries}
                                                        singleChildTreeEntry={singleChildTreeEntry}
                                                        childrenEntries={singleChildTreeEntry.children}
                                                        setChildNodes={this.setChildNode}
                                                    />
                                                )
                                            )}
                                        </td>
                                    </tr>
                                )}
                            </>
                        ) : (
                            <File
                                {...this.props}
                                maxEntries={maxEntries}
                                className={className}
                                handleTreeClick={this.handleTreeClick}
                                noopRowClick={this.noopRowClick}
                                linkRowClick={this.linkRowClick}
                            />
                        )}
                    </tbody>
                </table>
            </div>
        )
    }

    /**
     * Non-root tree layers call this to activate a prefetch request in the root tree layer
     */
    private invokeOnHover = (e: React.MouseEvent<HTMLElement>): void => {
        if (this.props.onHover) {
            e.stopPropagation()
            this.props.onHover(this.node.path)
        }
    }

    private handleTreeClick = (): void => {
        this.props.onSelect(this.node)
        const path = this.props.entryInfo ? this.props.entryInfo.path : ''
        this.props.onToggleExpand(path, !this.props.isExpanded, this.node)
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
