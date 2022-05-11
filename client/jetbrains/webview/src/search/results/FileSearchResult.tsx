import React from 'react'

import classNames from 'classnames'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'

import { RepoFileLink } from '@sourcegraph/shared/src/components/RepoFileLink'
import { RepoIcon } from '@sourcegraph/shared/src/components/RepoIcon'
import { SearchResultStar } from '@sourcegraph/shared/src/components/SearchResultStar'
import { ContentMatch, getFileMatchUrl } from '@sourcegraph/shared/src/search/stream'
import { formatRepositoryStarCount } from '@sourcegraph/shared/src/util/stars'
import { Icon } from '@sourcegraph/wildcard'

import { TrimmedCodeLineWithHighlights } from './TrimmedCodeLineWithHighlights'
import { getIdForLine } from './utils'

interface Props {
    selectResultFromId: (id: string) => void
    selectedResult: null | string
    result: ContentMatch
}

import styles from './FileSearchResult.module.scss'

export const FileSearchResult: React.FunctionComponent<Props> = ({
    result,
    selectedResult,
    selectResultFromId,
}: Props) => {
    const lines = result.lineMatches.map(line => {
        const key = getIdForLine(result, line)
        const onClick = (): void => selectResultFromId(key)

        return (
            // The below element's accessibility is handled via a document level event listener.
            //
            // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions
            <div
                id={`search-result-list-item-${key}`}
                className={classNames(styles.line, {
                    [styles.lineActive]: key === selectedResult,
                })}
                onMouseDown={preventAll}
                onClick={onClick}
                key={key}
            >
                <div className={styles.lineCode}>
                    <TrimmedCodeLineWithHighlights line={line} />
                </div>
                <div className={classNames(styles.lineLineNumber, 'text-muted')}>{line.lineNumber}</div>
            </div>
        )
    })

    const repoDisplayName = result.repository
    const repoAtRevisionURL = '#'
    const formattedRepositoryStarCount = formatRepositoryStarCount(result.repoStars)

    const title = (
        // eslint-disable-next-line jsx-a11y/no-static-element-interactions
        <div className={styles.header} onMouseDown={preventAll}>
            <div className={classNames(styles.headerTitle)} data-testid="result-container-header">
                <Icon role="img" title="File" className="flex-shrink-0" as={FileDocumentIcon} />
                <div className={classNames('mx-1', styles.headerDivider)} />
                <RepoIcon repoName={result.repository} className="text-muted flex-shrink-0" />
                <RepoFileLink
                    repoName={result.repository}
                    repoURL={repoAtRevisionURL}
                    filePath={result.path}
                    fileURL={getFileMatchUrl(result)}
                    repoDisplayName={repoDisplayName}
                    className={classNames('ml-1', 'flex-shrink-past-contents', 'text-truncate', styles.headerLink)}
                />
            </div>
            {formattedRepositoryStarCount && (
                <>
                    <SearchResultStar />
                    {formattedRepositoryStarCount}
                </>
            )}
        </div>
    )

    return (
        <>
            {title}
            {lines}
        </>
    )
}

function preventAll(event: React.MouseEvent): void {
    event.stopPropagation()
    event.preventDefault()
}
