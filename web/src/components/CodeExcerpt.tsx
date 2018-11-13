import { range } from 'lodash'
import React from 'react'
import VisibilitySensor from 'react-visibility-sensor'
import { combineLatest, Subject, Subscription } from 'rxjs'
import { filter, switchMap } from 'rxjs/operators'
import { Repo } from '../repo'
import { fetchHighlightedFileLines } from '../repo/backend'
import { highlightNode } from '../util/dom'

interface Props extends Repo {
    commitID: string
    filePath: string
    // How many extra lines to show in the excerpt before/after the ref.
    context?: number
    highlightRanges: HighlightRange[]
    className?: string
    isLightTheme: boolean
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
    blobLines?: string[]
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
                    switchMap(([{ repoPath, filePath, commitID, isLightTheme }]) =>
                        fetchHighlightedFileLines({ repoPath, commitID, filePath, isLightTheme, disableTimeout: true })
                    )
                )
                .subscribe(
                    blobLines => {
                        this.setState({ blobLines })
                    },
                    err => {
                        this.setState({ blobLines: [] })
                        console.error('failed to fetch blob content', err)
                    }
                )
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
                const code = visibleRows[highlight.line - this.getFirstLine()].lastChild as HTMLTableDataCellElement
                highlightNode(code, highlight.character, highlight.highlightLength)
            }
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private getFirstLine(): number {
        return Math.max(0, Math.min(...this.props.highlightRanges.map(r => r.line)) - (this.props.context || 1))
    }

    private getLastLine(): number {
        const lastLine = Math.max(...this.props.highlightRanges.map(r => r.line)) + (this.props.context || 1)
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

        return (
            <VisibilitySensor
                onChange={this.onChangeVisibility}
                partialVisibility={true}
                offset={this.visibilitySensorOffset}
            >
                <code className={`code-excerpt ${this.props.className || ''}`}>
                    {this.state.blobLines && (
                        <div
                            ref={this.setTableContainerElement}
                            dangerouslySetInnerHTML={{ __html: this.makeTableHTML() }}
                        />
                    )}
                    {!this.state.blobLines && (
                        <table>
                            <tbody>
                                {range(this.getFirstLine(), this.getLastLine()).map(i => (
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

    private makeTableHTML(): string {
        return (
            '<table>' + this.state.blobLines!.slice(this.getFirstLine(), this.getLastLine() + 1).join('') + '</table>'
        )
    }
}
