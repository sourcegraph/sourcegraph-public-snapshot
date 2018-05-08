import ChevronDownIcon from '@sourcegraph/icons/lib/ChevronDown'
import ChevronRightIcon from '@sourcegraph/icons/lib/ChevronRight'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { toBlobURL, toTreeURL } from '../util/url'
import { TreeNode } from './Tree3'
import { TreeLayer, TreeLayerProps } from './TreeLayer'

const treePadding = (depth: number, directory: boolean) => ({
    paddingLeft: depth * 12 + (directory ? 0 : 12) + 12 + 'px',
    paddingRight: '16px',
})

export interface TreeRowProps extends TreeLayerProps {
    index: number
    parent: TreeNode | null
    depth: number
    node: GQL.IFile | GQL.IDirectory
    isExpanded: boolean
    onChangeViewState: (path: string, resolveTo: boolean, node: TreeNode) => void
    onSelect: (node: TreeNode) => void
    onSelectedNodeChange: (node: TreeNode) => void
    setChildNodes: (node: TreeNode, index: number) => void
}

export class TreeRow extends React.Component<TreeRowProps, {}> {
    public node: TreeNode

    constructor(props: TreeRowProps) {
        super(props)

        this.node = {
            index: this.props.index,
            parent: this.props.parent,
            childNodes: [],
            path: this.props.node.path,
        }
    }

    public componentDidMount(): void {
        // Sets the selectedNode as the activePath when navigating directly to a file.
        if (
            this.props.selectedNode &&
            this.props.activePath !== '' &&
            this.props.selectedNode.path === '' &&
            this.props.selectedNode.path !== this.props.activePath &&
            this.props.activePath === this.node.path
        ) {
            this.props.onSelectedNodeChange(this.node)
        }

        // Set this row as a childNode of its TreeLayer parent
        this.props.setChildNodes(this.node, this.node.index)
    }

    public shouldComponentUpdate(nextProps: TreeRowProps): boolean {
        if (nextProps.selectedNode !== this.props.selectedNode) {
            // Update if this row will be the selected node
            if (nextProps.selectedNode === this.node) {
                return true
            }

            // Update if a parent of the next selected row
            let parent = nextProps.selectedNode && nextProps.selectedNode.parent
            while (parent || parent !== null) {
                if (parent === this.node) {
                    return true
                }
                parent = parent && parent.parent
            }

            // Update if currently selected node
            if (this.props.selectedNode === this.node) {
                return true
            }

            // Update if parent of currently selected node
            let currentParent = this.props.selectedNode && this.props.selectedNode.parent
            while (currentParent || currentParent !== null) {
                if (currentParent === this.node) {
                    return true
                }
                currentParent = currentParent && currentParent.parent
            }

            // Update if currently activePath
            if (this.props.activePath === this.props.node.path) {
                return true
            }

            return false
        }
        return true
    }

    public render(): JSX.Element | null {
        const { node, selectedNode } = this.props
        const className = [
            'tree__row',
            this.node === selectedNode && 'tree__row--selected',
            this.props.isExpanded && 'tree__row--expanded',
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
                                            ref={ref => {
                                                if (ref) {
                                                    this.node.childNodes = ref.node.childNodes
                                                }
                                            }}
                                            selectedNode={this.props.selectedNode}
                                            parent={this.node}
                                            activePath={this.props.activePath}
                                            activePathIsDir={this.props.activePathIsDir}
                                            repoPath={this.props.repoPath}
                                            rev={this.props.rev}
                                            depth={this.props.depth + 1}
                                            history={this.props.history}
                                            parentPath={node.path}
                                            resolveTo={this.props.resolveTo}
                                            onSelect={this.props.onSelect}
                                            onChangeViewState={this.props.onChangeViewState}
                                            onSelectedNodeChange={this.props.onSelectedNodeChange}
                                            focusTreeOnUnmount={this.props.focusTreeOnUnmount}
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
        this.props.onSelect(this.node)
        this.props.onChangeViewState(this.props.node.path, !this.props.isExpanded, this.node)
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
        this.props.onChangeViewState(this.props.node.path, !this.props.isExpanded, this.node)
    }

    /**
     * linkRowClick is the click handler for <Link>
     */
    private linkRowClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
        this.props.onSelect(this.node)
    }
}
