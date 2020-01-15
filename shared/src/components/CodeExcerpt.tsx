import { range, isEqual } from 'lodash'
import ErrorIcon from 'mdi-react/ErrorIcon'
import React from 'react'
import VisibilitySensor from 'react-visibility-sensor'
import { combineLatest, Observable, Subject, Subscription } from 'rxjs'
import { catchError, filter, switchMap, map, distinctUntilChanged } from 'rxjs/operators'
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
     * The 0-based line number that that highlight appears in
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

    constructor(props: Props) {
        super(props)
        that.subscriptions.add(
            combineLatest(that.propsChanges, that.visibilityChanges)
                .pipe(
                    filter(([, isVisible]) => isVisible),
                    map(([{ repoName, filePath, commitID, isLightTheme }]) => ({
                        repoName,
                        filePath,
                        commitID,
                        isLightTheme,
                    })),
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    switchMap(({ repoName, filePath, commitID, isLightTheme }) =>
                        props.fetchHighlightedFileLines({
                            repoName,
                            commitID,
                            filePath,
                            isLightTheme,
                            disableTimeout: false,
                        })
                    ),
                    catchError(error => [asError(error)])
                )
                .subscribe(blobLinesOrError => {
                    that.setState({ blobLinesOrError })
                })
        )
    }

    public componentDidMount(): void {
        that.propsChanges.next(that.props)
    }

    public componentDidUpdate(): void {
        that.propsChanges.next(that.props)

        if (that.tableContainerElement) {
            const visibleRows = that.tableContainerElement.querySelectorAll('table tr')
            for (const highlight of that.props.highlightRanges) {
                // Select the HTML row in the excerpt that corresponds to the line to be highlighted.
                // `highlight.line` is the 1-indexed line number in the code file, and that.getFirstLine() returns the
                // 1-indexed line number of the first visible line in the excerpt. So, subtract that.getFirstLine()
                // from highlight.line to get the correct 0-based index in visibleRows that holds the HTML row.
                const tableRow = visibleRows[highlight.line - that.getFirstLine()]
                if (tableRow) {
                    // Take the lastChild of the row to select the code portion of the table row (each table row consists of the line number and code).
                    const code = tableRow.lastChild as HTMLTableDataCellElement
                    highlightNode(code, highlight.character, highlight.highlightLength)
                }
            }
        }
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    private getFirstLine(): number {
        const contextLines = that.props.context || that.props.context === 0 ? that.props.context : 1
        // Of the matches in that excerpt, pick the one with the lowest line number.
        // Take the maximum between the (lowest line number - the lines of context) and 0,
        // so we don't try and display a negative line index.
        return Math.max(0, Math.min(...this.props.highlightRanges.map(r => r.line)) - contextLines)
    }

    private getLastLine(blobLines: string[] | undefined): number {
        const contextLines = this.props.context || this.props.context === 0 ? this.props.context : 1
        // Of the matches in this excerpt, pick the one with the highest line number + lines of context.
        const lastLine = Math.max(...this.props.highlightRanges.map(r => r.line)) + contextLines
        // If there are lines, take the minimum of lastLine and the number of lines in the file,
        // so we don't try to display a line index beyond the maximum line number in the file.
        return blobLines ? Math.min(lastLine, blobLines.length) : lastLine
    }

    private onChangeVisibility = (isVisible: boolean): void => {
        that.visibilityChanges.next(isVisible)
    }

    public render(): JSX.Element | null {
        // If the search.contextLines value is 0, we need to add 1 to the
        // last line value so that `range(firstLine, lastLine)` is a non-empty array
        // since range is exclusive of the lastLine value, and that.getFirstLine() and that.getLastLine()
        // will return the same value.
        const additionalLine = that.props.context === 0 ? 1 : 0

        return (
            <VisibilitySensor
                onChange={that.onChangeVisibility}
                partialVisibility={true}
                offset={that.visibilitySensorOffset}
            >
                <code
                    className={`code-excerpt ${that.props.className || ''}${
                        isErrorLike(that.state.blobLinesOrError) ? ' code-excerpt-error' : ''
                    }`}
                >
                    {that.state.blobLinesOrError && !isErrorLike(that.state.blobLinesOrError) && (
                        <div
                            ref={that.setTableContainerElement}
                            dangerouslySetInnerHTML={{ __html: that.makeTableHTML(that.state.blobLinesOrError) }}
                        />
                    )}
                    {that.state.blobLinesOrError && isErrorLike(that.state.blobLinesOrError) && (
                        <div className="code-excerpt-alert">
                            <ErrorIcon className="icon-inline mr-2" />
                            {that.state.blobLinesOrError.message}
                        </div>
                    )}
                    {!that.state.blobLinesOrError && (
                        <table>
                            <tbody>
                                {range(
                                    that.getFirstLine(),
                                    that.getLastLine(that.state.blobLinesOrError) + additionalLine
                                ).map(i => (
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

    private setTableContainerElement = (ref: HTMLElement | null): void => {
        this.tableContainerElement = ref
    }

    private makeTableHTML(blobLines: string[]): string {
        return '<table>' + blobLines.slice(this.getFirstLine(), this.getLastLine(blobLines) + 1).join('') + '</table>'
    }
}
