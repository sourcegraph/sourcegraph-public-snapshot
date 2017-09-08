import * as React from 'react'
import * as VisibilitySensor from 'react-visibility-sensor'
import { fetchHighlightedFileLines } from 'sourcegraph/repo/backend'
import { highlightNode } from 'sourcegraph/util/dom'
import { BlobPosition } from 'sourcegraph/util/types'

interface Props extends BlobPosition {
    // How many extra lines to show in the excerpt before/after the ref.
    previewWindowExtraLines?: number
    highlightLength: number
}

interface State {
    blobLines?: string[]
}

export class CodeExcerpt extends React.Component<Props, State> {
    private tableContainerElement: HTMLElement | null

    constructor(props: Props) {
        super(props)
        this.state = {}
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (this.props.highlightLength !== nextProps.highlightLength) {
            // Redraw the component so the matched range highlighting is updated
            this.setState({ blobLines: undefined })
        }
        this.fetchContents(nextProps)
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        if (this.tableContainerElement) {
            const rows = this.tableContainerElement.querySelectorAll('table tr')
            for (const row of rows) {
                const line = row.firstChild as HTMLTableDataCellElement
                const code = row.lastChild as HTMLTableDataCellElement
                if (line.getAttribute('data-line') === '' + (this.props.line + 1)) {
                    highlightNode(code, this.props.char!, this.props.highlightLength)
                }
            }
        }
    }

    public getPreviewWindowLines(): number[] {
        const targetLine = this.props.line
        let res = [targetLine]
        for (let i = targetLine - this.props.previewWindowExtraLines!; i < targetLine + this.props.previewWindowExtraLines! + 1; ++i) {
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

    public onChangeVisibility(isVisible: boolean): void {
        if (isVisible) {
            this.fetchContents(this.props)
        }
    }

    public render(): JSX.Element | null {
        return (
            <VisibilitySensor onChange={isVisible => this.onChangeVisibility(isVisible)} partialVisibility={true}>
                <div className='code-excerpt'>
                    {
                        this.state.blobLines &&
                        <div ref={this.setTableContainerElement} dangerouslySetInnerHTML={{ __html: this.makeTableHTML() }} />
                    }
                    {
                        !this.state.blobLines &&
                        <table >
                            <tbody>
                                {
                                    this.getPreviewWindowLines().map(i =>
                                        <tr key={i}>
                                            <td className='line'>{i + 1}</td>
                                            {/* create empty space to fill viewport (as if the blob content were already fetched, otherwise we'll overfetch) */}
                                            <td className='code'> </td>
                                        </tr>
                                    )
                                }
                            </tbody>
                        </table>
                    }
                </div>
            </VisibilitySensor>
        )
    }

    private setTableContainerElement = (ref: HTMLElement | null) => {
        this.tableContainerElement = ref
    }

    private fetchContents(props: Props): void {
        fetchHighlightedFileLines({
            repoPath: props.uri,
            commitID: props.rev,
            filePath: props.path,
            disableTimeout: true
        })
            .then(lines => this.setState({ blobLines: lines }))
            .catch(err => {
                console.error('failed to fetch blob content', err)
            })
    }

    private makeTableHTML(): string {
        const start = Math.max(0, this.props.line - (this.props.previewWindowExtraLines || 0))
        const end = this.props.line + (this.props.previewWindowExtraLines || 0) + 1
        const lineRange = this.state.blobLines!.slice(start, end)
        return '<table>' + lineRange.join('') + '</table>'
    }
}
