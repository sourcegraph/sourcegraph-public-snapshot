import * as React from 'react'

import { LinkOrSpan } from './LinkOrSpan'

/**
 * Returns the friendly display form of the repository name (e.g., removing "github.com/").
 */
export function displayRepoName(repoName: string): string {
    let parts = repoName.split('/')
    if (parts.length >= 3 && parts[0].includes('.')) {
        parts = parts.slice(1) // remove hostname from repo name (reduce visual noise)
    }
    return parts.join('/')
}

/**
 * Splits the repository name into the dir and base components.
 */
export function splitPath(path: string): [string, string] {
    const components = path.split('/')
    return [components.slice(0, -1).join('/'), components[components.length - 1]]
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
