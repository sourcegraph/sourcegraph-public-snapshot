import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './TreeLayerCell.module.scss'

type TreeLayerCellProps = HTMLAttributes<HTMLTableCellElement>

export const TreeLayerCell: React.FunctionComponent<React.PropsWithChildren<TreeLayerCellProps>> = ({
    className,
    children,
    ...rest
}) => (
    <td className={classNames(className, styles.cell)} {...rest}>
        {children}
    </td>
)
