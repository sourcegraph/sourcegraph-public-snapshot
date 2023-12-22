import { forwardRef, useCallback, useImperativeHandle, useRef, useMemo } from 'react'

import { mdiMenuRight, mdiMenuDown } from '@mdi/js'
import classNames from 'classnames'

import { Icon, LoadingSpinner } from '..'

import TreeView, { type INode, type ITreeViewProps } from './react-accessible-treeview'

import styles from './Tree.module.scss'

export type TreeNode = INode

export interface TreeProps<N extends TreeNode>
    extends Omit<ITreeViewProps, 'nodes' | 'onSelect' | 'onExpand' | 'onLoadData' | 'nodeRenderer'> {
    data: N[]

    onSelect?: (args: { element: N; isSelected: boolean }) => void
    onExpand?: (args: { element: N; isExpanded: boolean }) => void
    onLoadData?: (args: { element: N }) => Promise<void>

    renderNode?: (args: {
        element: N
        isBranch: boolean
        isExpanded: boolean
        handleSelect: (event: React.MouseEvent) => {}
        handleExpand: (event: React.MouseEvent) => {}
        props: { className: string; tabIndex: number }
    }) => React.ReactNode

    /**
     * A set of node IDs that had their children loaded. This is necessary
     * because we can not rely on the .length property to know if we're still
     * loading children.
     */
    loadedIds?: Set<number>

    /** Optional className that is passed through to the focused nodes. */
    nodeClassName?: string
}

export interface TreeRef {
    focus: () => void
}

function TreeComponent<N extends TreeNode>(props: TreeProps<N>, ref?: React.Ref<TreeRef>): JSX.Element {
    const { onSelect, onExpand, onLoadData, renderNode, loadedIds, nodeClassName, expandedIds, ...rest } = props

    const treeViewRef = useRef<HTMLUListElement>(null)
    useImperativeHandle(
        ref,
        () => ({
            focus: () => {
                if (treeViewRef.current?.contains(document.activeElement)) {
                    return
                }
                const element = treeViewRef.current?.querySelector("[tabindex]:not([tabindex='-1'])")
                if (element instanceof HTMLElement) {
                    element.focus()
                }
            },
        }),
        []
    )

    const _onSelect = useCallback(
        // TreeView expects nodes to be INode but ours are extending this type,
        // hence the any cast.
        (args: { element: any; isSelected: boolean }): void => {
            onSelect?.(args)
        },
        [onSelect]
    )
    const _onExpand = useCallback(
        // TreeView expects nodes to be INode but ours are extending this type,
        // hence the any cast.
        (args: { element: any; isExpanded: boolean }): void => {
            onExpand?.(args)
        },
        [onExpand]
    )

    const _onLoadData = useCallback(
        // TreeView expects nodes to be INode but ours are extending this type,
        // hence the any cast.
        async (args: { element: any }): Promise<void> => onLoadData?.(args),
        [onLoadData]
    )

    const _renderNode: any = useCallback(
        ({
            element,
            isBranch,
            isExpanded,
            isSelected,
            getNodeProps,
            level,
            handleSelect,
            handleExpand,
        }: {
            // TreeView expects nodes to be INode but ours are extending this type,
            // hence the any cast.
            element: any
            isBranch: boolean
            isExpanded: boolean
            isSelected: boolean
            getNodeProps: (props: { onClick: (event: React.MouseEvent) => {} }) => {
                onClick: (event: React.MouseEvent) => {}
            }
            level: number
            handleSelect: (event: React.MouseEvent) => {}
            handleExpand: (event: React.MouseEvent) => {}
        }): React.ReactNode => {
            const { onClick, ...props } = getNodeProps({ onClick: handleExpand })
            return (
                <div
                    {...props}
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{
                        marginLeft: getMarginLeft(level, isBranch),
                        minWidth: `calc(100% - 0.5rem - ${getMarginLeft(level, isBranch)})`,
                    }}
                    data-testid="tree-node"
                    data-tree-node-id={element.id}
                    data-tree-active={isSelected}
                    data-tree-selected={isSelected}
                    data-tree-expanded={isExpanded}
                    className={classNames(styles.node, isSelected && styles.selected, nodeClassName)}
                >
                    {isBranch ? (
                        // We already handle accessibility events for expansion in the <TreeView />
                        // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions
                        <div
                            className={classNames(styles.icon, styles.collapseIcon)}
                            data-testid="tree-expand-icon"
                            onClick={onClick}
                        >
                            {isExpanded &&
                            element.children.length === 0 &&
                            (loadedIds ? !loadedIds.has(element.id) : true) ? (
                                <LoadingSpinner />
                            ) : (
                                <Icon aria-hidden={true} svgPath={isExpanded ? mdiMenuDown : mdiMenuRight} />
                            )}
                        </div>
                    ) : null}
                    {renderNode
                        ? renderNode({
                              element,
                              isBranch,
                              isExpanded,
                              handleSelect,
                              handleExpand,
                              props: {
                                  className: classNames(styles.content, { [styles.contentInBranch]: isBranch }),
                                  // We don't want links or any other item inside the Tree to be focusable, as focus
                                  // should be managed by the TreeView only.
                                  tabIndex: -1,
                              },
                          })
                        : null}
                </div>
            )
        },
        [loadedIds, nodeClassName, renderNode]
    )

    // <TreeView /> quirk:
    //
    // The root node (id = 0) is not a valid target to be expanded. If it is set accidentally, it
    // can leave the tree in an invalid state where no tabIndex={0} item is rendered (because the
    // tabIndex is assumed it has to be on the root node).
    const validExpandedIds = useMemo(
        () => (expandedIds ? expandedIds.filter(id => id !== 0) : expandedIds),
        [expandedIds]
    )

    return (
        <TreeView
            {...rest}
            ref={treeViewRef}
            expandedIds={validExpandedIds}
            className={classNames(styles.fileTree, rest.className)}
            // TreeView expects nodes to be INode but ours are extending this type.
            onSelect={_onSelect}
            onExpand={onExpand ? _onExpand : undefined}
            onLoadData={_onLoadData}
            nodeRenderer={_renderNode}
        />
    )
}

