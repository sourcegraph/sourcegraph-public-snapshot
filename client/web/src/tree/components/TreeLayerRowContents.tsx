import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import { Link, LinkProps } from '@sourcegraph/wildcard'

import styles from './TreeLayerRowContents.module.scss'

type TreeLayerRowContentsProps = HTMLAttributes<HTMLDivElement> & {
    isNew?: boolean
}

export const TreeLayerRowContents: React.FunctionComponent<React.PropsWithChildren<TreeLayerRowContentsProps>> = ({
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

export const TreeLayerRowContentsLink: React.FunctionComponent<
    React.PropsWithChildren<TreeLayerRowContentsLinkProps>
> = ({ className, children, isNew, ...rest }) => (
    <Link className={classNames(styles.treeRowContents, className, isNew && styles.isNew)} {...rest}>
        {children}
    </Link>
)
