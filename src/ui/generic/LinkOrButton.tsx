import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Key } from 'ts-key-enum'

type Props = {
    /** A tooltip to display when the user hovers or focuses this element. */
    ['data-tooltip']?: string

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
          target?: '_self' | '_blank'

          onSelect?: never
      })

/**
 * A component that is displayed in the same way, regardless of whether it's a link (with a
 * destination URL) or a button (with a click handler).
 *
 * It is keyboard accessible: unlike <Link> or <a>, pressing the enter key triggers it. Unlike
 * <button>, it shows a focus ring.
 */
export class LinkOrButton extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        const className = `${this.props.className === undefined ? 'nav-link' : this.props.className} ${
            this.props.disabled ? 'disabled' : ''
        }`

        if ('onSelect' in this.props) {
            // Render using an <a> with no href, so that we get a focus ring (when using Bootstrap).
            // We need to set up a keypress listener because <a onclick> doesn't get triggered by
            // enter.
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

        // Render using <Link> (or <a> for external URLs).
        if (typeof this.props.to === 'string' && /^https?:\/\//.test(this.props.to)) {
            return (
                <a
                    href={this.props.to}
                    target={this.props.target}
                    className={className}
                    tabIndex={0}
                    data-tooltip={this.props['data-tooltip']}
                >
                    {this.props.children}
                </a>
            )
        }
        return (
            <Link
                to={this.props.to}
                target={this.props.target}
                className={className}
                tabIndex={0}
                data-tooltip={this.props['data-tooltip']}
            >
                {this.props.children}
            </Link>
        )
    }

    private onAnchorClick: React.MouseEventHandler<HTMLAnchorElement> = e => {
        if (this.props.onSelect) {
            this.props.onSelect()
        }
    }

    private onAnchorKeyPress: React.KeyboardEventHandler<HTMLAnchorElement> = e => {
        if (isSelectKeyPress(e)) {
            if (this.props.onSelect) {
                this.props.onSelect()
            }
        }
    }
}

function isSelectKeyPress(e: React.KeyboardEvent): boolean {
    return e.key === Key.Enter && !e.ctrlKey && !e.shiftKey && !e.metaKey && !e.altKey
}
