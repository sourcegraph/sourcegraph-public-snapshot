import React from 'react'

import classNames from 'classnames'

import { Link, type LinkProps } from '@sourcegraph/wildcard'

import { GitReferenceNode, type GitReferenceNodeProps } from '../../../GitReference'

import styles from './ConnectionPopoverNodeLink.module.scss'

type ConnectionPopoverNodeLinkProps = LinkProps & {
    active: boolean
}

export const ConnectionPopoverNodeLink: React.FunctionComponent<
    React.PropsWithChildren<ConnectionPopoverNodeLinkProps>
> = ({ className, children, active, ...rest }) => (
    <Link
        className={classNames(
            styles.connectionPopoverNode,
            styles.connectionPopoverNodeLink,
            active && styles.connectionPopoverNodeActive,
            className
        )}
        {...rest}
    >
        {children}
    </Link>
)

type ConnectionPopoverGitReferenceNodeProps = GitReferenceNodeProps & {
    active: boolean
}

export const ConnectionPopoverGitReferenceNode: React.FunctionComponent<
    React.PropsWithChildren<ConnectionPopoverGitReferenceNodeProps>
> = ({ className, active, ...rest }) => (
    <GitReferenceNode
        className={classNames(styles.connectionPopoverNode, active && styles.connectionPopoverNodeActive, className)}
        nodeLinkClassName={styles.connectionPopoverNodeLink}
        {...rest}
    />
)
