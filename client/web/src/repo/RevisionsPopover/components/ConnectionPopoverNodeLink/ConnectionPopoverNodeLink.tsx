import classNames from 'classnames'
import React from 'react'

import { Link, LinkProps } from '@sourcegraph/wildcard'

import { GitReferenceNode, GitReferenceNodeProps } from '../../../GitReference'

import styles from './ConnectionPopoverNodeLink.module.scss'

type ConnectionPopoverNodeLinkProps = LinkProps & {
    active: boolean
}

export const ConnectionPopoverNodeLink: React.FunctionComponent<ConnectionPopoverNodeLinkProps> = ({
    className,
    children,
    active,
    ...rest
}) => (
    <Link
        className={classNames(
            styles.connectionPopoverNodeLink,
            active && styles.connectionPopoverNodeLinkActive,
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

export const ConnectionPopoverGitReferenceNode: React.FunctionComponent<ConnectionPopoverGitReferenceNodeProps> = ({
    className,
    active,
    ...rest
}) => (
    <GitReferenceNode
        className={classNames(
            styles.connectionPopoverNodeLink,
            active && styles.connectionPopoverNodeLinkActive,
            className
        )}
        {...rest}
    />
)
