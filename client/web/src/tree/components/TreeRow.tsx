import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './TreeRow.module.scss'

type TreeRowProps = HTMLAttributes<HTMLTableRowElement> & {
    isActive?: boolean
    isSelected?: boolean
    isExpanded?: boolean
}

export const TreeRow: React.FunctionComponent<React.PropsWithChildren<TreeRowProps>> = ({
    isActive,
    isSelected,
    isExpanded,
    className,
    children,
    ...rest
}) => (
    <tr
        className={classNames(styles.row, isActive && styles.rowActive, isSelected && styles.rowSelected, className)}
        data-testid="tree-row"
        data-tree-row-active={isActive}
        data-tree-row-selected={isSelected}
        data-tree-row-expanded={isExpanded}
        {...rest}
    >
        {children}
    </tr>
)
