import * as React from 'react'
import { Link } from 'react-router-dom'

/**
 *  Returns the friendly display form of the repository path (e.g., removing "github.com/").
 */
export function displayRepoPath(repoPath: string): string {
    let parts = repoPath.split('/')
    if (parts.length >= 3 && parts[0].includes('.')) {
        parts = parts.slice(1) // remove hostname from repo path (reduce visual noise)
    }
    return parts.join('/')
}

/**
 * Splits the repository path into the dir and base components.
 */
export function splitPath(path: string): [string, string] {
    const components = path.split('/')
    return [components.slice(0, -1).join('/'), components[components.length - 1]]
}

interface Props {
    repoPath: string
    repoURL: string
    filePath: string
    fileURL: string
}

/**
 * A link to a repository or a file within a repository, formatted as "repo" or "repo > file". Unless you
 * absolutely need breadcrumb-like behavior, use this instead of FilePathBreadcrumb.
 */
export const RepoFileLink: React.SFC<Props> = ({ repoPath, repoURL, filePath, fileURL }) => {
    const [fileBase, fileName] = splitPath(filePath)
    return (
        <>
            <Link to={repoURL}>{displayRepoPath(repoPath)}</Link> ›{' '}
            <Link to={fileURL}>
                {fileBase ? `${fileBase}/` : null}
                <strong>{fileName}</strong>
            </Link>
        </>
    )
}
