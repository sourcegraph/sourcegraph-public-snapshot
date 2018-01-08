import CopyIcon from '@sourcegraph/icons/lib/Copy'
import copy from 'copy-to-clipboard'
import * as H from 'history'
import * as React from 'react'
import { eventLogger } from '../../tracking/eventLogger'

/**
 * A repository header action that copies the current page's URL to the clipboard.
 */
export class CopyLinkAction extends React.PureComponent<{ location: H.Location }, { copied: boolean }> {
    public state = { copied: false }

    public render(): JSX.Element | null {
        return (
            <button
                className="btn btn-link btn-link-sm composite-container__header-action"
                title="Copy link to clipboard"
                onClick={this.onClick}
            >
                <CopyIcon className="icon-inline" />
                <span className="composite-container__header-action-text">{this.state.copied ? 'Copied!' : ''}</span>
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
