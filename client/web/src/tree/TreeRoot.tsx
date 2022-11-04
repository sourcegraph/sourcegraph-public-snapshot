/* eslint jsx-a11y/no-noninteractive-tabindex: warn*/
import * as React from 'react'

import * as H from 'history'
import { EMPTY, merge, of, Subject, Subscription } from 'rxjs'
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

import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { FileDecorationsByPath } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { fetchTreeEntries } from '@sourcegraph/shared/src/backend/repo'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Scalars, TreeFields } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { AbsoluteRepo } from '@sourcegraph/shared/src/util/url'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { getFileDecorations } from '../backend/features'
import { requestGraphQL } from '../backend/graphql'

import { ChildTreeLayer } from './ChildTreeLayer'
import { TreeLayerTable, TreeLayerCell, TreeRowAlert } from './components'
import { MAX_TREE_ENTRIES } from './constants'
import { TreeNode } from './Tree'
import { TreeRootContext } from './TreeContext'
import { hasSingleChild, compareTreeProps, singleChildEntriesToGitTree, SingleChildGitTree } from './util'

import styles from './Tree.module.scss'

const errorWidth = (width?: string): { width: string } => ({
    width: width ? `${width}px` : 'auto',
})

export interface TreeRootProps extends AbsoluteRepo, ExtensionsControllerProps, ThemeProps, TelemetryProps {
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
    sizeKey: string
    onHover?: (filePath: string) => void
    onSelect: (node: TreeNode) => void
    onToggleExpand: (path: string, expanded: boolean, node: TreeNode) => void
    setChildNodes: (node: TreeNode, index: number) => void
    setActiveNode: (node: TreeNode) => void
    repoID: Scalars['ID']
    enableMergedFileSymbolSidebar: boolean
}

const LOADING = 'loading' as const
interface TreeRootState {
    treeOrError?: typeof LOADING | TreeFields | ErrorLike
    fileDecorationsByPath: FileDecorationsByPath
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
        this.state = {
            fileDecorationsByPath: {},
        }
    }

    public componentDidMount(): void {
        // Set this row as a childNode of its TreeLayer parent
        this.props.setChildNodes(this.node, this.node.index)

        const treeOrErrors = this.componentUpdates.pipe(
            distinctUntilChanged(compareTreeProps),
            filter(props => props.isExpanded),
            switchMap(props => {
                const treeFetch = fetchTreeEntries({
                    repoName: props.repoName,
                    revision: props.revision,
                    commitID: props.commitID,
                    filePath: props.parentPath || '',
                    first: MAX_TREE_ENTRIES,
                    requestGraphQL: ({ request, variables }) => requestGraphQL(request, variables),
                }).pipe(
                    catchError(error => [asError(error)]),
                    share()
                )
                return merge(treeFetch, of(LOADING).pipe(delay(300), takeUntil(treeFetch)))
            })
        )

        this.subscriptions.add(
            treeOrErrors.subscribe(
                treeOrError => {
                    // clear file decorations before latest file decorations come
                    this.setState({ treeOrError, fileDecorationsByPath: {} })
                },
                error => console.error(error)
            )
        )

        this.subscriptions.add(
            treeOrErrors
                .pipe(
                    switchMap(treeOrError =>
                        treeOrError !== 'loading' && !isErrorLike(treeOrError)
                            ? getFileDecorations({
                                  files: treeOrError.entries,
                                  repoName: this.props.repoName,
                                  commitID: this.props.commitID,
                                  extensionsController: this.props.extensionsController,
                                  parentNodeUri: treeOrError.url,
                              })
                            : EMPTY
                    )
                )
                .subscribe(fileDecorationsByPath => {
                    this.setState({ fileDecorationsByPath })
                })
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
                            repoName: this.props.repoName,
                            revision: this.props.revision,
                            commitID: this.props.commitID,
                            filePath: path,
                            first: MAX_TREE_ENTRIES,
                            requestGraphQL: ({ request, variables }) => requestGraphQL(request, variables),
                        }).pipe(catchError(error => [asError(error)]))
                    )
                )
                .subscribe()
        )

        // When we're at the root tree layer, fetch the tree contents on mount.
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }
    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
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
            <>
                {isErrorLike(treeOrError) ? (
                    <TreeRowAlert
                        // needed because of dynamic styling
                        style={errorWidth(localStorage.getItem(this.props.sizeKey) ? this.props.sizeKey : undefined)}
                        prefix="Error loading tree"
                        error={treeOrError}
                    />
                ) : (
                    /**
                     * TODO: Improve accessibility here.
                     * We should not be stealing focus here, we should let the user focus on the actual items listed.
                     * Issue: https://github.com/sourcegraph/sourcegraph/issues/19167
                     */
                    <TreeLayerTable tabIndex={0}>
                        <tbody>
                            <tr>
                                <TreeLayerCell>
                                    {treeOrError === LOADING ? (
                                        <div className={styles.treeLoadingSpinner}>
                                            <LoadingSpinner className="tree-page__entries-loader mr-2" />
                                            Loading tree
                                        </div>
                                    ) : (
                                        treeOrError && (
                                            <TreeRootContext.Provider
                                                value={{
                                                    rootTreeUrl: treeOrError.url,
                                                    repoID: this.props.repoID,
                                                    repoName: this.props.repoName,
                                                    revision: this.props.revision,
                                                    commitID: this.props.commitID,
                                                }}
                                            >
                                                <ChildTreeLayer
                                                    {...this.props}
                                                    parent={this.node}
                                                    depth={-1 as number}
                                                    entries={treeOrError.entries}
                                                    singleChildTreeEntry={singleChildTreeEntry}
                                                    childrenEntries={singleChildTreeEntry.children}
                                                    onHover={this.fetchChildContents}
                                                    setChildNodes={this.setChildNode}
                                                    fileDecorationsByPath={this.state.fileDecorationsByPath}
                                                    enableMergedFileSymbolSidebar={
                                                        this.props.enableMergedFileSymbolSidebar
                                                    }
                                                />
                                            </TreeRootContext.Provider>
                                        )
                                    )}
                                </TreeLayerCell>
                            </tr>
                        </tbody>
                    </TreeLayerTable>
                )}
            </>
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
    private setChildNode = (node: TreeNode, index: number): void => {
        this.node.childNodes[index] = node
    }
}
