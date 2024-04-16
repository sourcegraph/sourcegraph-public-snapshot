import * as React from 'react'

import { LinkOrSpan } from '@sourcegraph/wildcard'

/**
 * Returns the friendly display form of the repository name (e.g., removing "github.com/").
 */
export function displayRepoName(repoName: string): string {
    let parts = repoName.split('/')
    if (parts.length > 1 && parts[0].includes('.')) {
        parts = parts.slice(1) // remove hostname from repo name (reduce visual noise)
    }
    return parts.join('/')
}

/**
 * Returns the number of characters in the code host portion of the repository name
 * (e.g. "github.com/sourcegraph/sourcegraph") returns 11, the length of "github.com/"
 */
export function codeHostSubstrLength(repoName: string): number {
    const parts = repoName.split('/')
    if (parts.length >= 3 && parts[0].includes('.')) {
        return parts[0].length + 1 // add 1 to account for the trailing '/' in the code host name
    }
    return 0
}

/**
 * Splits the repository name into the dir and base components.
 */
export function splitPath(path: string): [string, string] {
    const components = path.split('/')
    return [components.slice(0, -1).join('/'), components.at(-1)!]
}

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
        <span className={className}>
            {repoBase ? `${repoBase}/` : null}
            <span className={repoClassName}>{repoName}</span>
        </span>
    )
    return (
        <LinkOrSpan className={className} to={to} onClick={onClick}>
            {children}
        </LinkOrSpan>
    )
}
