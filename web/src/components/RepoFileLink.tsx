import * as React from 'react'
import { Link } from 'react-router-dom'
import { RepoLink } from '../repo/RepoLink'
import { toPrettyBlobURL, toPrettyRepoURL, toTreeURL } from '../util/url'

export interface RepoBreadcrumbProps {
    repoPath: string
    rev?: string
    filePath?: string
    disableLinks?: boolean
    isDirectory?: boolean
}

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

const NotALink: React.SFC<{ to: any; children: React.ReactElement<any> }> = ({ children }) => children || null

/**
 * A link to a repository or a file within a repository, formatted as "repo" or "repo > file". Unless you
 * absolutely need breadcrumb-like behavior, use this instead of FilePathBreadcrumb.
 */
export const RepoFileLink: React.SFC<RepoBreadcrumbProps> = ({
    repoPath,
    rev,
    filePath,
    disableLinks,
    isDirectory,
}) => {
    const L = disableLinks ? NotALink : Link

    if (filePath) {
        const [fileBase, fileName] = splitPath(filePath)
        return (
            <>
                <L to={toPrettyRepoURL({ repoPath, rev })}>{displayRepoPath(repoPath)}</L> ›{' '}
                <L
                    to={
                        isDirectory
                            ? toTreeURL({ repoPath, rev, filePath })
                            : toPrettyBlobURL({ repoPath, rev, filePath })
                    }
                >
                    {fileBase ? `${fileBase}/` : null}
                    <strong>{fileName}</strong>
                </L>
            </>
        )
    }

    return <RepoLink repoPath={repoPath} rev={rev} to={disableLinks ? null : undefined} />
}
