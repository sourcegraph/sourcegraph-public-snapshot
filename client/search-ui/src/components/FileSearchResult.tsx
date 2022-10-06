import React, { useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { Observable } from 'rxjs'
import { AggregableBadge } from 'sourcegraph'

import { HoverMerged } from '@sourcegraph/client-api'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { isErrorLike, pluralize } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { LineRanking } from '@sourcegraph/shared/src/components/ranking/LineRanking'
import { MatchItem } from '@sourcegraph/shared/src/components/ranking/PerFileResultRanking'
import { ZoektRanking } from '@sourcegraph/shared/src/components/ranking/ZoektRanking'
import { Controller as ExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import { ContentMatch, getFileMatchUrl, getRepositoryUrl, getRevision } from '@sourcegraph/shared/src/search/stream'
import { isSettingsValid, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Badge } from '@sourcegraph/wildcard'

import { FetchFileParameters } from './CodeExcerpt'
import { FileMatchChildren } from './FileMatchChildren'
import { LegacyResultContainerProps, LegacyResultContainer } from './LegacyResultContainer'
import { RepoFileLink } from './RepoFileLink'

import styles from './SearchResult.module.scss'

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

    /**
     * CSS class name to be applied to the ResultContainer Component
     */
    containerClassName?: string

    /**
     * Clicking on a match opens the link in a new tab.
     */
    openInNewTab?: boolean

    extensionsController?: Pick<ExtensionsController, 'extHostAPI'> | null

    hoverifier?: Hoverifier<HoverContext, HoverMerged, ActionItemAction>

    as?: React.ElementType
    index: number
}

const sumHighlightRanges = (count: number, item: MatchItem): number => count + item.highlightRanges.length

const BY_LINE_RANKING = 'by-line-number'
const DEFAULT_CONTEXT = 1

type CommonResultContainerProps = Omit<
    LegacyResultContainerProps,
    | 'description'
    | 'collapsedChildren'
    | 'expandedChildren'
    | 'collapsible'
    | 'collapseLabel'
    | 'expandLabel'
    | 'matchCountLabel'
>

// This is a search result for types file (content), path, or symbol.
export const FileSearchResult: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const result = props.result
    const repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    const revisionDisplayName = getRevision(result.branches, result.commit)
    const settings = props.settingsCascade.final

    const ranking = useMemo(() => {
        if (!isErrorLike(settings) && settings?.experimentalFeatures?.clientSearchResultRanking === BY_LINE_RANKING) {
            return new LineRanking(5)
        }
        return new ZoektRanking(3)
    }, [settings])

    // The number of lines of context to show before and after each match.
    const context = useMemo(() => {
        if (props.location?.pathname === '/search') {
            // Check if search.contextLines is configured in settings.
            const contextLinesSetting =
                isSettingsValid(props.settingsCascade) &&
                props.settingsCascade.final &&
                (props.settingsCascade.final['search.contextLines'] as number | undefined)

            if (typeof contextLinesSetting === 'number' && contextLinesSetting >= 0) {
                return contextLinesSetting
            }
        }
        return DEFAULT_CONTEXT
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
                    <Badge
                        key={badge.text}
                        href={badge.linkURL}
                        tooltip={badge.hoverMessage}
                        variant="secondary"
                        small={true}
                        className="text-muted text-uppercase file-match__badge"
                    >
                        {badge.text}
                    </Badge>
                ))}
            </>
        ) : undefined

    const expandedMatchGroups = useMemo(() => ranking.expandedResults(items, context), [items, context, ranking])
    const collapsedMatchGroups = useMemo(() => ranking.collapsedResults(items, context), [items, context, ranking])
    const collapsedMatchCount = collapsedMatchGroups.matches.length

    const highlightRangesCount = useMemo(() => items.reduce(sumHighlightRanges, 0), [items])
    const collapsedHighlightRangesCount = useMemo(() => collapsedMatchGroups.matches.reduce(sumHighlightRanges, 0), [
        collapsedMatchGroups,
    ])

    const matchCount = highlightRangesCount
    const matchCountLabel = matchCount ? `${matchCount} ${pluralize('match', matchCount, 'matches')}` : ''

    const expandedChildren = <FileMatchChildren {...props} result={result} {...expandedMatchGroups} />

    const commonContainerProps: CommonResultContainerProps = {
        index: props.index,
        defaultExpanded: props.expanded,
        icon: props.icon,
        title: (
            <RepoFileLink
                repoName={result.repository}
                repoURL={repoAtRevisionURL}
                filePath={result.path}
                pathMatchRanges={'pathMatches' in result ? result.pathMatches : []}
                fileURL={getFileMatchUrl(result)}
                repoDisplayName={
                    props.repoDisplayName
                        ? `${props.repoDisplayName}${revisionDisplayName ? `@${revisionDisplayName}` : ''}`
                        : undefined
                }
                className={classNames(styles.titleInner, styles.mutedRepoFileLink)}
            />
        ),
        allExpanded: props.allExpanded,
        repoName: result.repository,
        repoStars: result.repoStars,
        repoLastFetched: result.repoLastFetched,
        onResultClicked: props.onSelect,
        className: props.containerClassName,
        resultType: result.type,
    }

    let containerProps: LegacyResultContainerProps

    if (props.showAllMatches) {
        containerProps = {
            ...commonContainerProps,
            collapsible: false,
            description,
            expandedChildren,
            matchCountLabel,
        }
    } else {
        const length = highlightRangesCount - collapsedHighlightRangesCount
        containerProps = {
            ...commonContainerProps,
            collapsible: items.length > collapsedMatchCount,
            description,
            collapsedChildren: <FileMatchChildren {...props} result={result} {...collapsedMatchGroups} />,
            expandedChildren,
            collapseLabel: 'Show less',
            expandLabel: `Show ${length} more ${pluralize('match', length, 'matches')}`,
            matchCountLabel,
            as: props.as,
        }
    }

    return <LegacyResultContainer {...containerProps} titleClassName="test-search-result-label" />
}

function aggregateBadges(items: MatchItem[]): AggregableBadge[] {
    const aggregatedBadges = new Map<string, AggregableBadge>()
    for (const badge of items.flatMap(item => item.aggregableBadges || [])) {
        aggregatedBadges.set(badge.text, badge)
    }

    return [...aggregatedBadges.values()].sort((a, b) => a.text.localeCompare(b.text))
}
