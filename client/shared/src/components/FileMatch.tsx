import * as H from 'history'
import React, { useMemo } from 'react'
import { Observable } from 'rxjs'
import { AggregableBadge, Badge } from 'sourcegraph'

import { ContentMatch, SymbolMatch, PathMatch, getFileMatchUrl, getRepositoryUrl, getRevision } from '../search/stream'
import { isSettingsValid, SettingsCascadeProps } from '../settings/settings'
import { TelemetryProps } from '../telemetry/telemetryService'
import { pluralize } from '../util/strings'

import { FetchFileParameters } from './CodeExcerpt'
import { FileMatchChildren } from './FileMatchChildren'
import { MatchGroup, calculateMatchGroups } from './FileMatchContext'
import { LinkOrSpan } from './LinkOrSpan'
import { RepoFileLink } from './RepoFileLink'
import { RepoIcon } from './RepoIcon'
import { Props as ResultContainerProps, ResultContainer } from './ResultContainer'

const SUBSET_MATCHES_COUNT = 10

export interface MatchItem extends Badge {
    highlightRanges: {
        start: number
        highlightLength: number
    }[]
    preview: string
    /**
     * The 0-based line number of this match.
     */
    line: number
    aggregableBadges?: AggregableBadge[]
}

interface Props extends SettingsCascadeProps, TelemetryProps {
    location: H.Location
    /**
     * The file match search result.
     */
    result: ContentMatch | SymbolMatch | PathMatch

    /**
     * Formatted repository name to be displayed in repository link. If not
     * provided, the default format will be displayed.
     */
    repoDisplayName?: string

    /**
     * The icon to show to the left of the title.
     */
    icon: React.ComponentType<{ className?: string }>

    /**
     * Called when the file's search result is selected.
     */
    onSelect: () => void

    /**
     * Whether this file should be rendered as expanded.
     */
    expanded: boolean

    /**
     * Whether or not to show all matches for this file, or only a subset.
     */
    showAllMatches: boolean

