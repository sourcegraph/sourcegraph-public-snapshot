import * as React from 'react'
import { useEffect, useRef } from 'react'

import classNames from 'classnames'

import { appendSubtreeQueryParameter, highlightNode } from '@sourcegraph/common'
import { displayRepoName, splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import { Range } from '@sourcegraph/shared/src/search/stream'
import { Link } from '@sourcegraph/wildcard'

interface Props {
    repoName: string
    repoURL: string
    filePath: string
    pathMatchRanges?: Range[]
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
    pathMatchRanges,
    fileURL,
    className,
}) => {
    const [fileBase, fileName] = splitPath(filePath)
    const containerElement = useRef<HTMLAnchorElement>(null)

    const repoFileLink = (): JSX.Element => (
        <div className={classNames(className)}>
            <Link to={repoURL}>{repoDisplayName || displayRepoName(repoName)}</Link>
            <span aria-hidden={true}> ›</span>{' '}
            <Link to={appendSubtreeQueryParameter(fileURL)} ref={containerElement}>
                {fileBase ? `${fileBase}/` : null}
                <strong>{fileName}</strong>
            </Link>
        </div>
    )

    useEffect((): void => {
        if (containerElement.current && pathMatchRanges && fileBase && fileName) {
            for (const range of pathMatchRanges) {
                highlightNode(
                    containerElement.current as HTMLElement,
                    range.start.column,
                    range.end.column - range.start.column
                )
            }
        }
    }, [pathMatchRanges, fileBase, fileName, containerElement])

    return repoFileLink()
}
