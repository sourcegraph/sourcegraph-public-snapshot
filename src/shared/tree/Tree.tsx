// tslint:disable:no-use-before-declare

import flatten from 'lodash/flatten'
import groupBy from 'lodash/groupBy'
import sortBy from 'lodash/sortBy'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import * as React from 'react'
import { BehaviorSubject, Subscription } from 'rxjs'
import { Repo } from '../repo'
import { toTreeURL } from '../util/url'
import { getParentDir, scrollIntoView } from './util'

export interface Props extends Repo {
    paths: string[]
    scrollRootSelector?: string
    selectedPath: string
}

interface TreeNodeState {
    collapsed: boolean
    selected: boolean
}

interface TreeNode {
    filePath: string
    children: TreeNode[]
    state: BehaviorSubject<TreeNodeState>
}

interface Store {
    nodes: TreeNode[]
    nodeMap: Map<string, TreeNode>
    selectedPath: string
}

interface State {
    store: Store
    relativeDir?: string
}

const treePadding = (depth: number, directory: boolean) => ({
    paddingLeft: depth * 12 + (directory ? 0 : 12) + 12 + 'px',
    paddingRight: '16px',
})

function selectRow(store: Store, path: string): void {
    const currSelectedNode = store.nodeMap.get(store.selectedPath)
    if (currSelectedNode) {
        currSelectedNode.state.next({ ...currSelectedNode.state.getValue(), selected: false })
    }
    store.selectedPath = path
    const nextSelectedNode = store.nodeMap.get(path)
    if (nextSelectedNode) {
        nextSelectedNode.state.next({ ...nextSelectedNode.state.getValue(), selected: true })
    }
}

function closeDirectory(store: Store, dir: string): void {
    selectRow(store, dir)
    const node = store.nodeMap.get(dir)
    if (node) {
        node.state.next({ ...node.state.getValue(), collapsed: true })
    } else {
        console.error('could not locate node', dir)
    }
}

export class Tree extends React.PureComponent<Props, State> {
    constructor(props: Props) {
        super(props)
        const { nodes, nodeMap } = this.parseNodes(props.paths, props.selectedPath)
        this.state = {
            store: { nodes, nodeMap, selectedPath: props.selectedPath },
            relativeDir: nodes.length > 0 ? getParentDir(nodes[0].filePath) : undefined,
        }
    }

    public componentDidMount(): void {
        if (this.props.selectedPath) {
            const el = this.locateDomNode(this.props.selectedPath!)
            if (el && !this.elementInViewport(el)) {
                el.scrollIntoView({ behavior: 'instant', inline: 'nearest' })
            }
        }
    }

    public componentWillReceiveProps(nextProps: Props): void {
        const selectedPath = nextProps.selectedPath
        if (this.props.paths !== nextProps.paths) {
            const { nodes, nodeMap } = this.parseNodes(nextProps.paths, selectedPath)
            this.setState({
                store: { nodes, nodeMap, selectedPath },
                relativeDir: nodes.length > 0 ? getParentDir(nodes[0].filePath) : undefined,
            })
        } else {
            // If we are trying to show a path not available on the tree, recreate the nodes.
            const loc = this.locateDomNodeInCollection(selectedPath)
            if (!loc) {
                const { nodes, nodeMap } = this.parseNodes(nextProps.paths, selectedPath)
                this.setState({
                    store: { nodes, nodeMap, selectedPath },
                    relativeDir: nodes.length > 0 ? getParentDir(nodes[0].filePath) : undefined,
                })
            } else {
                selectRow(this.state.store, selectedPath)
            }
        }
        if (this.props.selectedPath !== selectedPath) {
            setTimeout(() => {
                if (selectedPath) {
                    const el = this.locateDomNode(selectedPath)
                    if (el && !this.elementInViewport(el)) {
                        el.scrollIntoView({ behavior: 'instant', inline: 'nearest' })
                    }
                }
            }, 250)
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className="github-tree" tabIndex={1} onKeyDown={this.onKeyDown}>
                <TreeLayer
                    repoPath={this.props.repoPath}
                    rev={this.props.rev}
                    store={this.state.store}
                    currSubpath=""
                    relativeDir={this.state.relativeDir}
                />
            </div>
        )
    }

