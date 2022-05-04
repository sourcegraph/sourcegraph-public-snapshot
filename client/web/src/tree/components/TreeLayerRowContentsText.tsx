import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './TreeLayerRowContentsText.module.scss'

type TreeLayerRowContentsTextProps = HTMLAttributes<HTMLDivElement>

export const TreeLayerRowContentsText: React.FunctionComponent<
    React.PropsWithChildren<TreeLayerRowContentsTextProps>
> = ({ className, children, ...rest }) => (
    <div className={classNames(styles.treeRowContentsText, className)} {...rest}>
        {children}
    </div>
)
