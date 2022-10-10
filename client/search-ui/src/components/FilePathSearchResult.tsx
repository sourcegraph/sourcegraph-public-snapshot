import React from 'react'

import classNames from 'classnames'

import { getFileMatchUrl, getRepositoryUrl, getRevision, PathMatch } from '@sourcegraph/shared/src/search/stream'

import { LastSyncedIcon } from './LastSyncedIcon'
import { RepoFileLink } from './RepoFileLink'
import { ResultContainer } from './ResultContainer'

import styles from './SearchResult.module.scss'

export interface FilePathSearchResult {
    result: PathMatch
    repoDisplayName: string
    onSelect: () => void
    containerClassName?: string
    index: number
}

export const FilePathSearchResult: React.FunctionComponent<FilePathSearchResult> = ({
    result,
    repoDisplayName,
    onSelect,
    containerClassName,
    index,
}) => {
    const repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    const revisionDisplayName = getRevision(result.branches, result.commit)

    const title = (
        <RepoFileLink
            repoName={result.repository}
            repoURL={repoAtRevisionURL}
            filePath={result.path}
            pathMatchRanges={result.pathMatches ?? []}
            fileURL={getFileMatchUrl(result)}
            repoDisplayName={
                repoDisplayName
                    ? `${repoDisplayName}${revisionDisplayName ? `@${revisionDisplayName}` : ''}`
                    : undefined
            }
            className={classNames(styles.titleInner, styles.mutedRepoFileLink)}
        />
    )

    return (
        <ResultContainer
            index={index}
            title={title}
            resultType={result.type}
            onResultClicked={onSelect}
            repoName={result.repository}
            repoStars={result.repoStars}
            className={containerClassName}
        >
            <div className={classNames(styles.searchResultMatch, 'p-2')}>
                {result.repoLastFetched && <LastSyncedIcon lastSyncedTime={result.repoLastFetched} />}
                <small>{result.pathMatches ? 'Path match' : 'File contains matching content'}</small>
            </div>
        </ResultContainer>
    )
}
