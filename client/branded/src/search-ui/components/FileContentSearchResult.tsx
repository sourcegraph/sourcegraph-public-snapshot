import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'
import type * as H from 'history'
import VisibilitySensor from 'react-visibility-sensor'
import type { Observable, Subscription } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { asError, isErrorLike, pluralize } from '@sourcegraph/common'
import type { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import type { MatchGroup } from '@sourcegraph/shared/src/components/ranking/PerFileResultRanking'
import { HighlightResponseFormat } from '@sourcegraph/shared/src/graphql-operations'
import {
    type ContentMatch,
    type ChunkMatch,
    getFileMatchUrl,
    getRepositoryUrl,
    getRevision,
} from '@sourcegraph/shared/src/search/stream'
import { type SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Icon } from '@sourcegraph/wildcard'

import { CopyPathAction } from './CopyPathAction'
import { FileMatchChildren } from './FileMatchChildren'
import { RepoFileLink } from './RepoFileLink'
import { ResultContainer } from './ResultContainer'
import { SearchResultPreviewButton } from './SearchResultPreviewButton'

import resultContainerStyles from './ResultContainer.module.scss'
import styles from './SearchResult.module.scss'

const DEFAULT_VISIBILITY_OFFSET = { bottom: -500 }

interface Props extends SettingsCascadeProps, TelemetryProps {
    location: H.Location
    /**
     * The file match search result.
     */
    result: ContentMatch

    /**
     * Formatted repository name to be displayed in repository link. If not
     * provided, the default format will be displayed.
     */
    repoDisplayName?: string

    /**
     * Called when the file's search result is selected.
     */
    onSelect: () => void

    /**
     * Whether this file should be rendered as expanded by default.
     */
    defaultExpanded: boolean

    /**
     * Whether or not to show all matches for this file, or only a subset.
     */
    showAllMatches: boolean

    allExpanded?: boolean

    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>

    /**
     * CSS class name to be applied to the ResultContainer Component
     */
    containerClassName?: string

    /**
     * Clicking on a match opens the link in a new tab.
     */
    openInNewTab?: boolean

    index: number
}

const BY_LINE_RANKING = 'by-line-number'

export const FileContentSearchResult: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    containerClassName,
    result,
    settingsCascade,
    location,
    index,
    repoDisplayName,
    defaultExpanded,
    allExpanded,
    showAllMatches,
    openInNewTab,
    telemetryService,
    fetchHighlightedFileLineRanges,
    onSelect,
}) => {
    const repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    const revisionDisplayName = getRevision(result.branches, result.commit)

    const reranker = useMemo(() => {
        const settings = settingsCascade.final
        if (!isErrorLike(settings) && settings?.experimentalFeatures?.clientSearchResultRanking === BY_LINE_RANKING) {
            return rankByLine
        }
        return rankPassthrough
    }, [settingsCascade])

    const newSearchUIEnabled = useMemo(() => {
        const settings = settingsCascade.final
        if (!isErrorLike(settings)) {
            return settings?.experimentalFeatures?.newSearchNavigationUI
        }
        return false
    }, [settingsCascade])

    const [hasBeenVisible, setHasBeenVisible] = useState(false)

    const unhighlightedGroups: MatchGroup[] = useMemo(
        () => reranker(result.chunkMatches?.map(chunkToMatchGroup) ?? []),
        [result, reranker]
    )

    const [groups, setGroups] = useState(unhighlightedGroups)
    useEffect(() => {
        let subscription: Subscription | undefined
        if (hasBeenVisible) {
            subscription = fetchHighlightedFileLineRanges(
                {
                    repoName: result.repository,
                    commitID: result.commit || '',
                    filePath: result.path,
                    disableTimeout: false,
                    format: HighlightResponseFormat.HTML_HIGHLIGHT,
                    ranges: unhighlightedGroups,
                },
                false
            )
                .pipe(catchError(error => [asError(error)]))
                .subscribe(res => {
                    if (!isErrorLike(res)) {
                        setGroups(
                            unhighlightedGroups.map((group, i) => ({
                                ...group,
                                highlightedLines: res[i],
                            }))
                        )
                    }
                })
        }
        return () => subscription?.unsubscribe()
    }, [result, unhighlightedGroups, hasBeenVisible, fetchHighlightedFileLineRanges])

    const collapsedMatchGroups = truncateGroups(groups, 3)

    const highlightRangesCount = groups.reduce(sumHighlightRanges, 0)
    const collapsedHighlightRangesCount = collapsedMatchGroups.reduce(sumHighlightRanges, 0)

    const hiddenMatchesCount = highlightRangesCount - collapsedHighlightRangesCount
    const collapsible = !showAllMatches && highlightRangesCount > collapsedHighlightRangesCount

    const [expanded, setExpanded] = useState(allExpanded || defaultExpanded)
    useEffect(() => setExpanded(allExpanded || defaultExpanded), [allExpanded, defaultExpanded])

    const rootRef = useRef<HTMLDivElement>(null)
    const toggle = useCallback((): void => {
        if (collapsible) {
            setExpanded(expanded => !expanded)
        }

        // Scroll back to top of result when collapsing
        if (expanded) {
            setTimeout(() => {
                const reducedMotion = !window.matchMedia('(prefers-reduced-motion: no-preference)').matches
                rootRef.current?.scrollIntoView({ block: 'nearest', behavior: reducedMotion ? 'auto' : 'smooth' })
            }, 0)
        }
    }, [collapsible, expanded])

    const title = (
        <>
            <span className="d-flex align-items-center">
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
                    className={styles.titleInner}
                />
                <CopyPathAction
                    className={styles.copyButton}
                    filePath={result.path}
                    telemetryService={telemetryService}
                />
            </span>
        </>
    )

    useEffect(() => {
        const ref = rootRef.current
        if (!ref) {
            return
        }

        const expand = (): void => setExpanded(true)
        const collapse = (): void => setExpanded(false)
        const toggle = (): void => setExpanded(expanded => !expanded)

        // Custom events triggered by search results keyboard navigation (from the useSearchResultsKeyboardNavigation hook).
        ref.addEventListener('expandSearchResultsGroup', expand)
        ref.addEventListener('collapseSearchResultsGroup', collapse)
        ref.addEventListener('toggleSearchResultsGroup', toggle)

        return () => {
            ref.removeEventListener('expandSearchResultsGroup', expand)
            ref.removeEventListener('collapseSearchResultsGroup', collapse)
            ref.removeEventListener('toggleSearchResultsGroup', toggle)
        }
    }, [rootRef, setExpanded])

    return (
        <ResultContainer
            ref={rootRef}
            index={index}
            title={title}
            resultType={result.type}
            onResultClicked={onSelect}
            repoName={result.repository}
            repoStars={result.repoStars}
            className={classNames(styles.copyButtonContainer, containerClassName)}
            resultClassName={resultContainerStyles.highlightResult}
            rankingDebug={result.debug}
            repoLastFetched={result.repoLastFetched}
            actions={newSearchUIEnabled && <SearchResultPreviewButton result={result} />}
        >
            {/* <div data-testid="file-search-result" data-expanded={expanded}> */}
            <VisibilitySensor
                onChange={(visible: boolean) => setHasBeenVisible(visible || hasBeenVisible)}
                partialVisibility={true}
                offset={DEFAULT_VISIBILITY_OFFSET}
            >
                <div data-testid="file-search-result">
                    <FileMatchChildren
                        result={result}
                        grouped={expanded ? groups : collapsedMatchGroups}
                        settingsCascade={settingsCascade}
                        telemetryService={telemetryService}
                        openInNewTab={openInNewTab}
                    />
                    {collapsible && (
                        <button
                            type="button"
                            className={classNames(
                                styles.toggleMatchesButton,
                                expanded && styles.toggleMatchesButtonExpanded
                            )}
                            onClick={toggle}
                            data-testid="toggle-matches-container"
                        >
                            <Icon aria-hidden={true} svgPath={expanded ? mdiChevronUp : mdiChevronDown} />
                            <span className={styles.toggleMatchesButtonText}>
                                {expanded
                                    ? 'Show less'
                                    : `Show ${hiddenMatchesCount} more ${pluralize(
                                          'match',
                                          hiddenMatchesCount,
                                          'matches'
                                      )}`}
                            </span>
                        </button>
                    )}
                </div>
            </VisibilitySensor>
        </ResultContainer>
    )
}

