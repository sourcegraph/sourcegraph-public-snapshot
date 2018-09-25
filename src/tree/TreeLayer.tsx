import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
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
import * as GQL from '../backend/graphqlschema'
import { AbsoluteRepo } from '../repo'
import { fetchTreeEntries } from '../repo/backend'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { RepositoryIcon } from '../util/icons' // TODO: Switch to mdi icon
import { TreeNode } from './Tree'

interface TreeLayerProps extends AbsoluteRepo {
    history: H.History
    activeNode: TreeNode
    activePath: string
    activePathIsTree: boolean
    depth: number
    expandedTrees: string[]
    parent: TreeNode | null
    parentPath?: string
    index: number
    isExpanded: boolean
    isRoot: boolean
    entryInfo?: GQL.IGitBlob | GQL.IGitTree
    selectedNode: TreeNode
    /** Whether this tree layer is the only child of the parent layer. */
    isSingleChild: boolean
    onHover?: (filePath: string) => void
    onSelect: (node: TreeNode) => void
    onToggleExpand: (path: string, expanded: boolean, node: TreeNode) => void
    setChildNodes: (node: TreeNode, index: number) => void
    setActiveNode: (node: TreeNode) => void
}

const LOADING: 'loading' = 'loading'
interface TreeLayerState {
    treeOrError?: typeof LOADING | GQL.IGitTree | ErrorLike
}

const treePadding = (depth: number, isTree: boolean) => ({
    paddingLeft: depth * 12 + (isTree ? 0 : 12) + 12 + 'px',
    paddingRight: '16px',
})

