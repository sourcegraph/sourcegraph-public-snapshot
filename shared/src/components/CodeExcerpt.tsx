import { range } from 'lodash'
import React from 'react'
import VisibilitySensor from 'react-visibility-sensor'
import { combineLatest, Observable, Subject, Subscription } from 'rxjs'
import { catchError, filter, switchMap } from 'rxjs/operators'
import { highlightNode } from '../util/dom'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { Repo } from '../util/url'

export interface FetchFileCtx {
    repoName: string
    commitID: string
    filePath: string
    disableTimeout?: boolean
    isLightTheme: boolean
}

interface Props extends Repo {
    commitID: string
    filePath: string
    // How many extra lines to show in the excerpt before/after the ref.
    context?: number
    highlightRanges: HighlightRange[]
    className?: string
    isLightTheme: boolean
    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
}

interface HighlightRange {
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

export class CodeExcerpt extends React.PureComponent<Props, State> {
    public state: State = {}
    private tableContainerElement: HTMLElement | null = null
    private propsChanges = new Subject<Props>()
    private visibilityChanges = new Subject<boolean>()
    private subscriptions = new Subscription()
    private visibilitySensorOffset = { bottom: -500 }

    public constructor(props: Props) {
        super(props)
        this.subscriptions.add(
            combineLatest(this.propsChanges, this.visibilityChanges)
                .pipe(
                    filter(([, isVisible]) => isVisible),
                    switchMap(([{ repoName, filePath, commitID, isLightTheme }]) =>
                        props.fetchHighlightedFileLines({
                            repoName,
                            commitID,
                            filePath,
                            isLightTheme,
                            disableTimeout: true,
                        })
                    ),
                    catchError(error => [asError(error)])
                )
                .subscribe(blobLinesOrError => {
                    this.setState({ blobLinesOrError })
                })
        )
    }

    public componentDidMount(): void {
        this.propsChanges.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.propsChanges.next(nextProps)
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        if (this.tableContainerElement) {
            const visibleRows = this.tableContainerElement.querySelectorAll('table tr')
            for (const highlight of this.props.highlightRanges) {
                // Select the HTML row in the excerpt that corresponds to the line to be highlighted.
                // `highlight.line` is the 1-indexed line number in the code file, and this.getFirstLine() returns the
                // 1-indexed line number of the first visible line in the excerpt. So, subtract this.getFirstLine()
                // from highlight.line to get the correct 0-based index in visibleRows that holds the HTML row.
                const tableRow = visibleRows[highlight.line - this.getFirstLine()]
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

    private getFirstLine(): number {
        const contextLines = this.props.context || this.props.context === 0 ? this.props.context : 1
        // Of the matches in this excerpt, pick the one with the lowest line number.
        // Take the maximum between the (lowest line number - the lines of context) and 0,
        // so we don't try and display a negative line index.
        return Math.max(0, Math.min(...this.props.highlightRanges.map(r => r.line)) - contextLines)
    }

    private getLastLine(): number {
        const contextLines = this.props.context || this.props.context === 0 ? this.props.context : 1
        // Of the matches in this excerpt, pick the one with the highest line number + lines of context.
        const lastLine = Math.max(...this.props.highlightRanges.map(r => r.line)) + contextLines
        // If there are lines, take the minimum of lastLine and the number of lines in the file,
        // so we don't try to display a line index beyond the maximum line number in the file.
        return this.state.blobLines ? Math.min(lastLine, this.state.blobLines.length) : lastLine
    }

    private onChangeVisibility = (isVisible: boolean): void => {
        this.visibilityChanges.next(isVisible)
    }

    public render(): JSX.Element | null {
        if (this.state.blobLines && this.state.blobLines.length === 0) {
            // Show in case of error (e.g., repo not added). This at least lets the user click through, at
            // which point they'll see the full error reason (this is better than showing 3 empty lines of an
            // excerpt).
            return null
        }

        // If the search.contextLines value is 0, we need to add 1 to the
        // last line value so that `range(firstLine, lastLine)` is a non-empty array
        // since range is exclusive of the lastLine value, and this.getFirstLine() and this.getLastLine()
        // will return the same value.
        const additionalLine = this.props.context === 0 ? 1 : 0

        return (
            <VisibilitySensor
                onChange={this.onChangeVisibility}
                partialVisibility={true}
                offset={this.visibilitySensorOffset}
            >
                <code
                    className={`code-excerpt ${this.props.className || ''}${
                        isErrorLike(this.state.blobLinesOrError) ? ' code-excerpt-error' : ''
                    }`}
                >
                    {this.state.blobLinesOrError && !isErrorLike(this.state.blobLinesOrError) && (
                        <div
                            ref={this.setTableContainerElement}
                            dangerouslySetInnerHTML={{ __html: this.makeTableHTML(this.state.blobLinesOrError) }}
                        />
                    )}
                    {this.state.blobLinesOrError && isErrorLike(this.state.blobLinesOrError) && (
                        <div className="alert alert-danger">{this.state.blobLinesOrError.message}</div>
                    )}
                    {!this.state.blobLinesOrError && (
                        <table>
                            <tbody>
                                {range(this.getFirstLine(), this.getLastLine() + additionalLine).map(i => (
                                    <tr key={i}>
                                        <td className="line">{i + 1}</td>
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

    private setTableContainerElement = (ref: HTMLElement | null) => {
        this.tableContainerElement = ref
    }

    private makeTableHTML(blobLines: string[]): string {
        return '<table>' + blobLines.slice(this.getFirstLine(), this.getLastLine(blobLines) + 1).join('') + '</table>'
    }
}
