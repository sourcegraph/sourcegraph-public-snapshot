import * as React from 'react'

import classNames from 'classnames'

import { appendSubtreeQueryParameter } from '@sourcegraph/common'
import { displayRepoName, splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import { useIsTruncated, Link } from '@sourcegraph/wildcard'

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
    /**
     * Use the custom hook useIsTruncated to check if overflow: ellipsis is activated for the element
     * We want to do it on mouse enter as browser window size might change after the element has been
     * loaded initially
     */
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    return (
        <div
            ref={titleReference}
            onMouseEnter={checkTruncation}
            className={classNames(className)}
            data-tooltip={truncated ? (fileBase ? `${fileBase}/${fileName}` : fileName) : null}
        >
            <Link to={repoURL}>{repoDisplayName || displayRepoName(repoName)}</Link> ›{' '}
            <Link to={appendSubtreeQueryParameter(fileURL)}>
                {fileBase ? `${fileBase}/` : null}
                <strong>{fileName}</strong>
            </Link>
        </div>
    )
}
