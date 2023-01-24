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
    }) => React.ReactNode
}
export function Tree<N extends TreeNode>(props: Props<N>): JSX.Element {
    const { onSelect, onLoadData, renderNode, ...rest } = props

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
            getNodeProps: (props: { onClick: (event: Event) => {} }) => {}
            level: number
            handleSelect: (event: React.MouseEvent) => {}
            handleExpand: (event: Event) => {}
        }): React.ReactNode => (
            <div
                {...getNodeProps({ onClick: handleExpand })}
                // eslint-disable-next-line react/forbid-dom-props
                style={{
                    marginLeft: getMarginLeft(level),
                    minWidth: `calc(100% - 0.5rem - ${getMarginLeft(level)})`,
                }}
                data-tree-node-id={element.id}
                className={classNames(styles.node, isSelected && styles.selected, isBranch && styles.branch)}
            >
                {isBranch ? (
                    <div className={classNames('mr-1', styles.icon)}>
                        {isExpanded && element.children.length === 0 ? (
                            <LoadingSpinner />
                        ) : (
                            <Icon aria-hidden={true} svgPath={isExpanded ? mdiMenuDown : mdiMenuRight} />
                        )}
                    </div>
                ) : null}
                {renderNode ? renderNode({ element, isBranch, isExpanded, handleSelect }) : null}
            </div>
        ),
        [renderNode]
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

function getMarginLeft(level: number): string {
    return `${0.75 * (level - 1)}rem`
}
