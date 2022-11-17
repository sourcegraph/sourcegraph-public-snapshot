import React, { MouseEvent, KeyboardEvent, useCallback, useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import { useHistory } from 'react-router'
import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { HoverMerged } from '@sourcegraph/client-api'
import { Hoverifier } from '@sourcegraph/codeintellify'
import {
    appendLineRangeQueryParameter,
    appendSubtreeQueryParameter,
    isErrorLike,
    toPositionOrRangeQueryParameter,
} from '@sourcegraph/common'
import { HighlightLineRange, HighlightResponseFormat } from '@sourcegraph/search'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import { MatchGroup } from '@sourcegraph/shared/src/components/ranking/PerFileResultRanking'
import { Controller as ExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import { ContentMatch, SymbolMatch, PathMatch, getFileMatchUrl } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { SymbolTag } from '@sourcegraph/shared/src/symbols/SymbolTag'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { codeCopiedEvent } from '@sourcegraph/shared/src/tracking/event-log-creators'
import { useCodeIntelViewerUpdates } from '@sourcegraph/shared/src/util/useCodeIntelViewerUpdates'
import { useFocusOnLoadedMore } from '@sourcegraph/wildcard'

import { CodeExcerpt, onClickCodeExcerptHref } from './CodeExcerpt'
import { LastSyncedIcon } from './LastSyncedIcon'

import styles from './FileMatchChildren.module.scss'

interface FileMatchProps extends SettingsCascadeProps, TelemetryProps {
    location?: H.Location
    result: ContentMatch | SymbolMatch | PathMatch
    grouped: MatchGroup[]
    /* Clicking on a match opens the link in a new tab */
    openInNewTab?: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
    extensionsController?: Pick<ExtensionsController, 'extHostAPI'> | null
    hoverifier?: Hoverifier<HoverContext, HoverMerged, ActionItemAction>
}

/**
 * A helper function to replicate browser behavior when clicking on links.
 * A very common interaction is to open links in a new in the _background_ via
 * CTRL/CMD + click or middle click.
 * Unfortunately `window.open` doesn't give us much control over how the new
 * window/tab should be opened, and the behavior is inconcistent between
 * browsers.
 * In order to replicate the standard behvior as much as possible this function
 * dynamically creates an `<a>` element and triggers a click event on it.
 */
function openLinkInNewTab(
    url: string,
    event: Pick<MouseEvent, 'ctrlKey' | 'altKey' | 'shiftKey' | 'metaKey'>,
    button: 'primary' | 'middle'
): void {
    const link = document.createElement('a')
    link.href = url
    link.style.display = 'none'
    link.target = '_blank'
    link.rel = 'noopener noreferrer'
    const clickEvent = new window.MouseEvent('click', {
        bubbles: false,
        altKey: event.altKey,
        shiftKey: event.shiftKey,
        // Regarding middle click: Setting "button: 1:" doesn't seem to suffice:
        // Firefox doesn't react to the event at all, Chromium opens the tab in
        // the foreground. So in order to simulate a middle click, we set
        // ctrlKey and metaKey to `true` instead.
        ctrlKey: button === 'middle' ? true : event.ctrlKey,
        metaKey: button === 'middle' ? true : event.metaKey,
        view: window,
    })

    // It looks the link has to be part of the document, otherwise Firefox won't
    // trigger the default behavior (it works without appending in Chromium).
    document.body.append(link)
    link.dispatchEvent(clickEvent)
    link.remove()
}

/**
 * Since we are not using a real link anymore, we have to simulate opening
 * the file in a new tab when the search result is clicked on with the
 * middle mouse button.
 * This handler is bound to the `mouseup` event because the `auxclick`
 * (https://w3c.github.io/uievents/#event-type-auxclick) event is not
 * support by all browsers yet (https://caniuse.com/?search=auxclick)
 */
function navigateToFileOnMiddleMouseButtonClick(event: MouseEvent<HTMLElement>): void {
    const href = event.currentTarget.getAttribute('data-href')
    if (href && event.button === 1) {
        openLinkInNewTab(href, event, 'middle')
    }
}

export const FileMatchChildren: React.FunctionComponent<React.PropsWithChildren<FileMatchProps>> = props => {
    /**
     * If LazyFileResultSyntaxHighlighting is enabled, we fetch plaintext
     * line ranges _alongside_ the typical highlighted line ranges.
     */
    const enableLazyFileResultSyntaxHighlighting =
        props.settingsCascade.final &&
        !isErrorLike(props.settingsCascade.final) &&
        props.settingsCascade.final.experimentalFeatures?.enableLazyFileResultSyntaxHighlighting

    const { result, grouped, fetchHighlightedFileLineRanges, telemetryService, extensionsController } = props

    const fetchFileRangeMatches = useCallback(
        (args: { format?: HighlightResponseFormat; ranges: HighlightLineRange[] }): Observable<string[][]> =>
            fetchHighlightedFileLineRanges(
                {
                    repoName: result.repository,
                    commitID: result.commit || '',
                    filePath: result.path,
                    disableTimeout: false,
                    format: args.format,
                    ranges: args.ranges,
                },
                false
            ),
        [result, fetchHighlightedFileLineRanges]
    )

    const fetchHighlightedFileMatchLineRanges = React.useCallback(
        (startLine: number, endLine: number) => {
            const startTime = Date.now()
            return fetchFileRangeMatches({
                format: HighlightResponseFormat.HTML_HIGHLIGHT,
                ranges: grouped.map(
                    (group): HighlightLineRange => ({
                        startLine: group.startLine,
                        endLine: group.endLine,
                    })
                ),
            }).pipe(
                map(lines => {
                    const endTime = Date.now()
                    telemetryService.log(
                        'search.latencies.frontend.code-load',
                        { durationMs: endTime - startTime },
                        { durationMs: endTime - startTime }
                    )
                    return lines[grouped.findIndex(group => group.startLine === startLine && group.endLine === endLine)]
                })
            )
        },
        [fetchFileRangeMatches, grouped, telemetryService]
    )

    const fetchPlainTextFileMatchLineRanges = React.useCallback(
        (startLine: number, endLine: number) =>
            fetchFileRangeMatches({
                format: HighlightResponseFormat.HTML_PLAINTEXT,
                ranges: grouped.map(
                    (group): HighlightLineRange => ({
                        startLine: group.startLine,
                        endLine: group.endLine,
                    })
                ),
            }).pipe(
                map(
                    lines =>
                        lines[grouped.findIndex(group => group.startLine === startLine && group.endLine === endLine)]
                )
            ),
        [fetchFileRangeMatches, grouped]
    )

    const fetchHighlightedSymbolMatchLineRanges = React.useCallback(
        (startLine: number, endLine: number) => {
            if (result.type !== 'symbol') {
                return of([])
            }

            const startTime = Date.now()
            return fetchFileRangeMatches({
                format: HighlightResponseFormat.HTML_HIGHLIGHT,
                ranges: result.symbols.map(
                    (symbol): HighlightLineRange => ({
                        startLine: symbol.line - 1,
                        endLine: symbol.line,
                    })
                ),
            }).pipe(
                map(lines => {
                    const endTime = Date.now()
                    telemetryService.log(
                        'search.latencies.frontend.code-load',
                        { durationMs: endTime - startTime },
                        { durationMs: endTime - startTime }
                    )
                    return lines[
                        result.symbols.findIndex(symbol => symbol.line - 1 === startLine && symbol.line === endLine)
                    ]
                })
            )
        },
        [result, fetchFileRangeMatches, telemetryService]
    )

    const fetchPlainTextSymbolMatchLineRanges = React.useCallback(
        (startLine: number, endLine: number) => {
            if (result.type !== 'symbol') {
                return of([])
            }

            return fetchFileRangeMatches({
                format: HighlightResponseFormat.HTML_PLAINTEXT,
                ranges: result.symbols.map(
                    (symbol): HighlightLineRange => ({
                        startLine: symbol.line - 1,
                        endLine: symbol.line,
                    })
                ),
            }).pipe(
                map(
                    lines =>
                        lines[
                            result.symbols.findIndex(symbol => symbol.line - 1 === startLine && symbol.line === endLine)
                        ]
                )
            )
        },
        [result, fetchFileRangeMatches]
    )

    const createCodeExcerptLink = (group: MatchGroup): string => {
        const positionOrRangeQueryParameter = toPositionOrRangeQueryParameter({ position: group.position })
        return appendLineRangeQueryParameter(
            appendSubtreeQueryParameter(getFileMatchUrl(result)),
            positionOrRangeQueryParameter
        )
    }

    const codeIntelViewerUpdatesProps = useMemo(
        () =>
            grouped && result.type === 'content' && extensionsController
                ? {
                      extensionsController,
                      repositoryName: result.repository,
                      filePath: result.path,
                      revision: result.commit,
                  }
                : undefined,
        [extensionsController, result, grouped]
    )
    const viewerUpdates = useCodeIntelViewerUpdates(codeIntelViewerUpdatesProps)

    const history = useHistory()
    const navigateToFile = useCallback(
        (event: KeyboardEvent<HTMLElement> | MouseEvent<HTMLElement>): void => {
            onClickCodeExcerptHref(event, href => {
                if (props.openInNewTab || event.ctrlKey || event.metaKey || event.shiftKey) {
                    openLinkInNewTab(href, event, 'primary')
                } else {
                    history.push(href)
                }
            })
        },
        [props.openInNewTab, history]
    )

    const logEventOnCopy = useCallback(() => {
        telemetryService.log(...codeCopiedEvent('file-match'))
    }, [telemetryService])

    const getItemRef = useFocusOnLoadedMore<HTMLDivElement>(grouped.length ?? 0)

    return (
        <div
            className={classNames(styles.fileMatchChildren, result.type === 'symbol' && styles.symbols)}
            data-testid="file-match-children"
        >
            {result.repoLastFetched && <LastSyncedIcon lastSyncedTime={result.repoLastFetched} />}
            {/* Path */}
            {result.type === 'path' && (
                <div className={styles.item} data-testid="file-match-children-item">
                    <small>{result.pathMatches ? 'Path match' : 'File contains matching content'}</small>
                </div>
            )}

            {/* Symbols */}
            {((result.type === 'symbol' && result.symbols) || []).map(symbol => (
                <div
                    key={`symbol:${symbol.name}${String(symbol.containerName)}${symbol.url}`}
                    className={classNames('test-file-match-children-item', styles.symbol)}
                    data-href={symbol.url}
                    role="link"
                    tabIndex={0}
                    onClick={navigateToFile}
                    onMouseUp={navigateToFileOnMiddleMouseButtonClick}
                    onKeyDown={navigateToFile}
                >
                    <div className="mr-2 flex-shrink-0">
                        <SymbolTag kind={symbol.kind} />
                    </div>
                    <div className={styles.symbolCodeExcerpt}>
                        <CodeExcerpt
                            className="a11y-ignore"
                            repoName={result.repository}
                            commitID={result.commit || ''}
                            filePath={result.path}
                            startLine={symbol.line - 1}
                            endLine={symbol.line}
                            fetchHighlightedFileRangeLines={fetchHighlightedSymbolMatchLineRanges}
                            fetchPlainTextFileRangeLines={
                                enableLazyFileResultSyntaxHighlighting ? fetchPlainTextSymbolMatchLineRanges : undefined
                            }
                            viewerUpdates={viewerUpdates}
                            hoverifier={props.hoverifier}
                            onCopy={logEventOnCopy}
                            highlightRanges={[]}
                        />
                    </div>
                </div>
            ))}

            {/* Line matches */}
            {grouped.length > 0 && (
                <div>
                    {grouped.map((group, index) => (
                        <div
                            key={`linematch:${getFileMatchUrl(result)}${group.position.line}:${
                                group.position.character
                            }`}
                            className={classNames('test-file-match-children-item-wrapper', styles.itemCodeWrapper)}
                        >
                            <div
                                data-href={createCodeExcerptLink(group)}
                                className={classNames(
                                    'test-file-match-children-item',
                                    styles.item,
                                    styles.itemClickable
                                )}
                                onClick={navigateToFile}
                                onMouseUp={navigateToFileOnMiddleMouseButtonClick}
                                onKeyDown={navigateToFile}
                                data-testid="file-match-children-item"
                                tabIndex={0}
                                role="link"
                                ref={getItemRef(index)}
                            >
                                <CodeExcerpt
                                    repoName={result.repository}
                                    commitID={result.commit || ''}
                                    filePath={result.path}
                                    startLine={group.startLine}
                                    endLine={group.endLine}
                                    highlightRanges={group.matches}
                                    fetchHighlightedFileRangeLines={fetchHighlightedFileMatchLineRanges}
                                    fetchPlainTextFileRangeLines={
                                        enableLazyFileResultSyntaxHighlighting
                                            ? fetchPlainTextFileMatchLineRanges
                                            : undefined
                                    }
                                    blobLines={group.blobLines}
                                    viewerUpdates={viewerUpdates}
                                    hoverifier={props.hoverifier}
                                    onCopy={logEventOnCopy}
                                />
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    )
}
