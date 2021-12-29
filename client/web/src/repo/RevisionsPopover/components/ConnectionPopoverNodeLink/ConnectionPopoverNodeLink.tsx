import classNames from 'classnames'
import React from 'react'

import { RouterLink } from '@sourcegraph/wildcard'
import type { LinkProps } from '@sourcegraph/wildcard/src/components/Link'

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
    <RouterLink
        className={classNames(
            styles.connectionPopoverNodeLink,
            active && styles.connectionPopoverNodeLinkActive,
            className
        )}
        {...rest}
    >
        {children}
    </RouterLink>
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
