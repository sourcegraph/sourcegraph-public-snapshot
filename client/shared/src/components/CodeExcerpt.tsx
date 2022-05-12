import React from 'react'

import classNames from 'classnames'
import { range, isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import VisibilitySensor from 'react-visibility-sensor'
import { of, combineLatest, Observable, Subject, Subscription, BehaviorSubject, NEVER } from 'rxjs'
import { catchError, filter, switchMap, map, distinctUntilChanged } from 'rxjs/operators'

import { HoverMerged } from '@sourcegraph/client-api'
import { DOMFunctions, findPositionsFromEvents, Hoverifier } from '@sourcegraph/codeintellify'
import { asError, ErrorLike, isDefined, isErrorLike, highlightNode } from '@sourcegraph/common'
import { Icon } from '@sourcegraph/wildcard'

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

interface State {
    blobLinesOrError?: string[] | ErrorLike
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
export class CodeExcerpt extends React.PureComponent<Props, State> {
    public state: State = {}
    private tableContainerElements = new BehaviorSubject<HTMLElement | null>(null)
    private propsChanges = new Subject<Props>()
    private visibilityChanges = new Subject<boolean>()
    private subscriptions = new Subscription()
    private visibilitySensorOffset = { bottom: -500 }

    constructor(props: Props) {
        super(props)
        this.subscriptions.add(
            combineLatest([this.propsChanges, this.visibilityChanges])
                .pipe(
                    filter(([, isVisible]) => isVisible),
                    map(([props]) => props),
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    switchMap(({ blobLines, isFirst, startLine, endLine }) =>
                        blobLines ? of(blobLines) : props.fetchHighlightedFileRangeLines(isFirst, startLine, endLine)
                    ),
                    catchError(error => [asError(error)])
                )
                .subscribe(blobLinesOrError => {
                    this.setState({ blobLinesOrError })
                })
        )

        let hoverifierSubscription: Subscription | null
        this.subscriptions.add(
            combineLatest([
                props.viewerUpdates ?? NEVER,
                this.propsChanges.pipe(
                    map(props => props.hoverifier),
                    distinctUntilChanged(),
                    filter(isDefined)
                ),
            ]).subscribe(([{ viewerId, ...hoverContext }, hoverifier]) => {
                if (hoverifierSubscription) {
                    hoverifierSubscription.unsubscribe()
                    this.subscriptions.remove(hoverifierSubscription)
                }

                hoverifierSubscription = hoverifier.hoverify({
                    positionEvents: this.tableContainerElements.pipe(
                        filter(isDefined),
                        findPositionsFromEvents({ domFunctions })
                    ),
                    resolveContext: () => hoverContext,
                    dom: domFunctions,
                })
                this.subscriptions.add(hoverifierSubscription)
            })
        )
    }

    public componentDidMount(): void {
        this.propsChanges.next(this.props)
    }

    public componentDidUpdate(): void {
        this.propsChanges.next(this.props)

        if (this.tableContainerElements.value) {
            const visibleRows = this.tableContainerElements.value.querySelectorAll('table tr')
            for (const highlight of this.props.highlightRanges) {
                // Select the HTML row in the excerpt that corresponds to the line to be highlighted.
                // highlight.line is the 0-indexed line number in the code file, and this.props.startLine is the 0-indexed
                // line number of the first visible line in the excerpt. So, subtract this.props.startLine
                // from highlight.line to get the correct 0-based index in visibleRows that holds the HTML row.
                const tableRow = visibleRows[highlight.line - this.props.startLine]
                if (tableRow) {
                    // Take the lastChild of the row to select the code portion of the table row (each table row consists of the line number and code).
                    const code = tableRow.lastChild as HTMLTableDataCellElement
                    highlightNode(code, highlight.character, highlight.highlightLength)
                }
            }
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private onChangeVisibility = (isVisible: boolean): void => {
        this.visibilityChanges.next(isVisible)
    }

    public render(): JSX.Element | null {
        return (
            <VisibilitySensor
                onChange={this.onChangeVisibility}
                partialVisibility={true}
                offset={this.visibilitySensorOffset}
            >
                <code
                    data-testid="code-excerpt"
                    className={classNames(
                        styles.codeExcerpt,
                        this.props.className,
                        isErrorLike(this.state.blobLinesOrError) && styles.codeExcerptError
                    )}
                >
                    {this.state.blobLinesOrError && !isErrorLike(this.state.blobLinesOrError) && (
                        <div
                            ref={this.setTableContainerElement}
                            dangerouslySetInnerHTML={{ __html: this.makeTableHTML(this.state.blobLinesOrError) }}
                        />
                    )}
                    {this.state.blobLinesOrError && isErrorLike(this.state.blobLinesOrError) && (
                        <div className={styles.codeExcerptAlert}>
                            <Icon role="img" className="mr-2" as={AlertCircleIcon} aria-hidden={true} />
                            {this.state.blobLinesOrError.message}
                        </div>
                    )}
                    {!this.state.blobLinesOrError && (
                        <table>
                            <tbody>
                                {range(this.props.startLine, this.props.endLine).map(index => (
                                    <tr key={index}>
                                        <td className="line">{index + 1}</td>
                                        {/* create empty space to fill viewport (as if the blob content were already fetched, otherwise we'll overfetch) */}
                                        <td className="code"> </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    )}
                </code>
            </VisibilitySensor>
        )
    }

    private setTableContainerElement = (reference: HTMLElement | null): void => {
        this.tableContainerElements.next(reference)
    }

    private makeTableHTML(blobLines: string[]): string {
        return '<table>' + blobLines.join('') + '</table>'
    }
}
