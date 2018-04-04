// tslint:disable:no-use-before-declare

import ChevronDownIcon from '@sourcegraph/icons/lib/ChevronDown'
import ChevronRightIcon from '@sourcegraph/icons/lib/ChevronRight'
import * as H from 'history'
import isEqual from 'lodash/isEqual'
import sortBy from 'lodash/sortBy'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { skip } from 'rxjs/operators/skip'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Repo } from '../repo/index'
import { dirname } from '../util/path'
import { toBlobURL, toTreeURL } from '../util/url'
import { FileStat, toFileStat } from './fileStat'
import { scrollIntoView } from './util'

interface Props extends Repo {
    history: H.History
    scrollRootSelector?: string

    /** All file paths, relative to the root, with no leading slashes. */
    paths: string[]

    /** The tree entry that is currently active, or '' if none (which means the root). */
    activePath: string

    /** Whether the active path is a directory (including the root directory). False if it is a file. */
    activePathIsDir: boolean
}

interface State {
    /** The tree entry that is currently selected, or undefined if same as Props.activePath. */
    selectedPath: string | undefined

    /**
     * The root of the tree to show, or undefined for the root.
     *
     * This is initially the directory containing the first file viewed, but it can be changed to be an ancestor of
     * itself if the user browses to higher levels in the tree.
     */
    parentPath?: string | undefined

    /** Directories (including descendents multiple levels below this dir) that are resolveTo. */
    resolveTo: string[]

    /** The root directory that is displayed in this component. */
    dir: FileStat | null
}

const treePadding = (depth: number, directory: boolean) => ({
    paddingLeft: depth * 12 + (directory ? 0 : 12) + 12 + 'px',
    paddingRight: '16px',
})

/**
 * Tree2 is a file tree component that intends to be faster than the Tree component (to fix
 * https://github.com/sourcegraph/sourcegraph/issues/10263).
 */
