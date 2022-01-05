import classNames from 'classnames'
import * as H from 'history'
import React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { isErrorLike } from '@sourcegraph/common'

import { IHighlightLineRange } from '../graphql/schema'
import { ContentMatch, SymbolMatch, PathMatch, getFileMatchUrl } from '../search/stream'
import { SettingsCascadeProps } from '../settings/settings'
import { SymbolIcon } from '../symbols/SymbolIcon'
import { TelemetryProps } from '../telemetry/telemetryService'
import {
    appendLineRangeQueryParameter,
    toPositionOrRangeQueryParameter,
    appendSubtreeQueryParameter,
} from '../util/url'

import { CodeExcerpt, FetchFileParameters } from './CodeExcerpt'
import styles from './FileMatchChildren.module.scss'
import { LastSyncedIcon } from './LastSyncedIcon'
import { Link } from './Link'
import { MatchGroup } from './ranking/PerFileResultRanking'

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
        <div className={styles.fileMatchChildren} data-testid="file-match-children">
            {result.repoLastFetched && <LastSyncedIcon lastSyncedTime={result.repoLastFetched} />}
            {/* Path */}
            {result.type === 'path' && (
                <div className={styles.item} data-testid="file-match-children-item">
                    <small>Path match</small>
                </div>
            )}

            {/* Symbols */}
            {((result.type === 'symbol' && result.symbols) || []).map(symbol => (
                <Link
                    to={symbol.url}
                    className={classNames('test-file-match-children-item', styles.item)}
                    key={`symbol:${symbol.name}${String(symbol.containerName)}${symbol.url}`}
                    data-testid="file-match-children-item"
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
                            className={classNames('test-file-match-children-item-wrapper', styles.itemCodeWrapper)}
                        >
                            <Link
                                to={createCodeExcerptLink(group)}
                                className={classNames(
                                    'test-file-match-children-item',
                                    styles.item,
                                    styles.itemClickable
                                )}
                                onClick={props.onSelect}
                                data-testid="file-match-children-item"
                            >
                                <CodeExcerpt
                                    repoName={result.repository}
                                    commitID={result.commit || ''}
                                    filePath={result.path}
                                    startLine={group.startLine}
                                    endLine={group.endLine}
                                    highlightRanges={group.matches}
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
