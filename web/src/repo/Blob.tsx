
import * as H from 'history'
import * as React from 'react'
import { addAnnotations } from 'sourcegraph/tooltips'
import { getCodeCellsForAnnotation, getPathExtension, highlightAndScrollToLine, highlightLine, supportedExtensions } from 'sourcegraph/util'
import * as url from 'sourcegraph/util/url'

interface BlobProps {
    html: string
    repoPath
    filePath: string
    commitID: string
    rev?: string
    location: H.Location
    history: H.History
}

export class Blob extends React.Component<BlobProps, {}> {
    private ref: any

    public shouldComponentUpdate(nextProps: BlobProps): boolean {
        const shouldUpdate = this.props.html !== nextProps.html
        // unconditionally apply annotations
        this.applyAnnotations(nextProps)
        return shouldUpdate
    }

    public render(): JSX.Element | null {
        return (
            <div className='blob' onClick={this.handleBlobClick} ref={ref => {
                if (!this.ref && ref) {
                    // first mount, do initial scroll
                    this.scrollToLine(this.props)
                }
                this.ref = ref
                this.applyAnnotations(this.props)
            }} dangerouslySetInnerHTML={{ __html: this.props.html }} />
        )
    }

    private handleBlobClick: React.MouseEventHandler<HTMLDivElement> = e => {
        const target = e.target!
        const row: HTMLTableRowElement = (target as any).closest('tr')
        if (!row) {
            return
        }
        const line = parseInt(row.firstElementChild!.getAttribute('data-line')!, 10)
        highlightLine(this.props.history, this.props.repoPath, this.props.commitID, this.props.filePath!, line, getCodeCellsForAnnotation(), true)
    }

    private scrollToLine = (props: BlobProps) => {
        const line = url.parseHash(props.location.hash).line
        if (line) {
            highlightAndScrollToLine(props.history, props.repoPath,
                props.commitID, props.filePath!, line, getCodeCellsForAnnotation(), false)
        }
    }

    private applyAnnotations = (props: BlobProps) => {
        const cells = getCodeCellsForAnnotation()
        if (supportedExtensions.has(getPathExtension(props.filePath))) {
            addAnnotations(props.history, props.filePath!,
                { repoURI: props.repoPath!, rev: props.rev!, commitID: props.commitID }, cells)
        }
    }
}