export class Tree2 extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private entryViewStateChanges = new Subject<{ path: string; expand: boolean }>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        const parentPath = (props.activePathIsDir ? props.activePath : dirname(props.activePath)) || undefined
        // console.time('toFileStat (initial)')
        this.state = {
            selectedPath: undefined,
            parentPath,
            resolveTo: [],
            dir: toFileStat(props.paths, { parentPath }),
        }
        // console.timeEnd('toFileStat (initial)')
    }

    private setStateAndRecomputeTree<K extends keyof State>(
        state:
            | ((prevState: Readonly<State>, props: Props) => Pick<State, K> | State | null)
            | (Pick<State, K> | State | null),
        callback?: () => void
    ): void {
        this.setState((prevState, props) => {
            const stateUpdate = typeof state === 'function' ? state(prevState, props) : state
            if (stateUpdate === null) {
                return null
            }
            if (stateUpdate.dir) {
                throw new Error('unexpected update of computed field dir')
            }
            // const label = `toFileStat (parentPath: ${
            //     'parentPath' in stateUpdate ? stateUpdate.parentPath : prevState.parentPath
            // }, resolveTo: ${('resolveTo' in stateUpdate ? stateUpdate.resolveTo : prevState.resolveTo).join(',')})`
            // console.time(label)
            // console.profile('toFileStat')
            stateUpdate.dir = toFileStat(props.paths, {
                parentPath: 'parentPath' in stateUpdate ? stateUpdate.parentPath : prevState.parentPath,
                resolveTo: 'resolveTo' in stateUpdate ? stateUpdate.resolveTo : prevState.resolveTo,
            })
            // console.profileEnd()
            // console.timeEnd(label)
            return stateUpdate
        })
    }

    public componentDidMount(): void {
        if (this.props.activePath) {
            setTimeout(() => {
                const el = this.locateDomNode(this.props.activePath!)
                if (el) {
                    el.scrollIntoView({ behavior: 'instant', inline: 'nearest' })
                }
            })
        }

        this.subscriptions.add(
            this.entryViewStateChanges.subscribe(({ path, expand }) => {
                this.setStateAndRecomputeTree(prevState => ({
                    resolveTo: expand ? [...prevState.resolveTo, path] : prevState.resolveTo.filter(p => p !== path),
                }))
            })
        )

        this.subscriptions.add(
            this.componentUpdates.pipe(distinctUntilChanged(isEqual), skip(1)).subscribe((props: Props) => {
                // Recompute with new paths and parent path. But if the new active path is below where we are now,
                // preserve the current parent path, so that it's easy for the user to go back up.
                const newParentPath = props.activePathIsDir ? props.activePath : dirname(props.activePath)
                this.setStateAndRecomputeTree(
                    pathEqualToOrAncestor(this.state.parentPath || '', newParentPath)
                        ? {}
                        : {
                              parentPath:
                                  (props.activePathIsDir ? props.activePath : dirname(props.activePath)) || undefined,
                          }
                )
            })
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)

        if (this.props.activePath !== nextProps.activePath && nextProps.activePath) {
            setTimeout(() => {
                const el = this.locateDomNode(nextProps.activePath)
                if (el && !this.elementInViewport(el)) {
                    el.scrollIntoView({ behavior: 'instant', inline: 'nearest' })
                }
            })
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.dir) {
            return null
        }
        return (
            <div className="tree" tabIndex={1} onKeyDown={this.onKeyDown}>
                <TreeLayer
                    history={this.props.history}
                    repoPath={this.props.repoPath}
                    rev={this.props.rev}
                    dir={this.state.dir}
                    activePath={this.props.activePath}
                    selectedPath={this.state.selectedPath}
                    depth={0}
                    resolveTo={this.state.resolveTo}
                    onChangeViewState={this.onChangeEntryViewState}
                    onSelect={this.onSelectEntry}
                />
            </div>
        )
    }

    /** Called when a tree entry is expanded or collapsed. */
    private onChangeEntryViewState = (path: string, expand: boolean): void => {
        // console.time('end-to-end')
        this.entryViewStateChanges.next({ path, expand })
    }

    /** Called when a tree entry is selected. */
    private onSelectEntry = (path: string): void =>
        this.setState(prev => ({
            selectedPath: path,
        }))

    public ArrowDown = () => {
        const loc = this.locateDomNodeInCollection(this.state.selectedPath || this.props.activePath)
        if (loc) {
            const { items, i } = loc
            if (i < items.length - 1) {
                // select next
                this.selectElement(items[i + 1])
            } else {
                // select first
                this.selectElement(items[0])
            }
        }
    }

    public ArrowUp = () => {
        const loc = this.locateDomNodeInCollection(this.state.selectedPath || this.props.activePath)
        if (loc) {
            const { items, i } = loc
            if (i > 0) {
                // select previous
                this.selectElement(items[i - 1])
            } else {
                // select last
                this.selectElement(items[items.length - 1])
            }
        }
    }

    private isExpanded(path: string): boolean {
        return this.state.resolveTo.includes(path)
    }

    public ArrowLeft = () => {
        const selectedPath = this.state.selectedPath || this.props.activePath
        const isOpenDir = this.isExpanded(selectedPath)
        if (isOpenDir) {
            this.entryViewStateChanges.next({ path: selectedPath, expand: false })
            return
        }
        const loc = this.locateDomNodeInCollection(selectedPath)
        if (loc) {
            const { items, i } = loc
            const dir = dirname(selectedPath)
            const parentDir = dir ? this.locateDomNode(dir) : undefined
            if (parentDir) {
                this.selectElement(parentDir)
                return
            }

            if (i > 0) {
                // select previous
                this.selectElement(items[i - 1])
            } else {
                // select last
                this.selectElement(items[items.length - 1])
            }
        }
    }

    public ArrowRight = () => {
        const selectedPath = this.state.selectedPath || this.props.activePath
        const loc = this.locateDomNodeInCollection(selectedPath)
        if (loc) {
            const { items, i } = loc
            const isDirectory = Boolean(items[i].getAttribute('data-tree-directory'))
            if (!this.isExpanded(selectedPath) && isDirectory) {
                // First, show the group (but don't update selection)
                this.entryViewStateChanges.next({ path: selectedPath, expand: true })
            } else {
                if (i < items.length - 1) {
                    // select next
                    this.selectElement(items[i + 1])
                } else {
                    // select first
                    this.selectElement(items[0])
                }
            }
        }
    }

    public Enter = () => {
        const selectedPath = this.state.selectedPath || this.props.activePath
        const loc = this.locateDomNodeInCollection(selectedPath)
        if (loc) {
            const { items, i } = loc
            const isDir = Boolean(items[i].getAttribute('data-tree-directory'))
            if (isDir) {
                const isOpen = this.isExpanded(selectedPath)
                if (isOpen) {
                    this.entryViewStateChanges.next({ path: selectedPath, expand: false })
                    return
                }
            }
            this.entryViewStateChanges.next({ path: selectedPath, expand: true })
            this.onSelectEntry(selectedPath) // TODO!(sqs): hacky, should combine with above line's operation
            const urlProps = {
                repoPath: this.props.repoPath,
                rev: this.props.rev,
                filePath: selectedPath,
            }
            this.props.history.push(isDir ? toTreeURL(urlProps) : toBlobURL(urlProps))
        }
    }

    private elementInViewport(el: any): boolean {
        const rect = el.getBoundingClientRect()
        return (
            rect.top >= 0 &&
            rect.left >= 0 &&
            rect.bottom <= (window.innerHeight || document.documentElement.clientHeight) /*or $(window).height() */ &&
            rect.right <= (window.innerWidth || document.documentElement.clientWidth) /*or $(window).width() */
        )
    }

    private selectElement(el: HTMLElement): void {
        const root = (this.props.scrollRootSelector
            ? document.querySelector(this.props.scrollRootSelector)
            : document.querySelector('.tree-container')) as HTMLElement
        scrollIntoView(el, root)
        const path = el.getAttribute('data-tree-path')!
        this.setState({ selectedPath: path })
    }

    private locateDomNode(path: string): HTMLElement | undefined {
        return document.querySelector(`[data-tree-path="${path}"]`) as any
    }

    private locateDomNodeInCollection(path: string): { items: HTMLElement[]; i: number } | undefined {
        const items = document.querySelectorAll('.tree__row-contents') as any
        let i = 0
        for (i; i < items.length; ++i) {
            if (items[i].getAttribute('data-tree-path') === path) {
                return { items, i }
            }
        }
        return undefined
    }

    private onKeyDown = (event: React.KeyboardEvent<HTMLElement>): void => {
        const handler = (this as any)[event.key]
        if (handler) {
            event.preventDefault()
            handler.call(this, event)
        }
    }
}

