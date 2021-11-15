import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './TreeLayerTable.module.scss'

type TreeLayerTableProps = HTMLAttributes<HTMLTableElement>

export const TreeLayerTable: React.FunctionComponent<TreeLayerTableProps> = ({ className, children, ...rest }) => (
    <table className={classNames(styles.treeLayer, className)} {...rest}>
        {children}
    </table>
)
