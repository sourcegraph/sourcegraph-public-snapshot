import * as React from 'react'

import { Link } from '@sourcegraph/wildcard'

import { displayRepoName, splitPath } from './RepoFileLink'

interface Props {
    repoName: string

    /**
     * The link's destination. If null, the element is text, not a link.
     */
    to: string | null

    className?: string

    repoClassName?: string

    onClick?: React.MouseEventHandler<HTMLElement>
}

export const RepoLink: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repoName: fullRepoName,
    to,
    className,
    onClick,
    repoClassName,
}) => {
    const [repoBase, repoName] = splitPath(displayRepoName(fullRepoName))
    const children = (
        <span className={className || ''}>
            {' '}
            {repoBase ? `${repoBase}/` : null}
            <span className={repoClassName}>{repoName}</span>
        </span>
    )
    if (to === null) {
        return children
    }
    return (
        <Link className={className || ''} to={to} onClick={onClick}>
            {children}
        </Link>
    )
}