interface TreeLayerCommon extends Repo, State {
    history: H.History

    dir: FileStat

    activePath: string

    depth: number

    /** Called when a tree entry's resolveTo/collapse state is changed. */
    onChangeViewState: (path: string, resolveTo: boolean) => void

    /** Called when a tree entry is selected. */
    onSelect: (path: string) => void
}

interface TreeLayerProps extends TreeLayerCommon {}

// const endToEndLabel = 'end-to-end'
// console.time(endToEndLabel)
class TreeLayer extends React.PureComponent<TreeLayerProps, {}> {
    public render(): JSX.Element | null {
        const nodes = sortBy(this.props.dir.children, [(e: FileStat) => (e.isDirectory ? 0 : 1), 'text'])
        // if (this.props.dir.path === '.' && nodes && nodes.length > 2) {
        //     setTimeout(() => console.timeEnd(endToEndLabel))
        // }

        return (
            <table className="tree-layer">
                <tbody>
                    <tr>
                        <td>
                            {nodes &&
                                nodes.map((node, i) => (
                                    <TreeRow
                                        key={i}
                                        {...this.props}
                                        depth={this.props.depth}
                                        node={node}
                                        isExpanded={
                                            this.props.resolveTo.includes(
                                                node.path
                                            ) /* TODO!(sqs): also if resolveTo includes a descendent/descendent-sibling of node.path */
                                        }
                                        isSelected={this.props.selectedPath === node.path}
                                    />
                                ))}
                        </td>
                    </tr>
                </tbody>
            </table>
        )
    }
}

