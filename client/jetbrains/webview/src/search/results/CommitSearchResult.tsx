import React from 'react'

import classNames from 'classnames'

import { CodeHostIcon } from '@sourcegraph/shared/src/components/CodeHostIcon'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { SearchResultStar } from '@sourcegraph/shared/src/components/SearchResultStar'
import { CommitMatch, getCommitMatchUrl, getRepositoryUrl } from '@sourcegraph/shared/src/search/stream'
import { formatRepositoryStarCount } from '@sourcegraph/shared/src/util/stars'
// eslint-disable-next-line no-restricted-imports
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Link, Typography, useIsTruncated } from '@sourcegraph/wildcard'

import { getResultIdForCommitMatch } from './utils'

import styles from './SearchResult.module.scss'

interface Props {
    selectResult: (id: string) => void
    selectedResult: null | string
    match: CommitMatch
}

export const CommitSearchResult: React.FunctionComponent<Props> = ({ match, selectedResult, selectResult }: Props) => {
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)

    const resultId = getResultIdForCommitMatch(match)
    const onClick = (): void => selectResult(resultId)

    return (
        // The below element's accessibility is handled via a document level event listener.
        //
        // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions
        <div
            id={`search-result-list-item-${resultId}`}
            className={classNames(styles.line, {
                [styles.lineActive]: resultId === selectedResult,
            })}
            onMouseDown={preventAll}
            onClick={onClick}
            key={resultId}
        >
            <CodeHostIcon repoName={match.repository} className="text-muted flex-shrink-0" />
            <span
                onMouseEnter={checkTruncation}
                ref={titleReference}
                data-tooltip={(truncated && `${match.authorName}: ${match.message.split('\n', 1)[0]}`) || null}
            >
                <>
                    <Link to={getRepositoryUrl(match.repository)}>{displayRepoName(match.repository)}</Link>
                    {' › '}
                    <Link to={getCommitMatchUrl(match)}>{match.authorName}</Link>
                    {': '}
                    <Link to={getCommitMatchUrl(match)}>{match.message.split('\n', 1)[0]}</Link>
                </>
            </span>
            <span className={styles.spacer} />
            <Link to={getCommitMatchUrl(match)}>
                <Typography.Code className={styles.commitOid}>{match.oid.slice(0, 7)}</Typography.Code>{' '}
                <Timestamp date={match.authorDate} noAbout={true} strict={true} />
            </Link>
            {formattedRepositoryStarCount && (
                <>
                    <div className={styles.divider} />
                    <SearchResultStar />
                    {formattedRepositoryStarCount}
                </>
            )}
        </div>
    )
}

function preventAll(event: React.MouseEvent): void {
    event.stopPropagation()
    event.preventDefault()
}