    public ArrowDown = () => {
        const loc = this.locateDomNodeInCollection(this.state.store.selectedPath)
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
        const loc = this.locateDomNodeInCollection(this.state.store.selectedPath)
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

    public ArrowLeft = () => {
        const selectedPath = this.state.store.selectedPath
        const node = this.state.store.nodeMap.get(selectedPath)
        if (!node) {
            console.error('could not locate node (arrow down)', selectedPath)
            return
        }
        const isOpenDir = !node.state.getValue().collapsed
        if (isOpenDir) {
            closeDirectory(this.state.store, selectedPath)
            return
        }
        const loc = this.locateDomNodeInCollection(selectedPath)
        if (loc) {
            const { items, i } = loc
            const pathSplit = selectedPath.split('/')
            const dir = pathSplit.splice(0, pathSplit.length - 1).join('/')
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
        const selectedPath = this.state.store.selectedPath
        const node = this.state.store.nodeMap.get(selectedPath)
        if (!node) {
            console.error('could not locate node (arrow right)', selectedPath)
            return
        }
        const loc = this.locateDomNodeInCollection(selectedPath)
        if (loc) {
            const { items, i } = loc
            const isDirectory = Boolean(items[i].getAttribute('data-tree-directory'))
            if (node.state.getValue().collapsed && isDirectory) {
                // First, show the group (but don't update selection)
                node.state.next({ collapsed: false, selected: true })
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
        const selectedPath = this.state.store.selectedPath
        const node = this.state.store.nodeMap.get(selectedPath)
        if (!node) {
            console.error('could not locate node (enter)', selectedPath)
            return
        }
        const loc = this.locateDomNodeInCollection(selectedPath)
        if (loc) {
            const { items, i } = loc
            const isDir = Boolean(items[i].getAttribute('data-tree-directory'))
            if (isDir) {
                const isOpen = !node.state.getValue().collapsed
                if (isOpen) {
                    closeDirectory(this.state.store, selectedPath)
                    return
                }
            }
            node.state.next({ collapsed: false, selected: true })
            const element = this.locateDomNode(selectedPath) as HTMLElement
            if (element && !isDir) {
                element.click()
            }
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
        if (!this.elementInViewport(el)) {
            scrollIntoView(el, root)
        }
        const path = el.getAttribute('data-tree-path')!
        selectRow(this.state.store, path)
    }

    private locateDomNode(path: string): HTMLElement | undefined {
        return document.querySelector(`[data-tree-path = '${path}']`) as any
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

    private parseNodes = (paths: string[], selectedPath: string) => {
        const getFilePath = (prefix: string, restParts: string[]) => {
            if (prefix === '') {
                return restParts.join('/')
            }
            return prefix + '/' + restParts.join('/')
        }

        const parseHelper = (
            splits: string[][],
            subpath = '',
            nodeMap = new Map<string, TreeNode>()
        ): { nodes: TreeNode[]; nodeMap: Map<string, TreeNode> } => {
            const splitsByDir = groupBy(splits, split => {
                if (split.length === 1) {
                    return ''
                }
                return split[0]
            })
            const entries = flatten<TreeNode>(
                Object.entries(splitsByDir).map(([dir, pathSplits]) => {
                    if (dir === '') {
                        return pathSplits.map(split => {
                            const filePath = getFilePath(subpath, split)
                            const node: TreeNode = {
                                children: [],
                                filePath,
                                state: new BehaviorSubject<TreeNodeState>({
                                    selected: filePath === selectedPath,
                                    collapsed: true,
                                }),
                            }
                            nodeMap.set(filePath, node)
                            return node
                        })
                    }
                    const splits = subpath ? subpath + '/' + dir : dir

                    const dirPath = getFilePath(subpath, [dir])
                    const dirNode: TreeNode = {
                        children: parseHelper(
                            pathSplits.map(split => split.slice(1)),
                            subpath ? subpath + '/' + dir : dir,
                            nodeMap
                        ).nodes,
                        filePath: dirPath,
                        state: new BehaviorSubject<TreeNodeState>({
                            selected: dirPath === selectedPath,
                            collapsed: selectedPath.indexOf(splits) === -1,
                        }),
                    }
                    nodeMap.set(dirPath, dirNode)
                    return [dirNode]
                })
            )

            // directories first (nodes w/ children), then sort lexicographically
            return { nodes: sortBy(entries, [(e: TreeNode) => (e.children.length > 0 ? 0 : 1), 'text']), nodeMap }
        }

        return parseHelper(paths.map(path => path.split('/')))
    }
}

interface TreeLayerProps extends Repo {
    currSubpath: string
    relativeDir?: string
    store: Store
}

class TreeLayer extends React.PureComponent<TreeLayerProps, {}> {
    public getDepth(): number {
        if (this.props.currSubpath === '') {
            return 0
        }
        const subpathDepth = this.props.currSubpath.split('/').length
        if (this.props.relativeDir) {
            const relativeDirDepth = this.props.relativeDir.split('/').length
            return Math.max(0, subpathDepth - relativeDirDepth)
        }
        return subpathDepth
    }

    public render(): JSX.Element | null {
        const { currSubpath, store } = this.props
        const nodes = currSubpath === '' ? store.nodes : store.nodeMap.get(currSubpath)!.children
        return (
            <table className="tree-layer">
                <tbody>
                    <tr>
                        <td>
                            {nodes.map((node, i) => (
                                <TreeRow key={i} {...this.props} depth={this.getDepth()} node={node} />
                            ))}
                        </td>
                    </tr>
                </tbody>
            </table>
        )
    }
}

interface TreeRowProps extends TreeLayerProps {
    depth: number
    node: TreeNode
}

class TreeRow extends React.PureComponent<TreeRowProps, TreeNodeState> {
    private subscriptions = new Subscription()

    constructor(props: TreeRowProps) {
        super(props)
        this.state = props.node.state.getValue()
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.node.state.subscribe(state => {
                this.setState(state)
            })
        )
    }

    public componentWillReceiveProps(nextProps: TreeRowProps): void {
        if (this.props.node !== nextProps.node) {
            if (this.subscriptions) {
                this.subscriptions.unsubscribe()
                this.subscriptions = new Subscription()
            }
            this.subscriptions.add(
                nextProps.node.state.subscribe(state => {
                    this.setState(state)
                })
            )
        }
    }

    public componentWillUnmount(): void {
        if (this.subscriptions) {
            this.subscriptions.unsubscribe()
        }
    }

    public render(): JSX.Element | null {
        const { node, store } = this.props
        return (
            <table className="tree-row">
                <tbody>
                    {node.children.length > 0 && [
                        <tr
                            key={node.filePath}
                            className={[
                                'tree__row',
                                this.showSubpath(node.filePath) && 'tree__row--expanded',
                                node.filePath === store.selectedPath && 'tree__row--selected',
                            ]
                                .filter(c => !!c)
                                .join(' ')}
                        >
                            <td onClick={this.handleDirClick}>
                                <div
                                    className="tree__row-contents"
                                    data-tree-directory="true"
                                    data-tree-path={node.filePath}
                                    // tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling)
                                    style={treePadding(this.props.depth, true)}
                                >
                                    <a
                                        data-pjax="true"
                                        className="tree__row-icon"
                                        onClick={this.noopRowClick}
                                        href={toTreeURL({
                                            repoPath: this.props.repoPath,
                                            rev: this.props.rev,
                                            filePath: node.filePath,
                                        })}
                                    >
                                        {this.showSubpath(node.filePath) ? (
                                            <span className="icon-inline">
                                                <ChevronDownIcon size={17} />
                                            </span>
                                        ) : (
                                            <span className="icon-inline">
                                                <ChevronRightIcon size={17} />
                                            </span>
                                        )}
                                    </a>
                                    <a className="tree__row-label" draggable={false} title={node.filePath}>
                                        {node.filePath.split('/').pop()}
                                    </a>
                                </div>
                            </td>
                        </tr>,
                        this.showSubpath(node.filePath) && (
                            <tr key={'layer-' + node.filePath}>
                                <td>
                                    <TreeLayer
                                        repoPath={this.props.repoPath}
                                        rev={this.props.rev}
                                        store={this.props.store}
                                        currSubpath={node.filePath}
                                        relativeDir={this.props.relativeDir}
                                    />
                                </td>
                            </tr>
                        ),
                    ]}
                    {node.children.length === 0 && (
                        <tr
                            key={node.filePath}
                            className={'tree__row' + (node.filePath === store.selectedPath ? '--selected' : '')}
                        >
                            {/* tslint:disable-next-line:jsx-ban-props (needed because of dynamic styling) */}
                            <td className="tree__row-label" style={treePadding(this.props.depth, false)}>
                                <a
                                    href={`https://${this.props.repoPath}/blob/${this.props.rev}/${node.filePath}`}
                                    data-pjax="true"
                                    className="tree__row-contents"
                                    onClick={this.linkRowClick}
                                    data-tree-path={node.filePath}
                                    draggable={false}
                                    title={node.filePath}
                                >
                                    {node.filePath.split('/').pop()}
                                </a>
                            </td>
                        </tr>
                    )}
                </tbody>
            </table>
        )
    }

    private handleDirClick = () => {
        const state = this.props.node.state.getValue()
        selectRow(this.props.store, this.props.node.filePath)
        if (!state.collapsed) {
            closeDirectory(this.props.store, this.props.node.filePath)
        } else {
            this.props.node.state.next({ collapsed: false, selected: true })
        }
    }

    private showSubpath(dir: string): boolean {
        const node = this.props.store.nodeMap.get(dir)
        if (!node) {
            return false
        }
        return !node.state.getValue().collapsed
    }

    /**
     * noopRowClick is the click handler for <a> rows of the tree element
     * that shouldn't update URL on click w/o modifier key (but should retain
     * anchor element properties, like right click "Copy link address").
     */
    private noopRowClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
        if (!e.altKey && !e.metaKey && !e.shiftKey && !e.ctrlKey) {
            e.preventDefault()
        }
        selectRow(this.props.store, this.props.node.filePath)
    }

    /**
     * linkRowClick is the click handler for <Link>
     */
    private linkRowClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
        selectRow(this.props.store, this.props.node.filePath)
    }
}