interface TreeRowProps extends TreeLayerCommon {
    node: FileStat
    depth: number
    isExpanded: boolean
    isSelected: boolean
}

class TreeRow extends React.PureComponent<TreeRowProps> {
    public render(): JSX.Element | null {
        const { node, selectedPath } = this.props
        const className = [
            'tree__row',
            node.path === selectedPath && 'tree__row--selected',
            node.path === this.props.activePath && 'tree__row--active',
        ]
            .filter(c => !!c)
            .join(' ')
        return (
            <table className="tree-row">
                <tbody>
                    {node.isDirectory ? (
                        <>
                            <tr key={node.path} className={className}>
                                <td onClick={this.handleDirClick}>
                                    <div
                                        className="tree__row-contents"
                                        data-tree-directory="true"
                                        data-tree-path={node.path}
                                    >
                                        <a
                                            className="tree__row-icon"
                                            onClick={this.noopRowClick}
                                            href={toTreeURL({
                                                repoPath: this.props.repoPath,
                                                rev: this.props.rev,
                                                filePath: node.path,
                                            })}
                                            // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                            style={treePadding(this.props.depth, true)}
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
                                                filePath: node.path,
                                            })}
                                            className="tree__row-label"
                                            draggable={false}
                                            title={node.path}
                                        >
                                            {node.name}
                                        </Link>
                                    </div>
                                </td>
                            </tr>
                            {this.props.isExpanded && (
                                <tr>
                                    <td>
                                        <TreeLayer
                                            history={this.props.history}
                                            repoPath={this.props.repoPath}
                                            rev={this.props.rev}
                                            dir={node}
                                            activePath={this.props.activePath}
                                            selectedPath={this.props.selectedPath}
                                            depth={this.props.depth + 1}
                                            resolveTo={this.props.resolveTo}
                                            onChangeViewState={this.props.onChangeViewState}
                                            onSelect={this.props.onSelect}
                                        />
                                    </td>
                                </tr>
                            )}
                        </>
                    ) : (
                        <tr key={node.path} className={className}>
                            <td>
                                <Link
                                    className="tree__row-contents"
                                    onClick={this.linkRowClick}
                                    to={toBlobURL({
                                        repoPath: this.props.repoPath,
                                        rev: this.props.rev,
                                        filePath: node.path,
                                    })}
                                    data-tree-path={node.path}
                                    draggable={false}
                                    title={node.path}
                                    // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                    style={treePadding(this.props.depth, false)}
                                >
                                    {node.name}
                                </Link>
                            </td>
                        </tr>
                    )}
                </tbody>
            </table>
        )
    }

    private handleDirClick = () => {
        this.props.onSelect(this.props.node.path)
        this.props.onChangeViewState(this.props.node.path, !this.props.isExpanded) // TODO!(sqs): combine with above
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
        this.props.onSelect(this.props.node.path)
        this.props.onChangeViewState(this.props.node.path, !this.props.isExpanded) // TODO!(sqs): combine with above
    }

    /**
     * linkRowClick is the click handler for <Link>
     */
    private linkRowClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
        this.props.onSelect(this.props.node.path)
    }
}

/** Returns whether path is an ancestor of (or equal to) candidate. */
function pathEqualToOrAncestor(path: string, candidate: string): boolean {
    return path === candidate || path === '' || candidate.startsWith(path + '/')
}
