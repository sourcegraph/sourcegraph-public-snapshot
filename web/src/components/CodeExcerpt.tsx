import React from 'react'
import VisibilitySensor from 'react-visibility-sensor'
import { combineLatest } from 'rxjs/observable/combineLatest'
import { filter } from 'rxjs/operators/filter'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { AbsoluteRepoFile } from '../repo'
import { fetchHighlightedFileLines } from '../repo/backend'
import { highlightNode } from '../util/dom'

interface Props extends AbsoluteRepoFile {
    // How many extra lines to show in the excerpt before/after the ref.
    previewWindowExtraLines?: number
    line: number
    highlightRanges: HighlightRange[]
    className?: string
    isLightTheme: boolean
}

interface HighlightRange {
    /**
     * The 0-based character offset to start highlighting at
     */
    start: number
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
            const rows = this.tableContainerElement.querySelectorAll('table tr')
            for (const row of rows) {
                const line = row.firstChild as HTMLTableDataCellElement
                const code = row.lastChild as HTMLTableDataCellElement
                if (line.getAttribute('data-line') === '' + (this.props.line + 1)) {
                    for (const range of this.props.highlightRanges) {
                        highlightNode(code, range.start, range.highlightLength)
                    }
                }
            }
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private getPreviewWindowLines(): number[] {
        const targetLine = this.props.line
        let res = [targetLine]
        for (
            let i = targetLine - this.props.previewWindowExtraLines!;
            i < targetLine + this.props.previewWindowExtraLines! + 1;
            ++i
        ) {
            if (i > 0 && i < targetLine) {
                res = [i].concat(res)
            }
            if (this.state.blobLines) {
                if (i < this.state.blobLines!.length && i > targetLine) {
                    res = res.concat([i])
                }
            } else {
                if (i > targetLine) {
                    res = res.concat([i])
                }
            }
        }
        return res
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
            <VisibilitySensor onChange={this.onChangeVisibility} partialVisibility={true}>
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
                                {this.getPreviewWindowLines().map(i => (
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
        const start = Math.max(0, this.props.line - (this.props.previewWindowExtraLines || 0))
        const end = this.props.line + (this.props.previewWindowExtraLines || 0) + 1
        const lineRange = this.state.blobLines!.slice(start, end)
        return '<table>' + lineRange.join('') + '</table>'
    }
}
