import * as H from 'history'
import React, { useEffect, useState } from 'react'
import { Observable } from 'rxjs'
import { AggregableBadge, Badge } from 'sourcegraph'

import * as GQL from '../graphql/schema'
import { SettingsCascadeProps } from '../settings/settings'
import { pluralize } from '../util/strings'

import { FetchFileParameters } from './CodeExcerpt'
import { EventLogger, FileMatchChildren } from './FileMatchChildren'
import { LinkOrSpan } from './LinkOrSpan'
import { RepoFileLink } from './RepoFileLink'
import { Props as ResultContainerProps, ResultContainer } from './ResultContainer'

const SUBSET_COUNT_KEY = 'fileMatchSubsetCount'

export type FileLineMatch = Partial<Pick<GQL.IFileMatch, 'revSpec' | 'symbols' | 'limitHit'>> & {
    file: Pick<GQL.IFile, 'path' | 'url'> & { commit: Pick<GQL.IGitCommit, 'oid'> }
    repository: Pick<GQL.IRepository, 'name' | 'url'>
    lineMatches: LineMatch[]
}

export type LineMatch = Pick<GQL.ILineMatch, 'preview' | 'lineNumber' | 'offsetAndLengths' | 'limitHit'> & Badge

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
    result: FileLineMatch

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

    const result = props.result
    const items: MatchItem[] = props.result.lineMatches.map(match => ({
        highlightRanges: match.offsetAndLengths.map(([start, highlightLength]) => ({ start, highlightLength })),
        preview: match.preview,
        line: match.lineNumber,
        aggregableBadges: match.aggregableBadges,
    }))

    const { repoAtRevURL, revDisplayName } =
        result.revSpec?.__typename === 'GitRevSpecExpr' && result.revSpec.object?.commit
            ? { repoAtRevURL: result.revSpec.object?.commit?.url, revDisplayName: result.revSpec.expr }
            : result.revSpec?.__typename === 'GitRef'
            ? { repoAtRevURL: result.revSpec.url, revDisplayName: result.revSpec.displayName }
            : { repoAtRevURL: result.repository.url, revDisplayName: '' }

    const title = (
        <RepoFileLink
            repoName={result.repository.name}
            repoURL={repoAtRevURL}
            filePath={result.file.path}
            fileURL={result.file.url}
            repoDisplayName={
                props.repoDisplayName
                    ? `${props.repoDisplayName}${revDisplayName ? `@${revDisplayName}` : ''}`
                    : undefined
            }
        />
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
                        className="badge badge-secondary text-muted text-uppercase file-match__badge"
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

    if (props.showAllMatches) {
        containerProps = {
            collapsible: true,
            defaultExpanded: props.expanded,
            icon: props.icon,
            title,
            description,
            expandedChildren,
            allExpanded: props.allExpanded,
        }
    } else {
        const length = items.length - subsetMatches
        containerProps = {
            collapsible: items.length > subsetMatches,
            defaultExpanded: props.expanded,
            icon: props.icon,
            title,
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
            collapseLabel: `Hide ${length} ${pluralize('match', length, 'matches')}`,
            expandLabel: `Show ${length} more ${pluralize('match', length, 'matches')}`,
            allExpanded: props.allExpanded,
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
