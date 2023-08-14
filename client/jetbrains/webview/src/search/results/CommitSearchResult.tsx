import React from 'react'

import SourceCommitIcon from 'mdi-react/SourceCommitIcon'

import { formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/branded'
import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import type { CommitMatch } from '@sourcegraph/shared/src/search/stream'
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
    openResult: (id: string) => void
}

export const CommitSearchResult: React.FunctionComponent<Props> = ({
    match,
    selectedResult,
    selectResult,
    openResult,
}: Props) => {
    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)

    return (
        <SelectableSearchResult
            match={match}
            selectedResult={selectedResult}
            selectResult={selectResult}
            openResult={openResult}
        >
            {isActive => (
                <SearchResultLayout
                    isActive={isActive}
                    iconColumn={{
                        icon: SourceCommitIcon,
                        repoName: match.repository,
                    }}
                    infoColumn={
                        formattedRepositoryStarCount && (
                            <>
                                <Code className={styles.commitOid}>{match.oid.slice(0, 7)}</Code>{' '}
                                <Timestamp date={match.authorDate} noAbout={true} strict={true} />
                                <InfoDivider />
                                <SearchResultStar />
                                {formattedRepositoryStarCount}
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