const maxEntries = 2500

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
                            x.repoPath === y.repoPath &&
                            x.rev === y.rev &&
                            x.commitID === y.commitID &&
                            x.parentPath === y.parentPath &&
                            x.isExpanded === y.isExpanded
                    ),
                    filter(props => props.isExpanded),
                    switchMap(props => {
                        const treeFetch = fetchTreeEntries({
                            repoPath: props.repoPath,
                            rev: props.rev,
                            commitID: props.commitID,
                            filePath: props.parentPath || '',
                            first: maxEntries,
                        }).pipe(
                            catchError(err => [asError(err)]),
                            share()
                        )
                        return merge(
                            treeFetch,
                            of(LOADING).pipe(
                                delay(300),
                                takeUntil(treeFetch)
                            )
                        )
                    })
                )
                .subscribe(treeOrError => this.setState({ treeOrError }), err => console.error(err))
        )

        // If this layer is the only child, and it's a directory, expand it automatically. This
        // ensures the user doesn't have to open every directory if they are deeply nested and have no contents
        // other than another directory.
        if (this.props.isSingleChild && this.props.entryInfo && this.props.entryInfo.isDirectory) {
            this.props.onToggleExpand(this.props.entryInfo.path, true, this.node)
        }

        // When we're at the root tree layer or the tree is already expanded, fetch the tree contents on mount.
        // For other layers, fetch on hover or on expand.
        if (this.props.isRoot || this.props.isExpanded) {
            this.componentUpdates.next(this.props)
        }

        // If navigating directly to an entry, set the correct active node.
        if (this.props.activePath === this.node.path) {
            this.props.setActiveNode(this.node)
        }

        this.subscriptions.add(
            this.rowHovers
                .pipe(
                    debounceTime(100),
                    mergeMap(path =>
                        fetchTreeEntries({
                            repoPath: this.props.repoPath,
                            rev: this.props.rev,
                            commitID: this.props.commitID,
                            filePath: path || '',
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
                parent = parent && parent.parent
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
                currentParent = currentParent && currentParent.parent
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

        // If isRoot or there's no entry info, we are at the root layer, so simply load all top-level entries.
        if (this.props.isRoot || !entryInfo) {
            return (
                <table className="tree-layer" tabIndex={0}>
                    <tbody>
                        <tr>
                            <td className="tree__cell">
                                {this.state.treeOrError === LOADING ? (
                                    <div className="tree__row-loader">
                                        <LoadingSpinner className="icon-inline tree-page__entries-loader" />Loading tree
                                    </div>
                                ) : isErrorLike(this.state.treeOrError) ? (
                                    <div
                                        className="tree__row tree__row-alert alert alert-danger"
                                        // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                        style={treePadding(this.props.depth, true)}
                                    >
                                        Error loading tree: {this.state.treeOrError.message}
                                    </div>
                                ) : (
                                    this.state.treeOrError &&
                                    this.state.treeOrError.entries.map((item, i) => (
                                        <TreeLayer
                                            key={item.path}
                                            activeNode={this.props.activeNode}
                                            history={this.props.history}
                                            activePath={this.props.activePath}
                                            activePathIsTree={this.props.activePathIsTree}
                                            depth={0}
                                            index={i}
                                            isExpanded={this.props.expandedTrees.includes(item.path)}
                                            isRoot={false}
                                            expandedTrees={this.props.expandedTrees}
                                            repoPath={this.props.repoPath}
                                            rev={this.props.rev}
                                            commitID={this.props.commitID}
                                            entryInfo={item}
                                            parent={this.node}
                                            parentPath={item.path}
                                            onSelect={this.props.onSelect}
                                            onToggleExpand={this.props.onToggleExpand}
                                            onHover={this.fetchChildContents}
                                            selectedNode={this.props.selectedNode}
                                            setChildNodes={this.setChildNode}
                                            setActiveNode={this.props.setActiveNode}
                                            isSingleChild={
                                                (this.state.treeOrError as GQL.IGitTree).entries.length === 1
                                            }
                                        />
                                    ))
                                )}
                            </td>
                        </tr>
                    </tbody>
                </table>
            )
        }

        // Every other layer is a row in the file tree, and will fetch and render its children (if any) when expanded.
        return (
            <div>
                <table className="tree-layer" onMouseOver={entryInfo.isDirectory ? this.invokeOnHover : undefined}>
                    <tbody>
                        {entryInfo.isDirectory ? (
                            <>
                                <tr key={entryInfo.path} className={className} onClick={this.handleTreeClick}>
                                    <td className="tree__cell">
                                        <div
                                            className="tree__row-contents tree__row-contents-new"
                                            data-tree-is-directory="true"
                                            data-tree-path={entryInfo.path}
                                        >
                                            <div className="tree__row-contents-text">
                                                <a
                                                    className="tree__row-icon"
                                                    href={entryInfo.url}
                                                    onClick={this.noopRowClick}
                                                    // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                                    style={treePadding(this.props.depth, true)}
                                                    tabIndex={-1}
                                                >
                                                    {this.props.isExpanded ? (
                                                        <ChevronDownIcon className="icon-inline" />
                                                    ) : (
                                                        <ChevronRightIcon className="icon-inline" />
                                                    )}
                                                </a>
                                                <Link
                                                    to={entryInfo.url}
                                                    onClick={this.linkRowClick}
                                                    className="tree__row-label"
                                                    draggable={false}
                                                    title={entryInfo.path}
                                                    tabIndex={-1}
                                                >
                                                    {entryInfo.name}
                                                </Link>
                                            </div>
                                            {this.state.treeOrError === LOADING && (
                                                <div className="tree__row-loader">
                                                    <LoadingSpinner className="icon-inline tree-page__entries-loader" />
                                                </div>
                                            )}
                                        </div>
                                        {this.props.index === maxEntries - 1 && (
                                            <div
                                                className="tree__row-alert alert alert-warning"
                                                // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                                style={treePadding(this.props.depth, true)}
                                            >
                                                Too many entries. Use search to find a specific file.
                                            </div>
                                        )}
                                    </td>
                                </tr>
                                {this.props.isExpanded &&
                                    this.state.treeOrError !== LOADING && (
                                        <tr>
                                            <td className="tree__cell">
                                                {isErrorLike(this.state.treeOrError) ? (
                                                    <div
                                                        className="tree__row-alert alert alert-danger"
                                                        // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                                        style={treePadding(this.props.depth, true)}
                                                    >
                                                        Error loading file tree: {this.state.treeOrError.message}
                                                    </div>
                                                ) : (
                                                    this.state.treeOrError &&
                                                    this.state.treeOrError.entries.map((item, i) => (
                                                        <TreeLayer
                                                            key={item.path}
                                                            history={this.props.history}
                                                            activePath={this.props.activePath}
                                                            activePathIsTree={this.props.activePathIsTree}
                                                            activeNode={this.props.activeNode}
                                                            depth={this.props.depth + 1}
                                                            expandedTrees={this.props.expandedTrees}
                                                            index={i}
                                                            isExpanded={this.props.expandedTrees.includes(item.path)}
                                                            isRoot={false}
                                                            parent={this.node}
                                                            parentPath={item.path}
                                                            repoPath={this.props.repoPath}
                                                            rev={this.props.rev}
                                                            commitID={this.props.commitID}
                                                            entryInfo={item}
                                                            onSelect={this.props.onSelect}
                                                            onToggleExpand={this.props.onToggleExpand}
                                                            onHover={this.props.onHover}
                                                            selectedNode={this.props.selectedNode}
                                                            setChildNodes={this.setChildNode}
                                                            setActiveNode={this.props.setActiveNode}
                                                            isSingleChild={
                                                                (this.state.treeOrError as GQL.IGitTree).entries
                                                                    .length === 1
                                                            }
                                                        />
                                                    ))
                                                )}
                                            </td>
                                        </tr>
                                    )}
                            </>
                        ) : (
                            <tr key={entryInfo.path} className={className}>
                                <td className="tree__cell">
                                    {entryInfo.submodule ? (
                                        entryInfo.url ? (
                                            <Link
                                                to={entryInfo.url}
                                                onClick={this.linkRowClick}
                                                draggable={false}
                                                title={'Submodule: ' + entryInfo.submodule.url}
                                                className="tree__row-contents"
                                                data-tree-path={entryInfo.path}
                                            >
                                                <div className="tree__row-contents-text">
                                                    <span
                                                        className="tree__row-icon"
                                                        onClick={this.noopRowClick}
                                                        // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                                        style={treePadding(this.props.depth, true)}
                                                        tabIndex={-1}
                                                    >
                                                        <RepositoryIcon className="icon-inline" />
                                                    </span>
                                                    <span className="tree__row-label">
                                                        {entryInfo.name} @ {entryInfo.submodule.commit.substr(0, 7)}
                                                    </span>
                                                </div>
                                            </Link>
                                        ) : (
                                            <div
                                                className="tree__row-contents"
                                                title={'Submodule: ' + entryInfo.submodule.url}
                                            >
                                                <div className="tree__row-contents-text">
                                                    <span
                                                        className="tree__row-icon"
                                                        // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                                        style={treePadding(this.props.depth, true)}
                                                    >
                                                        <RepositoryIcon className="icon-inline" />
                                                    </span>
                                                    <span className="tree__row-label">
                                                        {entryInfo.name} @ {entryInfo.submodule.commit.substr(0, 7)}
                                                    </span>
                                                </div>
                                            </div>
                                        )
                                    ) : (
                                        <Link
                                            className="tree__row-contents"
                                            to={entryInfo.url}
                                            onClick={this.linkRowClick}
                                            data-tree-path={entryInfo.path}
                                            draggable={false}
                                            title={entryInfo.path}
                                            // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                            style={treePadding(this.props.depth, false)}
                                            tabIndex={-1}
                                        >
                                            {entryInfo.name}
                                        </Link>
                                    )}
                                    {this.props.index === maxEntries - 1 && (
                                        <div
                                            className="tree__row-alert alert alert-warning"
                                            // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                            style={treePadding(this.props.depth, true)}
                                        >
                                            Too many entries. Use search to find a specific file.
                                        </div>
                                    )}
                                </td>
                            </tr>
                        )}
                    </tbody>
                </table>
            </div>
        )
    }

    /**
     * Prefetches the children of hovered tree rows. Gets passed from the root tree layer to child tree layers
     * through the onHover prop. This method only gets called on the root tree layer component so we can debounce
     * the hover prefetch requests.
     */
    private fetchChildContents = (path: string): void => {
        if (this.props.isRoot) {
            this.rowHovers.next(path)
        } else {
            console.error('fetchChildContents should not be called from non-root tree layer components')
        }
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

    private handleTreeClick = () => {
        this.props.onSelect(this.node)
        const path = this.props.entryInfo ? this.props.entryInfo.path : ''
        this.props.onToggleExpand(path, !this.props.isExpanded, this.node)
    }

    /**
     * noopRowClick is the click handler for <a> rows of the tree element
     * that shouldn't update URL on click w/o modifier key (but should retain
     * anchor element properties, like right click "Copy link address").
     */
    private noopRowClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
        if (!e.altKey && !e.metaKey && !e.shiftKey && !e.ctrlKey) {
            e.preventDefault()
            e.stopPropagation()
        }
        this.props.onSelect(this.node)
        const path = this.props.entryInfo ? this.props.entryInfo.path : ''
        this.props.onToggleExpand(path, !this.props.isExpanded, this.node)
    }

    /**
     * linkRowClick is the click handler for <Link>
     */
    private linkRowClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
        this.props.setActiveNode(this.node)
        this.props.onSelect(this.node)
    }

    private setChildNode = (node: TreeNode, index: number) => {
        this.node.childNodes[index] = node
    }
}
