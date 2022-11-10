import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from '../Tree.module.scss'

type TreeLayerTableProps = HTMLAttributes<HTMLUListElement>

export const TreeLayerGroup: React.FunctionComponent<React.PropsWithChildren<TreeLayerTableProps>> = ({
    className,
    children,
    ...rest
}) => (
    <ul role="group" className={classNames(styles.treeLayer, className)} {...rest}>
        {children}
    </ul>
)
