import React, { useCallback, useEffect, useLayoutEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import { range } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import VisibilitySensor from 'react-visibility-sensor'
import { of, Observable, Subscription, BehaviorSubject } from 'rxjs'
import { catchError, filter } from 'rxjs/operators'

import { HoverMerged } from '@sourcegraph/client-api'
import { DOMFunctions, findPositionsFromEvents, Hoverifier } from '@sourcegraph/codeintellify'
import { asError, ErrorLike, isDefined, isErrorLike, highlightNode } from '@sourcegraph/common'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import { HoverContext } from '@sourcegraph/shared/src/hover/HoverOverlay.types'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Repo } from '@sourcegraph/shared/src/util/url'
import { Icon, Code } from '@sourcegraph/wildcard'

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

const makeTableHTML = (blobLines: string[]): string => '<table>' + blobLines.join('') + '</table>'
const visibilitySensorOffset = { bottom: -500 }

/**
 * A code excerpt that displays syntax highlighting and match range highlighting.
 */
export const CodeExcerpt: React.FunctionComponent<Props> = ({
    blobLines,
    fetchHighlightedFileRangeLines,
    isFirst,
    startLine,
    endLine,
    highlightRanges,
    viewerUpdates,
    hoverifier,
    className,
}) => {
    const [blobLinesOrError, setBlobLinesOrError] = useState<string[] | ErrorLike | null>(null)
    const [isVisible, setIsVisible] = useState(false)

    // Both the behavior subject and the React state are needed here. The behavior subject is
    // used for hoverified events while the React state is used for match highlighting.
    // The state is needed because React won't re-render when the behavior subject's value changes.
    const tableContainerElements = useMemo(() => new BehaviorSubject<HTMLElement | null>(null), [])
    const [tableContainerElement, setTableContainerElement] = useState<HTMLElement | null>(null)
    const updateTableContainerElementReference = useCallback(
        (reference: HTMLElement | null): void => {
            tableContainerElements.next(reference)
            setTableContainerElement(reference)
        },
        [tableContainerElements]
    )

    // Get the syntax highlighted blob lines
    useEffect(() => {
        let subscription: Subscription | undefined
        if (isVisible) {
            const observable = blobLines ? of(blobLines) : fetchHighlightedFileRangeLines(isFirst, startLine, endLine)
            subscription = observable.pipe(catchError(error => [asError(error)])).subscribe(blobLinesOrError => {
                setBlobLinesOrError(blobLinesOrError)
            })
        }
        return () => subscription?.unsubscribe()
    }, [blobLines, endLine, fetchHighlightedFileRangeLines, isFirst, isVisible, startLine])

    // Highlight the search matches
    useLayoutEffect(() => {
        if (tableContainerElement) {
            const visibleRows = tableContainerElement.querySelectorAll('table tr')
            for (const highlight of highlightRanges) {
                // Select the HTML row in the excerpt that corresponds to the line to be highlighted.
                // highlight.line is the 0-indexed line number in the code file, and startLine is the 0-indexed
                // line number of the first visible line in the excerpt. So, subtract startLine
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

    // Hook up the hover tooltips
    useEffect(() => {
        let hoverifierSubscription: Subscription | null

        const subscription = viewerUpdates?.subscribe(({ viewerId, ...hoverContext }) => {
            if (hoverifier) {
                if (hoverifierSubscription) {
                    hoverifierSubscription.unsubscribe()
                }

                hoverifierSubscription = hoverifier.hoverify({
                    positionEvents: tableContainerElements.pipe(
                        filter(isDefined),
                        findPositionsFromEvents({ domFunctions })
                    ),
                    resolveContext: () => hoverContext,
                    dom: domFunctions,
                })
            }
        })

        return () => {
            subscription?.unsubscribe()
            hoverifierSubscription?.unsubscribe()
        }
    }, [hoverifier, tableContainerElements, viewerUpdates])

    return (
        <VisibilitySensor onChange={setIsVisible} partialVisibility={true} offset={visibilitySensorOffset}>
            <Code
                data-testid="code-excerpt"
                className={classNames(
                    styles.codeExcerpt,
                    className,
                    isErrorLike(blobLinesOrError) && styles.codeExcerptError
                )}
            >
                {blobLinesOrError && !isErrorLike(blobLinesOrError) && (
                    <div
                        ref={updateTableContainerElementReference}
                        dangerouslySetInnerHTML={{ __html: makeTableHTML(blobLinesOrError) }}
                    />
                )}
                {blobLinesOrError && isErrorLike(blobLinesOrError) && (
                    <div className={styles.codeExcerptAlert}>
                        <Icon role="img" className="mr-2" as={AlertCircleIcon} aria-hidden={true} />
                        {blobLinesOrError.message}
                    </div>
                )}
                {!blobLinesOrError && (
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
            </Code>
        </VisibilitySensor>
    )
}
