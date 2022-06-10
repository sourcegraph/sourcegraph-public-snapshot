import React from 'react'

import AlphaSBoxIcon from 'mdi-react/AlphaSBoxIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'

import { formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/search-ui'
import { ContentMatch, SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'

import { InfoDivider } from './InfoDivider'
import { RepoName } from './RepoName'
import { SearchResultHeader } from './SearchResultHeader'
import { SearchResultLayout } from './SearchResultLayout'
import { SelectableSearchResult } from './SelectableSearchResult'
import { TrimmedCodeLineWithHighlights } from './TrimmedCodeLineWithHighlights'
import { getResultId } from './utils'

import styles from './FileSearchResult.module.scss'

interface Props {
    selectResult: (resultId: string) => void
    selectedResult: null | string
    match: ContentMatch | SymbolMatch
}

function getResultElementsForContentMatch(
    match: ContentMatch,
    selectResult: (resultId: string) => void,
    selectedResult: string | null
): JSX.Element[] {
    return match.lineMatches.map(line => (
        <SelectableSearchResult
            key={getResultId(match, line)}
            lineMatchOrSymbolName={line}
            match={match}
            selectedResult={selectedResult}
            selectResult={selectResult}
        >
            {isActive => (
                <SearchResultLayout infoColumn={line.lineNumber + 1} className={styles.code} isActive={isActive}>
                    <TrimmedCodeLineWithHighlights line={line} />
                </SearchResultLayout>
            )}
        </SelectableSearchResult>
    ))
}

function getResultElementsForSymbolMatch(
    match: SymbolMatch,
    selectResult: (resultId: string) => void,
    selectedResult: string | null
): JSX.Element[] {
    return match.symbols.map(symbol => (
        <SelectableSearchResult
            key={getResultId(match, symbol.name)}
            lineMatchOrSymbolName={symbol.name}
            match={match}
            selectedResult={selectedResult}
            selectResult={selectResult}
        >
            {isActive => (
                <SearchResultLayout className={styles.code} isActive={isActive}>
                    <SymbolIcon kind={symbol.kind} className="mr-1" />
                    {symbol.name} {symbol.containerName && <span className="text-muted">{symbol.containerName}</span>}
                </SearchResultLayout>
            )}
        </SelectableSearchResult>
    ))
}

export const FileSearchResult: React.FunctionComponent<Props> = ({ match, selectedResult, selectResult }: Props) => {
    const lines =
        match.type === 'content'
            ? getResultElementsForContentMatch(match, selectResult, selectedResult)
            : getResultElementsForSymbolMatch(match, selectResult, selectedResult)

    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)

    const title = (
        <SearchResultHeader>
            <SearchResultLayout
                iconColumn={{
                    icon: match.type === 'content' ? FileDocumentIcon : AlphaSBoxIcon,
                    repoName: match.repository,
                }}
                infoColumn={
                    formattedRepositoryStarCount && (
                        <>
                            {lines.length} {lines.length > 1 ? 'matches' : 'match'}
                            <InfoDivider />
                            <SearchResultStar />
                            {formattedRepositoryStarCount}
                        </>
                    )
                }
            >
                <RepoName repoName={match.repository} suffix={match.path} />
            </SearchResultLayout>
        </SearchResultHeader>
    )

    return (
        <>
            {title}
            {lines}
        </>
    )
}
