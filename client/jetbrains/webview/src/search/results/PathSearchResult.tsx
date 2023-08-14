import React from 'react'

import FileDocumentIcon from 'mdi-react/FileDocumentIcon'

import { formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/branded'
import type { PathMatch } from '@sourcegraph/shared/src/search/stream'

import { RepoName } from './RepoName'
import { SearchResultLayout } from './SearchResultLayout'
import { SelectableSearchResult } from './SelectableSearchResult'

interface Props {
    match: PathMatch
    selectedResult: null | string
    selectResult: (id: string) => void
    openResult: (id: string) => void
}

export const PathSearchResult: React.FunctionComponent<Props> = ({
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
                        icon: FileDocumentIcon,
                        repoName: match.repository,
                    }}
                    infoColumn={
                        formattedRepositoryStarCount && (
                            <>
                                <SearchResultStar />
                                {formattedRepositoryStarCount}
                            </>
                        )
                    }
                >
                    <RepoName repoName={match.repository} suffix={match.path} />
                </SearchResultLayout>
            )}
        </SelectableSearchResult>
    )
}
