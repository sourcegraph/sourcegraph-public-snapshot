import React, { useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import { range } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import VisibilitySensor from 'react-visibility-sensor'
import { Observable, NEVER, BehaviorSubject } from 'rxjs'
import { catchError, filter } from 'rxjs/operators'

import { HoverMerged } from '@sourcegraph/client-api'
import { DOMFunctions, findPositionsFromEvents, Hoverifier } from '@sourcegraph/codeintellify'
import { asError, ErrorLike, isDefined, isErrorLike, highlightNode } from '@sourcegraph/common'
import { Icon, useObservable, Typography } from '@sourcegraph/wildcard'

import { ActionItemAction } from '../actions/ActionItem'
import { ViewerId } from '../api/viewerTypes'
import { HoverContext } from '../hover/HoverOverlay.types'
import * as GQL from '../schema'
import { Repo } from '../util/url'

import styles from './CodeExcerpt.module.scss'

export interface FetchFileParameters {
    repoName: string
    commitID: string
    filePath: string
    disableTimeout?: boolean
    ranges: GQL.IHighlightLineRange[]
}

interface Props extends Repo {
    commitID: string
    filePath: string
    highlightRanges: HighlightRange[]
    /** The 0-based (inclusive) line number that this code excerpt starts at */
    startLine: number
    /** The 0-based (exclusive) line number that this code excerpt ends at */
    endLine: number
    /** Whether or not this is the first result being shown or not. */
    isFirst: boolean
    className?: string
    /** A function to fetch the range of lines this code excerpt will display. It will be provided
     * the same start and end lines properties that were provided as component props */
    fetchHighlightedFileRangeLines: (isFirst: boolean, startLine: number, endLine: number) => Observable<string[]>
    blobLines?: string[]

    viewerUpdates?: Observable<{ viewerId: ViewerId } & HoverContext>
    hoverifier?: Hoverifier<HoverContext, HoverMerged, ActionItemAction>
}

export interface HighlightRange {
    /**
     * The 0-based line number that this highlight appears in
     */
    line: number
    /**
     * The 0-based character offset to start highlighting at
     */
    character: number
    /**
     * The number of characters to highlight
     */
    highlightLength: number
}

const domFunctions: DOMFunctions = {
    getCodeElementFromTarget: target => {
        const row = target.closest('tr')
        if (!row) {
            return null
        }
        return row.cells[1]
    },
    getCodeElementFromLineNumber: (codeView: HTMLElement, line: number): HTMLTableCellElement | null => {
        const lineElement = codeView.querySelector(`td[data-line="${line}"]`)
        if (!lineElement) {
            return null
        }
        const row = lineElement.closest('tr')
        if (!row) {
            return null
        }
        return row.cells[1]
    },
    getLineNumberFromCodeElement: codeCell => {
        const row = codeCell.closest('tr')
        if (!row) {
            throw new Error('Could not find closest row for codeCell')
        }
        const numberCell = row.cells[0]
        if (!numberCell || !numberCell.dataset.line) {
            throw new Error('Could not find line number')
        }
        return parseInt(numberCell.dataset.line, 10)
    },
}

/**
 * A code excerpt that displays syntax highlighting and match range highlighting.
 */
export const CodeExcerpt: React.FunctionComponent<Props> = ({
    blobLines,
    isFirst,
    startLine,
    endLine,
    fetchHighlightedFileRangeLines,
    hoverifier,
    viewerUpdates,
    highlightRanges,
    className,
}) => {
    const tableContainerElement = useMemo(() => new BehaviorSubject<HTMLElement | null>(null), [])
    const setTableContainerElement = (reference: HTMLElement | null): void => tableContainerElement.next(reference)

    const [isVisible, setIsVisible] = useState(false)
    const visibilitySensorOffset = { bottom: -500 }

    const [highlightedLinesOrError, setHighlightedLinesOrError] = useState<string[] | ErrorLike | undefined>(undefined)

    const hoverContext = useObservable(useMemo(() => viewerUpdates ?? NEVER, [viewerUpdates]))

    useEffect(() => {
        if (hoverContext && hoverifier) {
            const subscription = hoverifier.hoverify({
                positionEvents: tableContainerElement.pipe(
                    filter(isDefined),
                    findPositionsFromEvents({ domFunctions })
                ),
                resolveContext: () => hoverContext,
                dom: domFunctions,
            })

            return () => subscription.unsubscribe()
        }
        return () => {}
    }, [hoverContext, hoverifier, tableContainerElement])

    useEffect(() => {
        if (isVisible) {
            if (blobLines) {
                setHighlightedLinesOrError(blobLines)
            } else {
                const subscription = fetchHighlightedFileRangeLines(isFirst, startLine, endLine)
                    .pipe(catchError(error => [asError(error)]))
                    .subscribe(value => setHighlightedLinesOrError(value))

                return () => subscription.unsubscribe()
            }
        }
        return () => {}
    }, [blobLines, endLine, fetchHighlightedFileRangeLines, highlightedLinesOrError, isFirst, isVisible, startLine])

    useEffect(() => {
        if (tableContainerElement.value) {
            const visibleRows = tableContainerElement.value.querySelectorAll('table tr')
            for (const highlight of highlightRanges) {
                // Select the HTML row in the excerpt that corresponds to the line to be highlighted.
                // highlight.line is the 0-indexed line number in the code file, and this.props.startLine is the 0-indexed
                // line number of the first visible line in the excerpt. So, subtract this.props.startLine
                // from highlight.line to get the correct 0-based index in visibleRows that holds the HTML row.
                const tableRow = visibleRows[highlight.line - startLine]
                if (tableRow) {
                    // Take the lastChild of the row to select the code portion of the table row (each table row consists of the line number and code).
                    const code = tableRow.lastChild as HTMLTableCellElement
                    highlightNode(code, highlight.character, highlight.highlightLength)
                }
            }
        }
    }, [highlightRanges, startLine, tableContainerElement])

    return (
        <VisibilitySensor onChange={setIsVisible} partialVisibility={true} offset={visibilitySensorOffset}>
            <Typography.Code
                data-testid="code-excerpt"
                className={classNames(
                    styles.codeExcerpt,
                    className,
                    isErrorLike(highlightedLinesOrError) && styles.codeExcerptError
                )}
            >
                {highlightedLinesOrError && !isErrorLike(highlightedLinesOrError) && (
                    <div
                        ref={setTableContainerElement}
                        dangerouslySetInnerHTML={{ __html: '<table>' + highlightedLinesOrError.join('') + '</table>' }}
                    />
                )}
                {highlightedLinesOrError && isErrorLike(highlightedLinesOrError) && (
                    <div className={styles.codeExcerptAlert}>
                        <Icon role="img" className="mr-2" as={AlertCircleIcon} aria-hidden={true} />
                        {highlightedLinesOrError.message}
                    </div>
                )}
                {!highlightedLinesOrError && (
                    <table>
                        <tbody>
                            {range(startLine, endLine).map(index => (
                                <tr key={index}>
                                    <td className="line">{index + 1}</td>
                                    {/* create empty space to fill viewport (as if the blob content were already fetched, otherwise we'll overfetch) */}
                                    <td className="code"> </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </Typography.Code>
        </VisibilitySensor>
    )
}
