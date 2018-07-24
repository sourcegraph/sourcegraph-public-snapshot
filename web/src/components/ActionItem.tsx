import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Key } from 'ts-key-enum'
import { eventLogger } from '../tracking/eventLogger'

type Props = {
    /** A tooltip to display when the user hovers or focuses this element. */
    ['data-tooltip']?: string

    /** The telemetry event to log; i.e., eventLogger.log(logEvent) is called. */
    logEvent?: string

    /** The component's CSS class name (defaults to "nav-link"). */
    className?: string

    disabled?: boolean
} & (
    | {
          /** For non-links, called when the user clicks or presses enter on this element. */
          onSelect: () => void

          to?: never
          target?: never
      }
    | {
          /** For links, the link destination URL. */
          to: H.LocationDescriptor

          /** The link target (use "_self" for external URLs). */
          target?: '_self'

          onSelect?: never
      })

/**
 * A button with an icon and optional text label displayed in a navbar.
 *
 * It is keyboard accessible: unlike <Link> or <a>, pressing the enter key triggers it. Unlike <button>, it shows a
 * focus ring.
 */
export class ActionItem extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        const className = `${this.props.className === undefined ? 'nav-link' : this.props.className} ${
            this.props.disabled ? 'disabled' : ''
        }`

        if ('onSelect' in this.props) {
            // Render using an <a> with no href, so that we get a focus ring. We need to set up a keypress listener
            // because <a onclick> doesn't get triggered by enter.
            return (
                <a
                    className={className}
                    tabIndex={0}
                    data-tooltip={this.props['data-tooltip']}
                    onClick={this.onAnchorClick}
                    onKeyPress={this.onAnchorKeyPress}
                >
                    {this.props.children}
                </a>
            )
        }

        // Render using Link.
        return (
            <Link
                to={this.props.to}
                target={this.props.target}
                className={className}
                tabIndex={0}
                data-tooltip={this.props['data-tooltip']}
                onClick={this.logEvent}
            >
                {this.props.children}
            </Link>
        )
    }

    private onAnchorClick: React.MouseEventHandler<HTMLAnchorElement> = e => {
        this.logEvent()
        if (this.props.onSelect) {
            this.props.onSelect()
        }
    }

    private onAnchorKeyPress: React.KeyboardEventHandler<HTMLAnchorElement> = e => {
        if (isSelectKeyPress(e)) {
            this.logEvent()
            if (this.props.onSelect) {
                this.props.onSelect()
            }
        }
    }

    private logEvent = () => {
        if (this.props.logEvent) {
            eventLogger.log(this.props.logEvent)
        }
    }
}

function isSelectKeyPress(e: React.KeyboardEvent): boolean {
    return e.key === Key.Enter && !e.ctrlKey && !e.shiftKey && !e.metaKey && !e.altKey
}