    allExpanded?: boolean

    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

const sumHighlightRanges = (count: number, item: MatchItem): number => count + item.highlightRanges.length

export const FileMatch: React.FunctionComponent<Props> = props => {
    const result = props.result
    const repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    const revisionDisplayName = getRevision(result.branches, result.commit)
    const renderTitle = (): JSX.Element => (
        <>
            <RepoIcon repoName={result.repository} className="icon-inline text-muted" />
            <RepoFileLink
                repoName={result.repository}
                repoURL={repoAtRevisionURL}
                filePath={result.path}
                fileURL={getFileMatchUrl(result)}
                repoDisplayName={
                    props.repoDisplayName
                        ? `${props.repoDisplayName}${revisionDisplayName ? `@${revisionDisplayName}` : ''}`
                        : undefined
                }
                className="ml-1"
            />
        </>
    )

    // The number of lines of context to show before and after each match.
    const context = useMemo(() => {
        if (props.location.pathname === '/search') {
            // Check if search.contextLines is configured in settings.
            const contextLinesSetting =
                isSettingsValid(props.settingsCascade) &&
                props.settingsCascade.final &&
                props.settingsCascade.final['search.contextLines']

            if (typeof contextLinesSetting === 'number' && contextLinesSetting >= 0) {
                return contextLinesSetting
            }
        }
        return 1
    }, [props.location, props.settingsCascade])

    const items: MatchItem[] = useMemo(
        () =>
            result.type === 'content'
                ? result.lineMatches?.map(match => ({
                      highlightRanges: match.offsetAndLengths.map(([start, highlightLength]) => ({
                          start,
                          highlightLength,
                      })),
                      preview: match.line,
                      line: match.lineNumber,
                      aggregableBadges: match.aggregableBadges,
                  })) || []
                : [],
        [result]
    )

    const description =
        items.length > 0 ? (
            <>
                {aggregateBadges(items).map(badge => (
                    <LinkOrSpan
                        key={badge.text}
                        to={badge.linkURL}
                        target="_blank"
                        rel="noopener noreferrer"
                        data-tooltip={badge.hoverMessage}
                        className="badge badge-secondary badge-sm text-muted text-uppercase file-match__badge"
                    >
                        {badge.text}
                    </LinkOrSpan>
                ))}
            </>
        ) : undefined

    let containerProps: ResultContainerProps

    const expandedMatchGroups = useMemo(() => calculateMatchGroups(items, 0, context), [items, context])
    const collapsedMatchGroups = useMemo(() => calculateMatchGroups(items, SUBSET_MATCHES_COUNT, context), [
        items,
        context,
    ])

    const highlightRangesCount = useMemo(() => items.reduce(sumHighlightRanges, 0), [items])
    const collapsedHighlightRangesCount = useMemo(() => collapsedMatchGroups.matches.reduce(sumHighlightRanges, 0), [
        collapsedMatchGroups,
    ])

    const matchCount = highlightRangesCount || (result.type === 'symbol' ? result.symbols?.length : 0)
    const matchCountLabel = matchCount ? `${matchCount} ${pluralize('match', matchCount, 'matches')}` : ''

    const expandedChildren = <FileMatchChildren {...props} result={result} {...expandedMatchGroups} />

    if (result.type === 'content' && result.hunks) {
        // We should only get here if the new streamed highlight format is sent
        const grouped: MatchGroup[] =
            result.hunks?.map(
                hunk =>
                    ({
                        blobLines: hunk.content.html?.split(/\r?\n/),
                        matches: hunk.matches.map(match => ({
                            line: match.start.line,
                            character: match.start.column,
                            highlightLength: match.end.column - match.start.column,
                            isInContext: false, // TODO(camdencheek) what is this for?
                        })),
                        startLine: hunk.lineStart,
                        endLine: hunk.lineStart + hunk.lineCount,
                        position: {
                            line: hunk.matches[0].start.line + hunk.lineStart + 1,
                            character: hunk.matches[0].start.column + 1,
                        },
                    } as MatchGroup)
            ) || []

        const matchCount = grouped.reduce((previous, group) => previous + group.matches.length, 0)
        const matchCountLabel = `${matchCount} ${pluralize('match', matchCount, 'matches')}`

        const { limitedGrouped, limitedMatchCount } = grouped.reduce(
            (previous, group) => {
                const remaining = SUBSET_MATCHES_COUNT - previous.limitedMatchCount
                if (remaining <= 0) {
                    return previous
                }

                if (group.matches.length <= remaining) {
                    // We have room for the whole group
                    previous.limitedGrouped.push(group)
                    previous.limitedMatchCount += group.matches.length
                    return previous
                }

                const limitedGroup = limitGroup(group, remaining)
                previous.limitedGrouped.push(limitedGroup)
                previous.limitedMatchCount += limitedGroup.matches.length
                return previous
            },
            { limitedGrouped: [] as MatchGroup[], limitedMatchCount: 0 }
        )

        if (props.showAllMatches) {
            containerProps = {
                collapsible: false,
                defaultExpanded: props.expanded,
                icon: props.icon,
                title: renderTitle(),
                description: undefined, // TODO we need badges for the descripiton
                allExpanded: props.allExpanded,
                collapsedChildren: <FileMatchChildren {...props} result={result} grouped={limitedGrouped} />,
                expandedChildren: <FileMatchChildren {...props} result={result} grouped={grouped} />,
                matchCountLabel,
                repoStars: result.repoStars,
                repoLastFetched: result.repoLastFetched,
            }
        } else {
            const hideCount = matchCount - limitedMatchCount
            containerProps = {
                collapsible: limitedMatchCount < matchCount,
                defaultExpanded: props.expanded,
                icon: props.icon,
                title: renderTitle(),
                description: undefined,
                collapsedChildren: <FileMatchChildren {...props} result={result} grouped={limitedGrouped} />,
                expandedChildren: <FileMatchChildren {...props} result={result} grouped={grouped} />,
                collapseLabel: `Hide ${hideCount}`,
                expandLabel: `${hideCount} more`,
                allExpanded: props.allExpanded,
                matchCountLabel,
                repoStars: result.repoStars,
                repoLastFetched: result.repoLastFetched,
            }
        }
    } else if (props.showAllMatches) {
        containerProps = {
            collapsible: false,
            defaultExpanded: props.expanded,
            icon: props.icon,
            title: renderTitle(),
            description,
            expandedChildren,
            allExpanded: props.allExpanded,
            matchCountLabel,
            repoStars: result.repoStars,
            repoLastFetched: result.repoLastFetched,
        }
    } else {
        const length = highlightRangesCount - collapsedHighlightRangesCount
        containerProps = {
            collapsible: items.length > SUBSET_MATCHES_COUNT,
            defaultExpanded: props.expanded,
            icon: props.icon,
            title: renderTitle(),
            description,
            collapsedChildren: <FileMatchChildren {...props} result={result} {...collapsedMatchGroups} />,
            expandedChildren,
            collapseLabel: `Hide ${length}`,
            expandLabel: `${length} more`,
            allExpanded: props.allExpanded,
            matchCountLabel,
            repoStars: result.repoStars,
            repoLastFetched: result.repoLastFetched,
        }
    }

    return <ResultContainer {...containerProps} titleClassName="test-search-result-label" />
}

function aggregateBadges(items: MatchItem[]): AggregableBadge[] {
    const aggregatedBadges = new Map<string, AggregableBadge>()
    for (const badge of items.flatMap(item => item.aggregableBadges || [])) {
        aggregatedBadges.set(badge.text, badge)
    }

    return [...aggregatedBadges.values()].sort((a, b) => a.text.localeCompare(b.text))
}

export function limitGroup(group: MatchGroup, limit: number): MatchGroup {
    if (limit < 1 || group.matches.length === 0) {
        throw new Error('cannot limit a group to less than one match')
    }

    if (group.matches.length <= limit) {
        return group
    }

    // Do a somewhat deep copy of the group so we can mutate it
    const partialGroup: MatchGroup = {
        blobLines: [...(group.blobLines || [])],
        matches: [...group.matches],
        position: { ...group.position },
        startLine: group.startLine,
        endLine: group.endLine,
    }

    partialGroup.matches = partialGroup.matches.slice(0, limit)

    // Add matches on the same line and next line (context line) as the limited match
    const [lastMatch] = partialGroup.matches.slice(-1)
    for (const match of group.matches.slice(limit, undefined)) {
        if (match.line <= lastMatch.line + 1) {
            // include an extra context line
            partialGroup.matches.push(match)
            continue
        }
        break
    }
    partialGroup.endLine = lastMatch.line + 2 // include an extra context line
    partialGroup.blobLines = partialGroup.blobLines?.slice(0, partialGroup.endLine - partialGroup.startLine)
    return partialGroup
}
