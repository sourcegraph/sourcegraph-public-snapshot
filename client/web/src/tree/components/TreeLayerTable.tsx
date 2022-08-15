import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from '../Tree.module.scss'

type TreeLayerTableProps = HTMLAttributes<HTMLTableElement>

export const TreeLayerTable: React.FunctionComponent<React.PropsWithChildren<TreeLayerTableProps>> = ({
    className,
    children,
    ...rest
}) => (
    <table className={classNames(styles.treeLayer, className)} {...rest}>
        {children}
    </table>
)
