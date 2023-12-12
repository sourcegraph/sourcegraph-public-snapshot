import React, { useCallback, type KeyboardEvent, type MouseEvent } from 'react'

import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'
import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { appendLineRangeQueryParameter, isErrorLike, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import type { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import type { MatchGroup } from '@sourcegraph/shared/src/components/ranking/PerFileResultRanking'
import { type HighlightLineRange, HighlightResponseFormat } from '@sourcegraph/shared/src/graphql-operations'
import { type ContentMatch, getFileMatchUrl } from '@sourcegraph/shared/src/search/stream'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { codeCopiedEvent } from '@sourcegraph/shared/src/tracking/event-log-creators'

import { CodeExcerpt } from './CodeExcerpt'
import { navigateToCodeExcerpt, navigateToFileOnMiddleMouseButtonClick } from './codeLinkNavigation'

import styles from './FileMatchChildren.module.scss'

interface FileMatchProps extends SettingsCascadeProps, TelemetryProps {
    result: ContentMatch
    grouped: MatchGroup[]
    /* Clicking on a match opens the link in a new tab */
    openInNewTab?: boolean
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
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

    const { result, grouped, fetchHighlightedFileLineRanges, telemetryService, telemetryRecorder } = props

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
                    telemetryRecorder.recordEvent('search.latencies.frontend.code-load', 'loaded', {
                        metadata: { durationMs: endTime - startTime },
                    })
                    return lines[grouped.findIndex(group => group.startLine === startLine && group.endLine === endLine)]
                })
            )
        },
        [fetchFileRangeMatches, grouped, telemetryService, telemetryRecorder]
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

    const createCodeExcerptLink = (group: MatchGroup): string => {
        const positionOrRangeQueryParameter = toPositionOrRangeQueryParameter({ position: group.position })
        return appendLineRangeQueryParameter(getFileMatchUrl(result), positionOrRangeQueryParameter)
    }

    const navigate = useNavigate()
    const navigateToFile = useCallback(
        (event: KeyboardEvent<HTMLElement> | MouseEvent<HTMLElement>): void =>
            navigateToCodeExcerpt(event, props.openInNewTab ?? false, navigate),
        [props.openInNewTab, navigate]
    )

    const logEventOnCopy = useCallback(() => {
        telemetryService.log(...codeCopiedEvent('file-match'))
        telemetryRecorder.recordEvent('CodeCopied', 'copied')
    }, [telemetryService, telemetryRecorder])

    return (
        <div
            className={styles.fileMatchChildren}
            data-testid="file-match-children"
            data-selectable-search-results-group="true"
        >
            {/* Line matches */}
            {grouped.length > 0 && (
                <div>
                    {grouped.map(group => (
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
                                data-selectable-search-result="true"
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
