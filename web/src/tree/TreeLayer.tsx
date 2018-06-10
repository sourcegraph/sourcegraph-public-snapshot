import ChevronDownIcon from '@sourcegraph/icons/lib/ChevronDown'
import ChevronRightIcon from '@sourcegraph/icons/lib/ChevronRight'
import { Loader } from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
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
import { fetchTreeEntries } from '../repo/backend'
import { Repo } from '../repo/index'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { toBlobURL, toTreeURL } from '../util/url'
import { TreeNode } from './Tree3'

interface TreeLayerProps extends Repo {
    history: H.History
    activeNode: TreeNode
    activePath: string
    activePathIsDir: boolean
    depth: number
    expandedDirectories: string[]
    parent: TreeNode | null
    parentPath?: string
    index: number
    isExpanded: boolean
    isRoot: boolean
    fileOrDirectoryInfo?: GQL.IFile | GQL.IDirectory
    selectedNode: TreeNode
    onHover?: (filePath: string) => void
    onSelect: (node: TreeNode) => void
    onToggleExpand: (path: string, expanded: boolean, node: TreeNode) => void
    setChildNodes: (node: TreeNode, index: number) => void
    setActiveNode: (node: TreeNode) => void
}

const LOADING: 'loading' = 'loading'
interface TreeLayerState {
    treeOrError?: typeof LOADING | GQL.ITree | ErrorLike
}

const treePadding = (depth: number, directory: boolean) => ({
    paddingLeft: depth * 12 + (directory ? 0 : 12) + 12 + 'px',
    paddingRight: '16px',
})

