import * as H from 'history'
import React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { IHighlightLineRange } from '../graphql/schema'
import { ContentMatch, SymbolMatch, PathMatch, getFileMatchUrl } from '../search/stream'
import { SettingsCascadeProps } from '../settings/settings'
import { SymbolIcon } from '../symbols/SymbolIcon'
import { TelemetryProps } from '../telemetry/telemetryService'
import { isErrorLike } from '../util/errors'
import {
    appendLineRangeQueryParameter,
    toPositionOrRangeQueryParameter,
    appendSubtreeQueryParameter,
} from '../util/url'

import { CodeExcerpt, FetchFileParameters } from './CodeExcerpt'
import { MatchGroup } from './FileMatchContext'
import { LastSyncedIcon } from './LastSyncedIcon'
import { Link } from './Link'

interface FileMatchProps extends SettingsCascadeProps, TelemetryProps {
    location: H.Location
    result: ContentMatch | SymbolMatch | PathMatch
    grouped: MatchGroup[]
    /* Called when the first result has fully loaded. */
    onFirstResultLoad?: () => void
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    /**
     * Called when the file's search result is selected.
     */
    onSelect: () => void
}

export const FileMatchChildren: React.FunctionComponent<FileMatchProps> = props => {
    // If optimizeHighlighting is enabled, compile a list of the highlighted file ranges we want to
    // fetch (instead of the entire file.)
    const optimizeHighlighting =
        props.settingsCascade.final &&
        !isErrorLike(props.settingsCascade.final) &&
        props.settingsCascade.final.experimentalFeatures &&
        props.settingsCascade.final.experimentalFeatures.enableFastResultLoading

    const { result, grouped, fetchHighlightedFileLineRanges, telemetryService, onFirstResultLoad } = props
    const fetchHighlightedFileRangeLines = React.useCallback(
        (isFirst, startLine, endLine) => {
            const startTime = Date.now()
            return fetchHighlightedFileLineRanges(
                {
                    repoName: result.repository,
                    commitID: result.commit || '',
                    filePath: result.path,
                    disableTimeout: false,
                    ranges: optimizeHighlighting
                        ? grouped.map(
                              (group): IHighlightLineRange => ({
                                  startLine: group.startLine,
                                  endLine: group.endLine,
                              })
                          )
                        : [{ startLine: 0, endLine: 2147483647 }], // entire file,
                },
                false
            ).pipe(
                map(lines => {
                    if (isFirst && onFirstResultLoad) {
                        onFirstResultLoad()
                    }
                    telemetryService.log(
                        'search.latencies.frontend.code-load',
                        { durationMs: Date.now() - startTime },
                        { durationMs: Date.now() - startTime }
                    )
                    return optimizeHighlighting
                        ? lines[grouped.findIndex(group => group.startLine === startLine && group.endLine === endLine)]
                        : lines[0].slice(startLine, endLine)
                })
            )
        },
        [result, fetchHighlightedFileLineRanges, grouped, optimizeHighlighting, telemetryService, onFirstResultLoad]
    )

    const createCodeExcerptLink = (group: MatchGroup): string => {
        const positionOrRangeQueryParameter = toPositionOrRangeQueryParameter({ position: group.position })
        return appendLineRangeQueryParameter(
            appendSubtreeQueryParameter(getFileMatchUrl(result)),
            positionOrRangeQueryParameter
        )
    }

    return (
        <div className="file-match-children">
            {result.repoLastFetched && <LastSyncedIcon lastSyncedTime={result.repoLastFetched} />}
            {/* Path */}
            {result.type === 'path' && (
                <div className="file-match-children__item">
                    <small>Path match</small>
                </div>
            )}

            {/* Symbols */}
            {((result.type === 'symbol' && result.symbols) || []).map(symbol => (
                <Link
                    to={symbol.url}
                    className="file-match-children__item test-file-match-children-item"
                    key={`symbol:${symbol.name}${String(symbol.containerName)}${symbol.url}`}
                >
                    <SymbolIcon kind={symbol.kind} className="icon-inline mr-1" />
                    <code>
                        {symbol.name}{' '}
                        {symbol.containerName && <span className="text-muted">{symbol.containerName}</span>}
                    </code>
                </Link>
            ))}

            {/* Line matches */}
            {grouped && (
                <div>
                    {grouped.map((group, index) => (
                        <div
                            key={`linematch:${getFileMatchUrl(result)}${group.position.line}:${
                                group.position.character
                            }`}
                            className="file-match-children__item-code-wrapper test-file-match-children-item-wrapper"
                        >
                            <Link
                                to={createCodeExcerptLink(group)}
                                className="file-match-children__item file-match-children__item-clickable test-file-match-children-item"
                                onClick={props.onSelect}
                            >
                                <CodeExcerpt
                                    repoName={result.repository}
                                    commitID={result.commit || ''}
                                    filePath={result.path}
                                    startLine={group.startLine}
                                    endLine={group.endLine}
                                    highlightRanges={group.matches}
                                    className="file-match-children__item-code-excerpt"
                                    fetchHighlightedFileRangeLines={fetchHighlightedFileRangeLines}
                                    isFirst={index === 0}
                                    blobLines={group.blobLines}
                                />
                            </Link>
                        </div>
                    ))}
                </div>
            )}
        </div>
    )
}