// Workaround to create a generic component as output of React.forwardRef: https://stackoverflow.com/a/58473012
export const Tree = forwardRef(TreeComponent) as <N extends TreeNode>(
    p: TreeProps<N> & { ref?: React.Ref<TreeRef> }
) => JSX.Element

function getMarginLeft(level: number, isBranch: boolean): string {
    if (isBranch) {
        return `${0.75 * level - 0.75}rem`
    }
    return `${0.75 * level}rem`
}

interface FlattenableTreeNode {
    name: string
    // TODO: My TS wizardry is not strong enough to make this work with a
    // generic type. 😅
    children?: any[]
}
type ReturnNode<T extends FlattenableTreeNode> = Omit<T, 'children'> & TreeNode
// This implementation is taken from `react-accessible-treeview` and modified to pass through all
// properties of the node instead of just the `name`.
export function flattenTree<N extends FlattenableTreeNode>(tree: N): ReturnNode<N>[] {
    let count = 0
    const flattenedTree: ReturnNode<N>[] = []

    const flattenTreeHelper = function (tree: N, parent: number | null): void {
        const { children, ...rest } = tree
        const node: ReturnNode<N> = {
            ...rest,
            id: count,
            name: tree.name,
            children: [],
            parent,
        }
        flattenedTree[count] = node
        count += 1
        if (tree.children === null || tree.children === undefined || tree.children.length === 0) {
            return
        }
        for (const child of tree.children) {
            flattenTreeHelper(child, node.id)
        }
        node.children = flattenedTree
            .filter((item: ReturnNode<N>) => item.parent === node.id)
            .map((item: ReturnNode<N>) => item.id)
    }

    flattenTreeHelper(tree, null)
    return flattenedTree
}
