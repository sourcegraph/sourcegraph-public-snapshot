import * as React from 'react'
import ReactDOM from 'react-dom'
import { DecorationAttachmentRenderOptions } from 'sourcegraph'
import { decorationAttachmentStyleForTheme } from '../../../../shared/src/api/client/services/decoration'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import { AbsoluteRepoFile } from '../../../../shared/src/util/url'
import { ThemeProps } from '../../theme'

interface LineDecorationAttachmentProps extends AbsoluteRepoFile, ThemeProps {
    line: number
    portalID: string
    attachment: DecorationAttachmentRenderOptions
}

/** Displays text after a line in Blob. */
export class LineDecorationAttachment extends React.PureComponent<LineDecorationAttachmentProps> {
    private portal: Element | null = null

    public componentDidMount(): void {
        this.portal = document.getElementById(this.props.portalID)
    }

    public componentDidUpdate(prevProps: Readonly<LineDecorationAttachmentProps>): void {
        if (
            this.props.repoName !== prevProps.repoName ||
            this.props.rev !== prevProps.rev ||
            this.props.filePath !== prevProps.filePath ||
            this.props.line !== prevProps.line ||
            this.props.portalID !== prevProps.portalID ||
            this.props.attachment !== prevProps.attachment
        ) {
            this.portal = document.getElementById(this.props.portalID)
        }
    }

    public render(): React.ReactPortal | null {
        if (!this.portal) {
            return null
        }

        const style = decorationAttachmentStyleForTheme(this.props.attachment, this.props.isLightTheme)

        return ReactDOM.createPortal(
            <LinkOrSpan
                className="line-decoration-attachment"
                to={this.props.attachment.linkURL}
                data-tooltip={this.props.attachment.hoverMessage}
                // Use target to open external URLs (or else react-router's Link will treat the URL as a URL path
                // and navigation will fail).
                target={
                    this.props.attachment.linkURL && /^https?:\/\//.test(this.props.attachment.linkURL)
                        ? '_blank'
                        : undefined
                }
                // Avoid leaking referrer URLs (which contain repository and path names, etc.) to external sites.
                rel="noreferrer noopener"
            >
                <span
                    className="line-decoration-attachment__contents"
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{
                        color: style.color,
                        backgroundColor: style.backgroundColor,
                    }}
                    data-contents={this.props.attachment.contentText || ''}
                />
            </LinkOrSpan>,
            this.portal
        )
    }
}
