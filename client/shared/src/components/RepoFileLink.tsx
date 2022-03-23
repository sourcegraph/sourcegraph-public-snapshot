import * as React from 'react'

import classNames from 'classnames'

import { appendSubtreeQueryParameter } from '@sourcegraph/common'
import { useIsTruncated, Link } from '@sourcegraph/wildcard'

import styles from './RepoFileLink.module.scss'
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
    const { isTruncated, elementReference: titleReference, onMouseEnter } = useIsTruncated<HTMLAnchorElement>()

    return (
        <div className={classNames(className, styles.repoFileLink)}>
            <Link to={repoURL}>{repoDisplayName || displayRepoName(repoName)}</Link> ›{' '}
            <Link
                onMouseEnter={onMouseEnter}
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
