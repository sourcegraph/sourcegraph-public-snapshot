import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from '../Tree.module.scss'

type TreeLayerCellProps = HTMLAttributes<HTMLLIElement>

export const TreeLayerItem: React.FunctionComponent<React.PropsWithChildren<TreeLayerCellProps>> = ({
    className,
    children,
    ...rest
}) => (
    <li tabIndex={0} role="treeitem" className={classNames(className, styles.cell)} {...rest}>
        {children}
    </li>
)
