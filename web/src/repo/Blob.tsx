
import * as React from 'react'

interface BlobProps {
    html: string
    onClick: React.MouseEventHandler<HTMLDivElement>
    applyAnnotations: () => void
    scrollToLine: () => void
}

export class Blob extends React.Component<BlobProps, {}> {
    private ref: any

    public shouldComponentUpdate(nextProps: BlobProps): boolean {
        return this.props.html !== nextProps.html
    }

    public render(): JSX.Element | null {
        return (
            <div className='blob' onClick={this.props.onClick} ref={ref => {
                if (!this.ref && ref) {
                    // first mount, do initial scroll
                    this.props.scrollToLine()
                }
                this.ref = ref
                this.props.applyAnnotations()
            }} dangerouslySetInnerHTML={{ __html: this.props.html }} />
        )
    }
}
