import React, { useCallback, useEffect, useState } from 'react'

import classNames from 'classnames'
import { range, isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import VisibilitySensor from 'react-visibility-sensor'
import { of, combineLatest, Observable, Subscription, BehaviorSubject, NEVER } from 'rxjs'
import { catchError, filter, switchMap, map, distinctUntilChanged } from 'rxjs/operators'

import { HoverMerged } from '@sourcegraph/client-api'
import { DOMFunctions, findPositionsFromEvents, Hoverifier } from '@sourcegraph/codeintellify'
import { asError, ErrorLike, isDefined, isErrorLike, highlightNode } from '@sourcegraph/common'
import { Icon, Typography } from '@sourcegraph/wildcard'

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

const makeTableHTML = (blobLines: string[]): string => '<table>' + blobLines.join('') + '</table>'

/**
 * A code excerpt that displays syntax highlighting and match range highlighting.
 */
export const CodeExcerpt: React.FunctionComponent<Props> = (props: Props) => {
    const [blobLinesOrError, setBlobLinesOrError] = useState<string[] | ErrorLike | null>(null)
    const [tableContainerElements] = useState(new BehaviorSubject<HTMLElement | null>(null))
    const [propsChanges] = useState(new BehaviorSubject<Props>(props))
    const [visibilityChanges] = useState(new BehaviorSubject<boolean | null>(null))
    const visibilitySensorOffset = { bottom: -500 }

    useEffect(() => {
        propsChanges.next(props)
    }, [props, propsChanges])

    // Get the syntax highlighted blob lines
    useEffect(() => {
        const subscription = combineLatest([propsChanges, visibilityChanges])
            .pipe(
                filter(([, isVisible]) => isVisible === true),
                map(([props]) => props),
                distinctUntilChanged((a, b) => isEqual(a, b)),
                switchMap(({ blobLines, isFirst, startLine, endLine }) => {
                    if (blobLines) {
                        return of(blobLines)
                    }
                    return props.fetchHighlightedFileRangeLines(isFirst, startLine, endLine)
                }),
                catchError(error => [asError(error)])
            )
            .subscribe(blobLinesOrError => {
                setBlobLinesOrError(blobLinesOrError)
            })
        return () => subscription.unsubscribe()
    }, [props, propsChanges, visibilityChanges])

    // Highlight the search matches
    useEffect(() => {
        const subscription = tableContainerElements.subscribe(tableContainerElement => {
            if (tableContainerElement) {
                const visibleRows = tableContainerElement.querySelectorAll('table tr')
                for (const highlight of props.highlightRanges) {
                    // Select the HTML row in the excerpt that corresponds to the line to be highlighted.
                    // highlight.line is the 0-indexed line number in the code file, and this.props.startLine is the 0-indexed
                    // line number of the first visible line in the excerpt. So, subtract this.props.startLine
                    // from highlight.line to get the correct 0-based index in visibleRows that holds the HTML row.
                    const tableRow = visibleRows[highlight.line - props.startLine]
                    if (tableRow) {
                        // Take the lastChild of the row to select the code portion of the table row (each table row consists of the line number and code).
                        const code = tableRow.lastChild as HTMLTableCellElement
                        highlightNode(code, highlight.character, highlight.highlightLength)
                    }
                }
            }
        })
        return () => subscription.unsubscribe()
    }, [props.highlightRanges, props.startLine, tableContainerElements])

    // Hook up the hover tooltips
    useEffect(() => {
        let hoverifierSubscription: Subscription | null
        const subscription = combineLatest([
            props.viewerUpdates ?? NEVER,
            propsChanges.pipe(
                map(props => props.hoverifier),
                distinctUntilChanged(),
                filter(isDefined)
            ),
        ]).subscribe(([{ viewerId, ...hoverContext }, hoverifier]) => {
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
        })
        return () => {
            subscription.unsubscribe()
            hoverifierSubscription?.unsubscribe()
        }
    }, [props.viewerUpdates, propsChanges, tableContainerElements])

    const onChangeVisibility = useCallback(
        (isVisible: boolean): void => {
            visibilityChanges.next(isVisible)
        },
        [visibilityChanges]
    )

    const setTableContainerElement = useCallback(
        (reference: HTMLElement | null): void => {
            tableContainerElements.next(reference)
        },
        [tableContainerElements]
    )

    return (
        <VisibilitySensor onChange={onChangeVisibility} partialVisibility={true} offset={visibilitySensorOffset}>
            <Typography.Code
                data-testid="code-excerpt"
                className={classNames(
                    styles.codeExcerpt,
                    props.className,
                    isErrorLike(blobLinesOrError) && styles.codeExcerptError
                )}
            >
                {blobLinesOrError && !isErrorLike(blobLinesOrError) && (
                    <div
                        ref={setTableContainerElement}
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
                            {range(props.startLine, props.endLine).map(index => (
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
