import * as React from 'react'

import classNames from 'classnames'

import { appendSubtreeQueryParameter } from '@sourcegraph/common'
import { displayRepoName, splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import { Link } from '@sourcegraph/wildcard'

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
export const RepoFileLink: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repoDisplayName,
    repoName,
    repoURL,
    filePath,
    fileURL,
    className,
}) => {
    const [fileBase, fileName] = splitPath(filePath)

    return (
        <div className={classNames(className)}>
            <Link to={repoURL}>{repoDisplayName || displayRepoName(repoName)}</Link>
            <span aria-hidden={true}> ›</span>{' '}
            <Link to={appendSubtreeQueryParameter(fileURL)}>
                {fileBase ? `${fileBase}/` : null}
                <strong>{fileName}</strong>
            </Link>
        </div>
    )
}
