import React from 'react'

import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'

import { formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/branded'
import type { RepositoryMatch } from '@sourcegraph/shared/src/search/stream'

import { RepoName } from './RepoName'
import { SearchResultLayout } from './SearchResultLayout'
import { SelectableSearchResult } from './SelectableSearchResult'

export interface RepoSearchResultProps {
    match: RepositoryMatch
    selectedResult: null | string
    selectResult: (id: string) => void
    openResult: (id: string) => void
}

export const RepoSearchResult: React.FunctionComponent<RepoSearchResultProps> = ({
    match,
    selectedResult,
    selectResult,
    openResult,
}) => {
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
                    iconColumn={{ icon: SourceRepositoryIcon, repoName: match.repository }}
                    infoColumn={
                        formattedRepositoryStarCount && (
                            <>
                                <SearchResultStar />
                                {formattedRepositoryStarCount}
                            </>
                        )
                    }
                    isActive={isActive}
                >
                    <RepoName repoName={match.repository} suffix={match.description} />
                </SearchResultLayout>
            )}
        </SelectableSearchResult>
    )
}
