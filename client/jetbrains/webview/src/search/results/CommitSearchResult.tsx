import React from 'react'

import { mdiSourceCommit } from '@mdi/js'
import { formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/search-ui'
import { CommitMatch } from '@sourcegraph/shared/src/search/stream'
// eslint-disable-next-line no-restricted-imports
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Code } from '@sourcegraph/wildcard'

import { InfoDivider } from './InfoDivider'
import { RepoName } from './RepoName'
import { SearchResultLayout } from './SearchResultLayout'
import { SelectableSearchResult } from './SelectableSearchResult'

import styles from './CommitSearchResult.module.scss'

interface Props {
    match: CommitMatch
    selectedResult: null | string
    selectResult: (id: string) => void
}

export const CommitSearchResult: React.FunctionComponent<Props> = ({ match, selectedResult, selectResult }: Props) => {
    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)

    return (
        <SelectableSearchResult match={match} selectResult={selectResult} selectedResult={selectedResult}>
            {isActive => (
                <SearchResultLayout
                    isActive={isActive}
                    iconColumn={{
                        icon: mdiSourceCommit,
                        repoName: match.repository,
                    }}
                    infoColumn={
                        formattedRepositoryStarCount && (
                            <>
                                <Code className={styles.commitOid}>{match.oid.slice(0, 7)}</Code>{' '}
                                <Timestamp date={match.authorDate} noAbout={true} strict={true} />
                                <InfoDivider />
                                <SearchResultStar aria-label={`${match.repoStars} stars`} />
                                <span aria-hidden={true}>{formattedRepositoryStarCount}</span>
                            </>
                        )
                    }
                >
                    <RepoName
                        repoName={match.repository}
                        suffix={`${match.authorName}: ${match.message.split('\n', 1)[0]}`}
                    />
                </SearchResultLayout>
            )}
        </SelectableSearchResult>
    )
}
