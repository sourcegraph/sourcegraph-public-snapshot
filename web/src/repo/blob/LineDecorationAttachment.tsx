import * as React from 'react'
import ReactDOM from 'react-dom'
import { DecorationAttachmentRenderOptions } from '../../backend/lsp'
import { LinkOrSpan } from '../../components/LinkOrSpan'
import { AbsoluteRepoFile } from '../index'

interface LineDecorationAttachmentProps extends AbsoluteRepoFile {
    line: number
    portalID: string
    attachment: DecorationAttachmentRenderOptions
}

/** Displays text after a line in Blob2. */
export class LineDecorationAttachment extends React.PureComponent<LineDecorationAttachmentProps> {
    private portal: Element | null = null

    public componentWillMount(): void {
        this.portal = document.getElementById(this.props.portalID)
    }

    public componentWillReceiveProps(nextProps: Readonly<LineDecorationAttachmentProps>): void {
        if (nextProps.portalID !== this.props.portalID) {
            this.portal = document.getElementById(nextProps.portalID)
        }
    }

    public render(): React.ReactPortal | null {
        if (!this.portal) {
            return null
        }

        return ReactDOM.createPortal(
            <LinkOrSpan className="line-decoration-attachment" to={this.props.attachment.linkURL}>
                <span
                    className="line-decoration-attachment__contents"
                    // tslint:disable-next-line:jsx-ban-props
                    style={{
                        color: this.props.attachment.color,
                        backgroundColor: this.props.attachment.backgroundColor,
                    }}
                    data-contents={this.props.attachment.contentText || ''}
                />
            </LinkOrSpan>,
            this.portal
        )
    }
}