// rankPassthrough is a no-op re-ranker
const rankPassthrough = (groups: MatchGroup[]): MatchGroup[] => groups

// rankByLine re-ranks a set of groups to order them by starting line number
const rankByLine = (groups: MatchGroup[]): MatchGroup[] => {
    const groupsCopy = [...groups]
    groupsCopy.sort((a, b) => {
        if (a.startLine < b.startLine) {
            return -1
        }
        if (a.startLine > b.startLine) {
            return 1
        }
        return 0
    })
    return groupsCopy
}

const truncateGroups = (groups: MatchGroup[], maxMatches: number): MatchGroup[] => {
    const visibleGroups = []
    let visibleMatches = 0
    for (const group of groups) {
        if (visibleMatches > maxMatches) {
            break
        }
        visibleGroups.push(group)
        visibleMatches += 1
    }

    return visibleGroups
}

const chunkToMatchGroup = (chunk: ChunkMatch): MatchGroup => {
    const matches = chunk.ranges.map(range => ({
        startLine: range.start.line,
        startCharacter: range.start.column,
        endLine: range.end.line,
        endCharacter: range.end.column,
    }))

    const plaintextLines = chunk.content.split(/\r?\n/)
    let minPosition = { line: Number.MAX_VALUE, character: Number.MAX_VALUE }
    for (const match of matches) {
        const position = { line: match.startLine + 1, character: match.startCharacter + 1 }
        if (position.line <= minPosition.line && position.character < minPosition.character) {
            minPosition = position
        }
    }

    return {
        plaintextLines,
        // Populated asynchronously in the useEffect below
        highlightedLines: undefined,
        matches,
        position: minPosition,
        startLine: chunk.contentStart.line,
        endLine: chunk.contentStart.line + plaintextLines.length,
    }
}

const sumHighlightRanges = (count: number, group: MatchGroup): number => count + group.matches.length
