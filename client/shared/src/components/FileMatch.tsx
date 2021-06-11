import classNames from 'classnames'
import * as H from 'history'
import React, { useEffect, useState } from 'react'
import { Observable } from 'rxjs'
import { AggregableBadge, Badge } from 'sourcegraph'

import { FileLineMatch, FileSymbolMatch, getFileMatchUrl, getRepositoryUrl, getRevision } from '../search/stream'
import { SettingsCascadeProps } from '../settings/settings'
import { pluralize } from '../util/strings'
import { useRedesignToggle } from '../util/useRedesignToggle'

import { FetchFileParameters } from './CodeExcerpt'
import { EventLogger, FileMatchChildren } from './FileMatchChildren'
import { LinkOrSpan } from './LinkOrSpan'
import { RepoFileLink } from './RepoFileLink'
import { RepoIcon } from './RepoIcon'
import { Props as ResultContainerProps, ResultContainer } from './ResultContainer'

const SUBSET_COUNT_KEY = 'fileMatchSubsetCount'

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
}

interface Props extends SettingsCascadeProps {
    location: H.Location
    eventLogger?: EventLogger
    /**
     * The file match search result.
     */
    result: FileLineMatch | FileSymbolMatch

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

    isLightTheme: boolean

    allExpanded?: boolean

    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

export const FileMatch: React.FunctionComponent<Props> = props => {
    const [subsetMatches, setSubsetMatches] = useState(10)
    useEffect(() => {
        const subsetMatches = parseInt(localStorage.getItem(SUBSET_COUNT_KEY) || '', 10)
        if (!isNaN(subsetMatches)) {
            setSubsetMatches(subsetMatches)
        }
    }, [])

    const [isRedesignEnabled] = useRedesignToggle()

    const result = props.result
    const items: MatchItem[] =
        result.type === 'file'
            ? result.lineMatches.map(match => ({
                  highlightRanges: match.offsetAndLengths.map(([start, highlightLength]) => ({
                      start,
                      highlightLength,
                  })),
                  preview: match.line,
                  line: match.lineNumber,
              }))
            : []

    const repoAtRevisionURL = getRepositoryUrl(result.repository, result.branches)
    const revisionDisplayName = getRevision(result.branches, result.version)

    const renderTitle = (): JSX.Element => (
        <>
            {isRedesignEnabled && <RepoIcon repoName={result.repository} className="icon-inline text-muted" />}
            <RepoFileLink
                repoName={result.repository}
                repoURL={repoAtRevisionURL}
                filePath={result.name}
                fileURL={getFileMatchUrl(result)}
                repoDisplayName={
                    props.repoDisplayName
                        ? `${props.repoDisplayName}${revisionDisplayName ? `@${revisionDisplayName}` : ''}`
                        : undefined
                }
                className={isRedesignEnabled ? 'ml-1' : undefined}
            />
        </>
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
                        className={classNames(
                            'badge badge-secondary text-muted text-uppercase file-match__badge',
                            isRedesignEnabled && 'badge-sm'
                        )}
                    >
                        {badge.text}
                    </LinkOrSpan>
                ))}
            </>
        ) : undefined

    let containerProps: ResultContainerProps

    const expandedChildren = (
        <FileMatchChildren {...props} items={items} result={result} allMatches={true} subsetMatches={subsetMatches} />
    )

    const matchCount = items.length || (result.type === 'symbol' ? result.symbols?.length : 0)
    const matchCountLabel = matchCount ? `${matchCount} ${pluralize('match', matchCount, 'matches')}` : ''

    if (props.showAllMatches) {
        containerProps = {
            collapsible: !isRedesignEnabled,
            defaultExpanded: props.expanded,
            icon: props.icon,
            title: renderTitle(),
            description,
            expandedChildren,
            allExpanded: props.allExpanded,
            matchCountLabel,
        }
    } else {
        const length = items.length - subsetMatches
        containerProps = {
            collapsible: items.length > subsetMatches,
            defaultExpanded: props.expanded,
            icon: props.icon,
            title: renderTitle(),
            description,
            collapsedChildren: (
                <FileMatchChildren
                    {...props}
                    items={items}
                    result={result}
                    allMatches={false}
                    subsetMatches={subsetMatches}
                />
            ),
            expandedChildren,
            collapseLabel: isRedesignEnabled
                ? `Hide ${length}`
                : `Hide ${length} ${pluralize('match', length, 'matches')}`,
            expandLabel: isRedesignEnabled
                ? `${length} more`
                : `Show ${length} more ${pluralize('match', length, 'matches')}`,
            allExpanded: props.allExpanded,
            matchCountLabel,
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
