import { Loader } from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import * as React from 'react'
import { merge, of, Subject, Subscription } from 'rxjs'
import { catchError, delay, distinctUntilChanged, share, switchMap, takeUntil } from 'rxjs/operators'
import * as GQL from '../backend/graphqlschema'
import { fetchTree } from '../repo/backend'
import { Repo } from '../repo/index'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { TreeNode } from './Tree3'
import { TreeRow } from './TreeRow'

export interface TreeLayerProps extends Repo {
    history: H.History
    activePath: string
    parent: TreeNode | null
    activePathIsDir: boolean
    repoPath: string
    depth: number
    selectedNode: TreeNode | undefined
    /** This must not be mutated */
    resolveTo: string[]
    parentPath?: string
    onSelect: (node: TreeNode) => void
    onChangeViewState: (path: string, resolveTo: boolean, node: TreeNode) => void
    /**
     * The tree loses focus when an active row is unmounted when its parent directory collapses.
     * This function sets the focus back on the tree.
     */
    focusTreeOnUnmount: () => void
}

const LOADING: 'loading' = 'loading'
export interface TreeLayerState {
    treeOrError?: typeof LOADING | GQL.ITree | ErrorLike
}

export class TreeLayer extends React.PureComponent<TreeLayerProps, TreeLayerState> {
    private subscriptions = new Subscription()
    private componentUpdates = new Subject<TreeLayerProps>()

    public node: TreeNode

    constructor(props: TreeLayerProps) {
        super(props)
        this.node = {
            index: 0,
            parent: this.props.parent,
            childNodes: [],
            path: '',
        }

        this.state = {}
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    // Only fetch tree contents if these props change, as it shouldn't be fetched on every prop change.
                    // For example, if the active path changes (i.e. a user selects a file), we don't need to re-fetch the
                    // contents of this tree layer, because it does not change. On the other hand, if the parent path changes,
                    // the tree that we show changes, so we would have to re-fetch the tree contents.
                    distinctUntilChanged(
                        (x, y) =>
                            x.repoPath === y.repoPath &&
                            x.rev === y.rev &&
                            x.parentPath === y.parentPath &&
                            x.resolveTo === y.resolveTo
                    ),
                    switchMap(props => {
                        const treeFetch = fetchTree({
                            repoPath: props.repoPath,
                            rev: props.rev || '',
                            filePath: props.parentPath || '',
                        }).pipe(catchError(err => [asError(err)]), share())
                        return merge(treeFetch, of(LOADING).pipe(delay(300), takeUntil(treeFetch)))
                    })
                )
                .subscribe(treeOrError => this.setState({ treeOrError }), err => console.error(err))
        )
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(prevProps: TreeLayerProps): void {
        // Reset the childNodes of TreeLayer to none if the parent path changes, so we don't have children of past visited layers in the childNodes.
        if (prevProps.parentPath !== this.props.parentPath) {
            this.node.childNodes = []
        }

        // Make a fetch for tree contents when the parent path changes.
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.props.focusTreeOnUnmount()
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <table className="tree-layer">
                <tbody>
                    <tr>
                        <td>
                            {this.state.treeOrError === LOADING ? (
                                this.props.depth > 0 ? (
                                    <div className="tree__row-loader">
                                        <Loader className="icon-inline directory-page__entries-loader" />
                                    </div>
                                ) : (
                                    <div>
                                        <Loader className="icon-inline directory-page__entries-loader" /> Loading files
                                        and directories
                                    </div>
                                )
                            ) : isErrorLike(this.state.treeOrError) ? (
                                <div className="alert alert-danger">
                                    Error loading file tree: {this.state.treeOrError.message}
                                </div>
                            ) : (
                                this.state.treeOrError &&
                                [...this.state.treeOrError.directories, ...this.state.treeOrError.files].map(
                                    (item, i) => (
                                        <TreeRow
                                            index={i}
                                            parent={this.props.parent !== null ? this.node.parent : this.node}
                                            key={item.path}
                                            history={this.props.history}
                                            activePath={this.props.activePath}
                                            activePathIsDir={this.props.activePathIsDir}
                                            repoPath={this.props.repoPath}
                                            rev={this.props.rev}
                                            selectedNode={this.props.selectedNode}
                                            resolveTo={this.props.resolveTo}
                                            depth={this.props.depth}
                                            node={item}
                                            isExpanded={this.props.resolveTo.includes(item.path)}
                                            onChangeViewState={this.props.onChangeViewState}
                                            onSelect={this.props.onSelect}
                                            setChildNodes={this.setChildNode}
                                            focusTreeOnUnmount={this.props.focusTreeOnUnmount}
                                        />
                                    )
                                )
                            )}
                        </td>
                    </tr>
                </tbody>
            </table>
        )
    }

    public setChildNode = (node: TreeNode, index: number) => {
        this.node.childNodes[index] = node
    }
}
