import * as React from 'react'

import { appendSubtreeQueryParameter } from '@sourcegraph/common'
import { Link } from '@sourcegraph/wildcard'

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
    repoURL: string
    filePath: string
    fileURL: string
    repoDisplayName?: string
    className?: string
}

/**
 * A link to a repository or a file within a repository, formatted as "repo" or "repo > file". Unless you
 * absolutely need breadcrumb-like behavior, use this instead of FilePathBreadcrumb.
 */
export const RepoFileLink: React.FunctionComponent<Props> = ({
    repoDisplayName,
    repoName,
    repoURL,
    filePath,
    fileURL,
    className,
}) => {
    const [fileBase, fileName] = splitPath(filePath)
    const [isTruncated, setIsTruncated] = React.useState<boolean>(false)
    const titleReference = React.useRef<HTMLAnchorElement>(null)

    function showTooltip(): void {
        if (titleReference.current) {
            setIsTruncated(titleReference.current.clientWidth < titleReference.current.scrollWidth)
        }
    }

    return (
        <div className={className}>
            <Link to={repoURL}>{repoDisplayName || displayRepoName(repoName)}</Link> ›{' '}
            <Link
                onMouseEnter={showTooltip}
                className="text-truncate"
                data-tooltip={isTruncated ? (fileBase ? `${fileBase}/${fileName}` : fileName) : null}
                to={appendSubtreeQueryParameter(fileURL)}
                ref={titleReference}
            >
                {fileBase ? `${fileBase}/` : null}
                <strong>{fileName}</strong>
            </Link>
        </div>
    )
}
