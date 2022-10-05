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

function renderResultElementsForContentMatch(
    match: ContentMatch,
    selectedResult: string | null,
    selectResult: (resultId: string) => void,
    openResult: (resultId: string) => void
): JSX.Element[] {
    return match.lineMatches?.map(line => (
        <SelectableSearchResult
            key={getResultId(match, line)}
            lineOrSymbolMatch={line}
            match={match}
            selectedResult={selectedResult}
            selectResult={selectResult}
            openResult={openResult}
        >
            {isActive => (
                <SearchResultLayout infoColumn={line.lineNumber + 1} className={styles.code} isActive={isActive}>
                    <TrimmedCodeLineWithHighlights line={line} />
                </SearchResultLayout>
            )}
        </SelectableSearchResult>
    )) || []
}

interface Props {
    selectResult: (resultId: string) => void
    selectedResult: null | string
    match: ContentMatch | SymbolMatch
    openResult: (resultId: string) => void
}

function renderResultElementsForSymbolMatch(
    match: SymbolMatch,
    selectedResult: string | null,
    selectResult: (resultId: string) => void,
    openResult: (resultId: string) => void
): JSX.Element[] {
    return match.symbols.map(symbol => (
        <SelectableSearchResult
            key={getResultId(match, symbol)}
            lineOrSymbolMatch={symbol}
            match={match}
            selectedResult={selectedResult}
            selectResult={selectResult}
            openResult={openResult}
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

export const FileSearchResult: React.FunctionComponent<Props> = ({
    match,
    selectedResult,
    selectResult,
    openResult,
}: Props) => {
    const lines =
        match.type === 'content'
            ? renderResultElementsForContentMatch(match, selectedResult, selectResult, openResult)
            : renderResultElementsForSymbolMatch(match, selectedResult, selectResult, openResult)

    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)

    const onClick = (): void =>
        lines.length
            ? selectResult(getResultId(match, match.type === 'content' ? match.lineMatches![0] : match.symbols[0]))
            : undefined

    const title = (
        <SearchResultHeader onClick={onClick}>
            <SearchResultLayout
                iconColumn={{
                    icon: match.type === 'content' ? FileDocumentIcon : AlphaSBoxIcon,
                    repoName: match.repository,
                }}
                infoColumn={
                    formattedRepositoryStarCount && (
                        <>
                            <span className={styles.matches}>
                                {lines.length} {lines.length > 1 ? 'matches' : 'match'}
                            </span>
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
