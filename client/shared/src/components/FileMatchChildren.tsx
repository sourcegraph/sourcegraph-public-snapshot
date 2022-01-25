import classNames from 'classnames'
import * as H from 'history'
import React, { useEffect, useMemo } from 'react'
import { Observable, ReplaySubject } from 'rxjs'
import { map } from 'rxjs/operators'

import { Hoverifier } from '@sourcegraph/codeintellify'
import { isErrorLike } from '@sourcegraph/common'
import { Link } from '@sourcegraph/wildcard'

import { ActionItemAction } from '../actions/ActionItem'
import { HoverMerged } from '../api/client/types/hover'
import { ViewerId } from '../api/viewerTypes'
import { Controller as ExtensionsController } from '../extensions/controller'
import { HoverContext } from '../hover/HoverOverlay.types'
import { getModeFromPath } from '../languages'
import { IHighlightLineRange } from '../schema'
import { ContentMatch, SymbolMatch, PathMatch, getFileMatchUrl } from '../search/stream'
import { SettingsCascadeProps } from '../settings/settings'
import { SymbolIcon } from '../symbols/SymbolIcon'
import { TelemetryProps } from '../telemetry/telemetryService'
import {
    appendLineRangeQueryParameter,
    toPositionOrRangeQueryParameter,
    appendSubtreeQueryParameter,
    toURIWithPath,
} from '../util/url'

import { CodeExcerpt, FetchFileParameters } from './CodeExcerpt'
import styles from './FileMatchChildren.module.scss'
import { LastSyncedIcon } from './LastSyncedIcon'
import { MatchGroup } from './ranking/PerFileResultRanking'

interface FileMatchProps extends SettingsCascadeProps, TelemetryProps {
    location: H.Location
    result: ContentMatch | SymbolMatch | PathMatch
    grouped: MatchGroup[]
    /* Called when the first result has fully loaded. */
    onFirstResultLoad?: () => void
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>

    extensionsController?: Pick<ExtensionsController, 'extHostAPI'>
    hoverifier?: Hoverifier<HoverContext, HoverMerged, ActionItemAction>
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

    // Inform the extension host about the file (if we have code to render).
    // Code excerpt will call `hoverifier.hoverify`.
    const viewerUpdates = useMemo(
        () =>
            new ReplaySubject<
                {
                    viewerId: ViewerId
                } & HoverContext
            >(1),
        []
    )
    useEffect(() => {
        if (!props.extensionsController || result.type !== 'content' || !grouped) {
            return
        }

        let previousViewerId: ViewerId | undefined
        const commitID = result.commit || 'HEAD'
        const uri = toURIWithPath({
            repoName: result.repository,
            filePath: result.path,
            commitID,
        })
        const languageId = getModeFromPath(result.path)
        const text = ''
        // HACK: code intel extensions don't depend on the `text` field.
        // Fix to support other hover extensions on search results
        // (likely too expensive).

        props.extensionsController.extHostAPI
            .then(extensionHostAPI =>
                Promise.all([
                    // This call should be made before adding viewer, but since
                    // messages to web worker are handled in order, we can use Promise.all
                    extensionHostAPI.addTextDocumentIfNotExists({
                        uri,
                        languageId,
                        text,
                    }),
                    extensionHostAPI.addViewerIfNotExists({
                        type: 'CodeEditor' as const,
                        resource: uri,
                        selections: [],
                        isActive: true,
                    }),
                ])
            )
            .then(([, viewerId]) => {
                viewerUpdates.next({
                    viewerId,
                    repoName: result.repository,
                    revision: commitID,
                    commitID,
                    filePath: result.path,
                })
            })
            .catch(error => {
                console.error('Extension host API error', error)
            })

        return () => {
            // Remove from extension host
            props.extensionsController?.extHostAPI
                .then(extensionHostAPI => previousViewerId && extensionHostAPI.removeViewer(previousViewerId))
                .catch(error => console.error('Error removing viewer from extension host', error))
        }
    }, [grouped, result, viewerUpdates, props.extensionsController])

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
                                    viewerUpdates={viewerUpdates}
                                    hoverifier={props.hoverifier}
                                />
                            </Link>
                        </div>
                    ))}
                </div>
            )}
        </div>
    )
}
