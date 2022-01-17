import classNames from 'classnames'
import React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { CodeExcerpt, FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import styles from '@sourcegraph/shared/src/components/FileMatchChildren.module.scss'
import { MatchGroup } from '@sourcegraph/shared/src/components/FileMatchContext'
import { LastSyncedIcon } from '@sourcegraph/shared/src/components/LastSyncedIcon'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { IHighlightLineRange } from '@sourcegraph/shared/src/graphql/schema'
import { ContentMatch, SymbolMatch, PathMatch, getFileMatchUrl } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { SymbolIcon } from '@sourcegraph/shared/src/symbols/SymbolIcon'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import {
    appendLineRangeQueryParameter,
    toPositionOrRangeQueryParameter,
    appendSubtreeQueryParameter,
} from '@sourcegraph/shared/src/util/url'

import { useQueryState } from '..'

interface FileMatchProps extends SettingsCascadeProps, TelemetryProps {
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

    const searchActions = useQueryState(({ actions }) => actions)

    // set selected result's position suffix state for file opening onSelect in SearchResult
    const onSelectFileResult = (group: MatchGroup): void => {
        const codeExcerptLink = createCodeExcerptLink(group)
        if (codeExcerptLink.endsWith('&subtree=true')) {
            searchActions.setSelectedResultPositionSuffix(`?L${group.position.line - 1}:${group.position.character}`)
        }
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
                    onClick={props.onSelect}
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
                                // Link does not work in VSCE so we will need to set positionsuffix states on Click
                                // to open file at the selected group from SearchResult.tsx
                                onClick={() => onSelectFileResult(group)}
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
