import * as H from 'history'
import { noop } from 'lodash'
import React, { useCallback, AnchorHTMLAttributes } from 'react'
import { Key } from 'ts-key-enum'

import { RouterLink, AnchorLink } from '@sourcegraph/wildcard'

import { Button } from '..'
import { ButtonProps } from '../Button'

const isSelectKeyPress = (event: React.KeyboardEvent): boolean =>
    event.key === Key.Enter && !event.ctrlKey && !event.shiftKey && !event.metaKey && !event.altKey

export type ButtonLinkProps = Omit<ButtonProps, 'as'> &
    AnchorHTMLAttributes<HTMLAnchorElement> & {
        /** The link destination URL. */
        to?: H.LocationDescriptor

        /**
         * Called when the user clicks or presses enter on this element.
         */
        onSelect?: (event: React.MouseEvent<HTMLElement> | React.KeyboardEvent<HTMLElement>) => void
    }

/**
 * A component that is displayed in the same way, regardless of whether it's a link (with a
 * destination URL) or a button (with a click handler).
 *
 * It is keyboard accessible: unlike `<Link>` or `<a>`, pressing the enter key triggers it.
 */
export const ButtonLink: React.FunctionComponent<ButtonLinkProps> = React.forwardRef(
    ({ className, to, disabled, onSelect = noop, ...rest }, reference) => {
        // We need to set up a keypress listener because <a onclick> doesn't get
        // triggered by enter.
        const onAnchorKeyPress: React.KeyboardEventHandler<HTMLElement> = useCallback(
            event => {
                if (!disabled && isSelectKeyPress(event)) {
                    onSelect(event)
                }
            },
            [onSelect, disabled]
        )

        const onClickPreventDefault: React.MouseEventHandler<HTMLElement> = useCallback(
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

        const commonProps = {
            disabled,
            onClick: onSelect,
            onKeyPress: onAnchorKeyPress,
            className,
            ref: reference,
        }

        if (!to || disabled) {
            return (
                <Button
                    {...commonProps}
                    as={AnchorLink}
                    to=""
                    onClick={onClickPreventDefault}
                    onAuxClick={onClickPreventDefault}
                    role="button"
                    {...rest}
                />
            )
        }

        return <Button {...commonProps} as={RouterLink} to={to} {...rest} />
    }
)
