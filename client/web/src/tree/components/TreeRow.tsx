import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from '../Tree.module.scss'

const MAX_DIRECTORY_DEPTH = 100

type TreeRowProps = HTMLAttributes<HTMLTableRowElement> & {
    isActive?: boolean
    isSelected?: boolean
    isExpanded?: boolean
    depth?: number
}

export const TreeRow: React.FunctionComponent<React.PropsWithChildren<TreeRowProps>> = ({
    isActive,
    isSelected,
    isExpanded,
    className,
    children,
    depth = 0,
    ...rest
}) => (
    <tr
        className={classNames(
            styles.row,
            isActive && styles.rowActive,
            isSelected && styles.rowSelected,
            isExpanded && styles.rowExpanded,
            className
        )}
        data-testid="tree-row"
        data-tree-row-active={isActive}
        data-tree-row-selected={isSelected}
        data-tree-row-expanded={isExpanded}
        // eslint-disable-next-line react/forbid-dom-props
        style={{ top: `${depth * 2}em`, zIndex: MAX_DIRECTORY_DEPTH > depth ? MAX_DIRECTORY_DEPTH - depth : 'inherit' }}
        {...rest}
    >
        {children}
    </tr>
)
