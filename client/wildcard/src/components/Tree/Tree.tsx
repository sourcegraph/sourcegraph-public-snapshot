import { useCallback, useEffect } from 'react'

import { mdiMenuRight, mdiMenuDown } from '@mdi/js'
import classNames from 'classnames'
import TreeView, { INode, ITreeViewProps } from 'react-accessible-treeview'

import { Icon, LoadingSpinner } from '..'

import styles from './Tree.module.scss'

interface Props<N extends INode> extends Omit<ITreeViewProps, 'nodes' | 'onSelect' | 'onLoadData' | 'nodeRenderer'> {
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
export function Tree<N extends INode>(props: Props<N>): JSX.Element {
    usePatchFocusToFixScrollIssues()

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
        async (args: { element: any }): Promise<void> => {
            return onLoadData?.(args)
        },
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
            className={`${styles.fileTree} ${rest.className ?? ''}`}
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

// By default, react-accessible-tree uses .focus() on the `<li>` to scroll to
// the element. In case of a larger nested list, this is causing unexpected
// jumps in the UI because the browser wants to fit as much content from the
// `<li>` into the viewport as possible.
//
// This is not what we want though, because we only render the focus outline
// on the folder name as well and thus the jumping is unexpected. To fix this,
// we have to patch `.focus()` and handle the case ourselves.
//
// TODO: This can be removed once the upstream fix is landed:
//       https://github.com/dgreene1/react-accessible-treeview/pull/81
function usePatchFocusToFixScrollIssues(): void {
    useEffect(() => {
        // eslint-disable-next-line @typescript-eslint/unbound-method
        const originalFocus = HTMLElement.prototype.focus
        HTMLElement.prototype.focus = function (...args) {
            const isBranch = this.nodeName === 'LI' && this.classList.contains('tree-branch-wrapper')
            const isNode = this.nodeName === 'DIV' && this.classList.contains(styles.node)
            if (isBranch || isNode) {
                const focusableNode = isNode ? this : this.querySelector(`.${styles.node}`)
                if (focusableNode) {
                    focusableNode.scrollIntoView({ block: 'nearest' })
                    return originalFocus.call(this, { preventScroll: true })
                }
            }
            return originalFocus.call(this, ...args)
        }
        return () => {
            HTMLElement.prototype.focus = originalFocus
        }
    }, [])
}
