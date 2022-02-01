import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './TreeLayerRowContentsText.module.scss'

type TreeLayerRowContentsTextProps = HTMLAttributes<HTMLDivElement>

export const TreeLayerRowContentsText: React.FunctionComponent<TreeLayerRowContentsTextProps> = ({
    className,
    children,
    ...rest
}) => (
    <div className={classNames(styles.treeRowContentsText, className)} {...rest}>
        {children}
    </div>
)
