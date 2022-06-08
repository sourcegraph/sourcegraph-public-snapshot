import React from 'react'

import classNames from 'classnames'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'

import { appendSubtreeQueryParameter } from '@sourcegraph/common'
import { CodeHostIcon, formatRepositoryStarCount, SearchResultStar } from '@sourcegraph/search-ui'
import { displayRepoName, splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import { ContentMatch, getFileMatchUrl, SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'
import { Code, Icon, Link, Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import { SearchResultHeader } from './SearchResultHeader'
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
            <div className={styles.lineCode}>
                <TrimmedCodeLineWithHighlights line={line} />
            </div>
            <div className={classNames(styles.lineLineNumber, 'text-muted')}>{line.lineNumber + 1}</div>
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
            <div>
                <SymbolIcon kind={symbol.kind} className="mr-1" />
                <Code>
                    {symbol.name} {symbol.containerName && <span className="text-muted">{symbol.containerName}</span>}
                </Code>
            </div>
        </SelectableSearchResult>
    ))
}

export const FileSearchResult: React.FunctionComponent<Props> = ({ match, selectedResult, selectResult }: Props) => {
    const lines =
        match.type === 'content'
            ? getResultElementsForContentMatch(match, selectResult, selectedResult)
            : getResultElementsForSymbolMatch(match, selectResult, selectedResult)

    const repoDisplayName = match.repository
    const repoAtRevisionURL = '#'
    const formattedRepositoryStarCount = formatRepositoryStarCount(match.repoStars)

    const title = (
        <SearchResultHeader>
            <div className={classNames(styles.headerTitle)} data-testid="result-container-header">
                <Icon role="img" aria-label="File" className="flex-shrink-0" as={FileDocumentIcon} />
                <div className={classNames('mx-1', styles.headerDivider)} />
                <CodeHostIcon repoName={match.repository} className="text-muted flex-shrink-0" />
                <RepoFileLinkWithoutTabStop
                    repoName={match.repository}
                    repoURL={repoAtRevisionURL}
                    filePath={match.path}
                    fileURL={getFileMatchUrl(match)}
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
        </SearchResultHeader>
    )

    return (
        <>
            {title}
            {lines}
        </>
    )
}

/**
 * This is a fork of RepoFileLink with an added tabIndex of -1 so that it's not possible to tab
 * navigate to the individual links (since we want to use manual arrow navigation instead)
 */
interface RepoFileLinkWithoutTabStopProps {
    repoName: string
    repoURL: string
    filePath: string
    fileURL: string
    repoDisplayName?: string
    className?: string
}

const RepoFileLinkWithoutTabStop: React.FunctionComponent<React.PropsWithChildren<RepoFileLinkWithoutTabStopProps>> = ({
    repoDisplayName,
    repoName,
    repoURL,
    filePath,
    fileURL,
    className,
}) => {
    const [fileBase, fileName] = splitPath(filePath)
    /**
     * Use the custom hook useIsTruncated to check for an overflow: ellipsis is activated for the element
     * We want to do it on mouse enter as browser window size might change after the element has been
     * loaded initially
     */
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    return (
        <Tooltip content={truncated ? (fileBase ? `${fileBase}/${fileName}` : fileName) : null}>
            <div ref={titleReference} onMouseEnter={checkTruncation} className={classNames(className)}>
                <Link tabIndex={-1} to={repoURL}>
                    {repoDisplayName || displayRepoName(repoName)}
                </Link>{' '}
                ›{' '}
                <Link tabIndex={-1} to={appendSubtreeQueryParameter(fileURL)}>
                    {fileBase ? `${fileBase}/` : null}
                    <strong>{fileName}</strong>
                </Link>
            </div>
        </Tooltip>
    )
}
