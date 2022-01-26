import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import { Link, LinkProps } from '@sourcegraph/wildcard'

import styles from './TreeLayerRowContents.module.scss'

type TreeLayerRowContentsProps = HTMLAttributes<HTMLDivElement> & {
    isNew?: boolean
}

export const TreeLayerRowContents: React.FunctionComponent<TreeLayerRowContentsProps> = ({
    className,
    children,
    isNew,
    ...rest
}) => (
    <div className={classNames(styles.treeRowContents, className, isNew && styles.isNew)} {...rest}>
        {children}
    </div>
)

type TreeLayerRowContentsLinkProps = LinkProps & {
    isNew?: boolean
}

export const TreeLayerRowContentsLink: React.FunctionComponent<TreeLayerRowContentsLinkProps> = ({
    className,
    children,
    isNew,
    ...rest
}) => (
    <Link className={classNames(styles.treeRowContents, className, isNew && styles.isNew)} {...rest}>
        {children}
    </Link>
)
