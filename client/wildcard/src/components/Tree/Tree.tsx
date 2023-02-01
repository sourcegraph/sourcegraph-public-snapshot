import { useCallback } from 'react'

import { mdiMenuRight, mdiMenuDown } from '@mdi/js'
import classNames from 'classnames'

import { Icon, LoadingSpinner } from '..'

import TreeView, { INode, ITreeViewProps } from './react-accessible-treeview'

import styles from './Tree.module.scss'

export type TreeNode = INode

interface Props<N extends TreeNode> extends Omit<ITreeViewProps, 'nodes' | 'onSelect' | 'onLoadData' | 'nodeRenderer'> {
    data: N[]

    onSelect?: (args: { element: N; isSelected: boolean }) => void
    onLoadData?: (args: { element: N }) => Promise<void>

    renderNode?: (args: {
        element: N
        isBranch: boolean
        isExpanded: boolean
        handleSelect: (event: React.MouseEvent) => {}
        handleExpand: (event: React.MouseEvent) => {}
        props: { className: string; tabIndex: number }
    }) => React.ReactNode

    // A set of node IDs that had their children loaded. This is necessary
    // because we can not rely on the .length property to know if we're still
    // loading children.
    loadedIds?: Set<number>
}
export function Tree<N extends TreeNode>(props: Props<N>): JSX.Element {
    const { onSelect, onLoadData, renderNode, loadedIds, ...rest } = props

    const _onSelect = useCallback(
        // TreeView expects nodes to be INode but ours are extending this type,
        // hence the any cast.
        (args: { element: any; isSelected: boolean }): void => {
            onSelect?.(args)
        },
        [onSelect]
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
                    data-tree-node-id={element.id}
                    className={classNames(styles.node, isSelected && styles.selected)}
                >
                    {isBranch ? (
                        // We already handle accessibility events for expansion in the <TreeView />
                        // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions
                        <div className={classNames(styles.icon, styles.collapseIcon)} onClick={onClick}>
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
        [loadedIds, renderNode]
    )

    return (
        <TreeView
            {...rest}
            className={classNames(styles.fileTree, rest.className)}
            // TreeView expects nodes to be INode but ours are extending this type.
            onSelect={_onSelect}
            onLoadData={_onLoadData}
            nodeRenderer={_renderNode}
        />
    )
}

function getMarginLeft(level: number, isBranch: boolean): string {
    // The level starts with 1 so the least margin by this logic is 0.75 * 1.
    //
    // Since folders render a chevron icon that is 1.25rem wide and we want to
    // render it to the left of the item, we need to add 0.5rem so we don't have
    // a negative margin
    level += 0.5

    if (isBranch) {
        return `${0.75 * level - 1.25}rem`
    }
    return `${0.75 * level}rem`
}
