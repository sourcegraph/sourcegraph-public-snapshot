import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
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
import * as GQL from '../backend/graphqlschema'
import { AbsoluteRepo } from '../repo'
import { fetchTreeEntries } from '../repo/backend'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { ChildTreeLayer } from './ChildTreeLayer'
import { TreeNode } from './Tree'
import { treePadding } from './util'
import { hasSingleChild, singleChildEntriesToGitTree, SingleChildGitTree } from './util'

const maxEntries = 2500

export interface TreeRootProps extends AbsoluteRepo {
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
    selectedNode: TreeNode
    onHover?: (filePath: string) => void
    onSelect: (node: TreeNode) => void
    onToggleExpand: (path: string, expanded: boolean, node: TreeNode) => void
    setChildNodes: (node: TreeNode, index: number) => void
    setActiveNode: (node: TreeNode) => void
}

const LOADING: 'loading' = 'loading'
interface TreeRootState {
    treeOrError?: typeof LOADING | GQL.IGitTree | ErrorLike
}

export class TreeRoot extends React.Component<TreeRootProps, TreeRootState> {
    public node: TreeNode
    private subscriptions = new Subscription()
    private componentUpdates = new Subject<TreeRootProps>()
    private rowHovers = new Subject<string>()

    constructor(props: TreeRootProps) {
        super(props)
        this.node = {
            index: this.props.index,
            parent: this.props.parent,
            childNodes: [],
            path: '',
            url: '',
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
                            x.isExpanded === y.isExpanded &&
                            x.location === y.location
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

        // This handles pre-fetching when a user
        // hovers over a directory. The `subscribe` is empty because
        // we simply want to cache the network request.
        this.subscriptions.add(
            this.rowHovers
                .pipe(
                    debounceTime(100),
                    mergeMap(path =>
                        fetchTreeEntries({
                            repoPath: this.props.repoPath,
                            rev: this.props.rev,
                            commitID: this.props.commitID,
                            filePath: path,
                            first: maxEntries,
                        }).pipe(catchError(err => [asError(err)]))
                    )
                )
                .subscribe()
        )

        // When we're at the root tree layer, fetch the tree contents on mount.
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(prevProps: TreeRootProps): void {
        this.componentUpdates.next(this.props)
    }
    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const { treeOrError } = this.state

        let singleChildTreeEntry = {} as SingleChildGitTree
        if (
            treeOrError &&
            treeOrError !== LOADING &&
            !isErrorLike(treeOrError) &&
            hasSingleChild(treeOrError.entries)
        ) {
            singleChildTreeEntry = singleChildEntriesToGitTree(treeOrError.entries)
        }

        return (
            <table className="tree-layer" tabIndex={0}>
                <tbody>
                    <tr>
                        <td className="tree__cell">
                            {treeOrError === LOADING ? (
                                <div className="tree__row-loader">
                                    <LoadingSpinner className="icon-inline tree-page__entries-loader" />Loading tree
                                </div>
                            ) : isErrorLike(treeOrError) ? (
                                <div
                                    className="tree__row tree__row-alert alert alert-danger"
                                    // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                    style={treePadding(this.props.depth, true)}
                                >
                                    Error loading tree: {treeOrError.message}
                                </div>
                            ) : (
                                treeOrError && (
                                    <ChildTreeLayer
                                        {...this.props}
                                        parent={this.node}
                                        depth={-1 as number}
                                        entries={treeOrError.entries}
                                        singleChildTreeEntry={singleChildTreeEntry}
                                        childrenEntries={singleChildTreeEntry.children}
                                        onHover={this.fetchChildContents}
                                        setChildNodes={this.setChildNode}
                                    />
                                )
                            )}
                        </td>
                    </tr>
                </tbody>
            </table>
        )
    }
    /**
     * Prefetches the children of hovered tree rows. Gets passed from the root tree layer to child tree layers
     * through the onHover prop. This method only gets called on the root tree layer component so we can debounce
     * the hover prefetch requests.
     */
    private fetchChildContents = (path: string): void => {
        this.rowHovers.next(path)
    }
    private setChildNode = (node: TreeNode, index: number) => {
        this.node.childNodes[index] = node
    }
}