const maxFilesOrDirs = 2500

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
            path: this.props.fileOrDirectoryInfo ? this.props.fileOrDirectoryInfo.path : '',
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
                            x.parentPath === y.parentPath &&
                            x.expandedDirectories === y.expandedDirectories
                    ),
                    filter(props => props.isExpanded),
                    switchMap(props => {
                        const treeFetch = fetchTreeEntries({
                            repoPath: props.repoPath,
                            rev: props.rev || '',
                            filePath: props.parentPath || '',
                            first: maxFilesOrDirs,
                        }).pipe(catchError(err => [asError(err)]), share())
                        return merge(treeFetch, of(LOADING).pipe(delay(300), takeUntil(treeFetch)))
                    })
                )
                .subscribe(treeOrError => this.setState({ treeOrError }), err => console.error(err))
        )

        // When we're at the root tree layer or the dir is already expanded, fetch the tree contents on mount.
        // For other layers, fetch on hover or on expand.
        if (this.props.isRoot || this.props.isExpanded) {
            this.componentUpdates.next(this.props)
        }

        // If navigating directly to a file or directory, set the correct active node.
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
                            rev: this.props.rev || '',
                            filePath: path || '',
                            first: maxFilesOrDirs,
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

        // If the file or directory being viewed changes, set the new active node.
        if (prevProps.activePath !== this.props.activePath && this.node.path === this.props.activePath) {
            this.props.setActiveNode(this.node)
        }

        this.componentUpdates.next(this.props)

        const fileOrDir = this.props.fileOrDirectoryInfo && this.props.fileOrDirectoryInfo.isDirectory
        // When scrolling through the tree with the keyboard, if we scroll over a directory node, prefetch the contents.
        if (this.node === this.props.selectedNode && fileOrDir && this.props.onHover) {
            this.props.onHover(this.node.path)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const fileOrDirInfo = this.props.fileOrDirectoryInfo
        const className = [
            'tree__row',
            this.props.isExpanded && 'tree__row--expanded',
            this.node === this.props.activeNode && 'tree__row--active',
            this.node === this.props.selectedNode && 'tree__row--selected',
        ]
            .filter(c => !!c)
            .join(' ')

        // If isRoot or there's no file or directory info, we are at the root layer, so simply load all top-level directories and files.
        if (this.props.isRoot || !fileOrDirInfo) {
            return (
                <table className="tree-layer" tabIndex={0}>
                    <tbody>
                        <tr>
                            <td className="tree__cell">
                                {this.state.treeOrError === LOADING ? (
                                    <div className="tree__row-loader">
                                        <Loader className="icon-inline directory-page__entries-loader" />Loading files
                                        and directories
                                    </div>
                                ) : isErrorLike(this.state.treeOrError) ? (
                                    <div
                                        className="tree__row tree__row-alert alert alert-danger"
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
                                            activeNode={this.props.activeNode}
                                            history={this.props.history}
                                            activePath={this.props.activePath}
                                            activePathIsDir={this.props.activePathIsDir}
                                            depth={0}
                                            index={i}
                                            isExpanded={this.props.expandedDirectories.includes(item.path)}
                                            isRoot={false}
                                            expandedDirectories={this.props.expandedDirectories}
                                            repoPath={this.props.repoPath}
                                            rev={this.props.rev}
                                            fileOrDirectoryInfo={item}
                                            parent={this.node}
                                            parentPath={item.path}
                                            onSelect={this.props.onSelect}
                                            onToggleExpand={this.props.onToggleExpand}
                                            onHover={this.fetchChildContents}
                                            selectedNode={this.props.selectedNode}
                                            setChildNodes={this.setChildNode}
                                            setActiveNode={this.props.setActiveNode}
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
                <table className="tree-layer" onMouseOver={fileOrDirInfo.isDirectory ? this.invokeOnHover : undefined}>
                    <tbody>
                        {fileOrDirInfo.isDirectory ? (
                            <>
                                <tr key={fileOrDirInfo.path} className={className} onClick={this.handleDirClick}>
                                    <td className="tree__cell">
                                        <div
                                            className="tree__row-contents tree__row-contents-new"
                                            data-tree-directory="true"
                                            data-tree-path={fileOrDirInfo.path}
                                        >
                                            <div className="tree__row-contents-text">
                                                <a
                                                    className="tree__row-icon"
                                                    href={toTreeURL({
                                                        repoPath: this.props.repoPath,
                                                        rev: this.props.rev,
                                                        filePath: fileOrDirInfo.path,
                                                    })}
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
                                                    to={toTreeURL({
                                                        repoPath: this.props.repoPath,
                                                        rev: this.props.rev,
                                                        filePath: fileOrDirInfo.path,
                                                    })}
                                                    onClick={this.linkRowClick}
                                                    className="tree__row-label"
                                                    draggable={false}
                                                    title={fileOrDirInfo.path}
                                                    tabIndex={-1}
                                                >
                                                    {fileOrDirInfo.name}
                                                </Link>
                                            </div>
                                            {this.state.treeOrError === LOADING && (
                                                <div className="tree__row-loader">
                                                    <Loader className="icon-inline directory-page__entries-loader" />
                                                </div>
                                            )}
                                        </div>
                                        {this.props.index === maxFilesOrDirs - 1 && (
                                            <div
                                                className="tree__row-alert alert alert-warning"
                                                // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                                style={treePadding(this.props.depth, true)}
                                            >
                                                Too many entries in this directory. Use search to find a specific file.
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
                                                            activePathIsDir={this.props.activePathIsDir}
                                                            activeNode={this.props.activeNode}
                                                            depth={this.props.depth + 1}
                                                            expandedDirectories={this.props.expandedDirectories}
                                                            index={i}
                                                            isExpanded={this.props.expandedDirectories.includes(
                                                                item.path
                                                            )}
                                                            isRoot={false}
                                                            parent={this.node}
                                                            parentPath={item.path}
                                                            repoPath={this.props.repoPath}
                                                            rev={this.props.rev}
                                                            fileOrDirectoryInfo={item}
                                                            onSelect={this.props.onSelect}
                                                            onToggleExpand={this.props.onToggleExpand}
                                                            onHover={this.props.onHover}
                                                            selectedNode={this.props.selectedNode}
                                                            setChildNodes={this.setChildNode}
                                                            setActiveNode={this.props.setActiveNode}
                                                        />
                                                    ))
                                                )}
                                            </td>
                                        </tr>
                                    )}
                            </>
                        ) : (
                            <tr key={fileOrDirInfo.path} className={className}>
                                <td className="tree__cell">
                                    <Link
                                        className="tree__row-contents"
                                        to={toBlobURL({
                                            repoPath: this.props.repoPath,
                                            rev: this.props.rev,
                                            filePath: fileOrDirInfo.path,
                                        })}
                                        onClick={this.linkRowClick}
                                        data-tree-path={fileOrDirInfo.path}
                                        draggable={false}
                                        title={fileOrDirInfo.path}
                                        // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                        style={treePadding(this.props.depth, false)}
                                        tabIndex={-1}
                                    >
                                        {fileOrDirInfo.name}
                                    </Link>
                                    {this.props.index === maxFilesOrDirs - 1 && (
                                        <div
                                            className="tree__row-alert alert alert-warning"
                                            // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                            style={treePadding(this.props.depth, true)}
                                        >
                                            Too many entries in this directory. Use search to find a specific file.
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
     * Prefetches the directory contents of hovered tree rows. Gets passed from the root tree layer to child tree layers
     * through the onHover prop. This method only gets called on the root tree layer component
     * so we can debounce the hover prefetch requests.
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

    private handleDirClick = () => {
        this.props.onSelect(this.node)
        const path = this.props.fileOrDirectoryInfo ? this.props.fileOrDirectoryInfo.path : ''
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
        const path = this.props.fileOrDirectoryInfo ? this.props.fileOrDirectoryInfo.path : ''
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
