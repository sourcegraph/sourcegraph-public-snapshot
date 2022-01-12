import classNames from 'classnames'
import * as H from 'history'
import { noop } from 'lodash'
import React, { useCallback, AnchorHTMLAttributes } from 'react'
import { Key } from 'ts-key-enum'

import { isDefined } from '@sourcegraph/common'

import { Link } from './Link'

const isSelectKeyPress = (event: React.KeyboardEvent): boolean =>
    event.key === Key.Enter && !event.ctrlKey && !event.shiftKey && !event.metaKey && !event.altKey

export interface ButtonLinkProps extends Pick<AnchorHTMLAttributes<never>, 'target' | 'rel'> {
    /** The link destination URL. */
    to?: H.LocationDescriptor

    /**
     * Called when the user clicks or presses enter on this element.
     */
    onSelect?: (event: React.MouseEvent<HTMLElement> | React.KeyboardEvent<HTMLElement>) => void

    /** A tooltip to display when the user hovers or focuses this element. */
    ['data-tooltip']?: string

    /**
     * If given, the element is treated as a toggle with the boolean indicating its state.
     * Applies `aria-pressed`.
     */
    pressed?: boolean

    /**
     * The component's CSS class name
     *
     * @default "nav-link"
     */
    className?: string

    disabled?: boolean

    disabledClassName?: string

    id?: string

    buttonLinkRef?: React.Ref<HTMLAnchorElement>

    ['data-content']?: string

    /** Override default tab index */
    tabIndex?: number
}

/**
 * A component that is displayed in the same way, regardless of whether it's a link (with a
 * destination URL) or a button (with a click handler).
 *
 * It is keyboard accessible: unlike `<Link>` or `<a>`, pressing the enter key triggers it.
 */
export const ButtonLink: React.FunctionComponent<ButtonLinkProps> = ({
    className = 'nav-link',
    to,
    target,
    rel,
    disabled,
    disabledClassName,
    pressed,
    'data-tooltip': tooltip,
    onSelect = noop,
    children,
    id,
    buttonLinkRef: buttonLinkReference = null,
    'data-content': dataContent,
    tabIndex,
}) => {
    // We need to set up a keypress listener because <a onclick> doesn't get
    // triggered by enter.
    const onAnchorKeyPress: React.KeyboardEventHandler<HTMLAnchorElement> = useCallback(
        event => {
            if (!disabled && isSelectKeyPress(event)) {
                onSelect(event)
            }
        },
        [onSelect, disabled]
    )

    const commonProps: React.AnchorHTMLAttributes<HTMLAnchorElement> & {
        'data-tooltip': string | undefined
    } = {
        // `.disabled` will only be selected if the `.btn` class is applied as well
        className: classNames(className, disabled && ['disabled', disabledClassName]),
        'data-tooltip': tooltip,
        'aria-label': tooltip,
        role: typeof pressed === 'boolean' ? 'button' : undefined,
        'aria-pressed': pressed,
        tabIndex: isDefined(tabIndex) ? tabIndex : disabled ? -1 : 0,
        onClick: onSelect,
        onKeyPress: onAnchorKeyPress,
        id,
    }

    const onClickPreventDefault: React.MouseEventHandler<HTMLAnchorElement> = useCallback(
        event => {
            // Prevent default action of reloading page because of empty href
            event.preventDefault()

            if (disabled) {
                return false
            }

            return onSelect(event)
        },
        [onSelect, disabled]
    )

    if (!to || disabled) {
        return (
            // Need empty href for styling reasons
            // Use onAuxClick so that middle-clicks are caught.
            // Ideally this should a <button> but we can't guarantee we have the .btn-link class here.
            // eslint-disable-next-line jsx-a11y/anchor-is-valid
            <a
                href=""
                {...commonProps}
                onClick={onClickPreventDefault}
                onAuxClick={onClickPreventDefault}
                role="button"
                ref={buttonLinkReference}
                data-content={dataContent}
            >
                {children}
            </a>
        )
    }

    return (
        <Link {...commonProps} to={to} target={target} rel={rel} ref={buttonLinkReference} data-content={dataContent}>
            {children}
        </Link>
    )
}
