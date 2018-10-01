import copy from 'copy-to-clipboard'
import * as H from 'history'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import * as React from 'react'
import { Tooltip } from '../../components/tooltip/Tooltip'
import { eventLogger } from '../../tracking/eventLogger'

interface Props {
    location: H.Location
}

interface State {
    copied: boolean
}

/**
 * A repository header action that copies the current page's URL to the clipboard.
 */
export class CopyLinkAction extends React.PureComponent<Props, State> {
    public state = { copied: false }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        if (prevState.copied !== this.state.copied) {
            Tooltip.forceUpdate()
        }
    }

    public render(): JSX.Element | null {
        return (
            <button
                className="copy-link-action btn btn-link btn-link-sm"
                data-tooltip={this.state.copied ? 'Copied!' : 'Copy link to clipboard'}
                onClick={this.onClick}
            >
                <ContentCopyIcon className="icon-inline" />
            </button>
        )
    }

    private onClick: React.MouseEventHandler<HTMLElement> = event => {
        event.preventDefault()
        eventLogger.log('ShareButtonClicked')
        const loc = this.props.location
        const shareLink = new URL(loc.pathname + loc.search + loc.hash, window.location.href)
        shareLink.searchParams.set('utm_source', 'share')
        copy(shareLink.href)
        this.setState({ copied: true })

        setTimeout(() => {
            this.setState({ copied: false })
        }, 1000)
    }
}
