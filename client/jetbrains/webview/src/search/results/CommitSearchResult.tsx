import React from 'react'

import { CodeHostIcon, formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/search-ui'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoLink'
import { CommitMatch } from '@sourcegraph/shared/src/search/stream'
// eslint-disable-next-line no-restricted-imports
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Code, Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import { SelectableSearchResult } from './SelectableSearchResult'

import styles from './CommitSearchResult.module.scss'

interface Props {
    match: CommitMatch
    selectedResult: null | string
    selectResult: (id: string) => void
}

export const CommitSearchResult: React.FunctionComponent<Props> = ({ match, selectedResult, selectResult }: Props) => {
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)

    return (
        <SelectableSearchResult match={match} selectResult={selectResult} selectedResult={selectedResult}>
            <CodeHostIcon repoName={match.repository} className="text-muted flex-shrink-0" />
            <Tooltip content={(truncated && `${match.authorName}: ${match.message.split('\n', 1)[0]}`) || null}>
                <span onMouseEnter={checkTruncation} ref={titleReference}>
                    {`${displayRepoName(match.repository)} › ${match.authorName}: ${match.message.split('\n', 1)[0]}`}
                </span>
            </Tooltip>
            <span className={styles.spacer} />
            <Code className={styles.commitOid}>{match.oid.slice(0, 7)}</Code>{' '}
            <Timestamp date={match.authorDate} noAbout={true} strict={true} />
            {formattedRepositoryStarCount && (
                <>
                    <div className={styles.divider} />
                    <SearchResultStar />
                    {formattedRepositoryStarCount}
                </>
            )}
        </SelectableSearchResult>
    )
}
