import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './TreeRow.module.scss'

type TreeRowProps = HTMLAttributes<HTMLTableRowElement> & {
    isActive?: boolean
    isSelected?: boolean
}

export const TreeRow: React.FunctionComponent<TreeRowProps> = ({
    isActive,
    isSelected,
    className,
    children,
    ...rest
}) => (
    <tr
        className={classNames(styles.row, isActive && styles.rowActive, isSelected && styles.rowSelected, className)}
        {...rest}
    >
        {children}
    </tr>
)
