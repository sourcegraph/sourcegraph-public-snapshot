import React, { type AnchorHTMLAttributes, useRef } from 'react'

import classNames from 'classnames'
import { Key } from 'ts-key-enum'
import { useMergeRefs } from 'use-callback-ref'

import { isDefined } from '@sourcegraph/common'

import type { ForwardReferenceComponent } from '../../types'
import { Button, type ButtonProps } from '../Button'
import { Link, AnchorLink } from '../Link'

const isSelectKeyPress = (event: React.KeyboardEvent): boolean =>
    (event.key === Key.Enter || event.key === ' ') &&
    !event.ctrlKey &&
    !event.shiftKey &&
    !event.metaKey &&
    !event.altKey

export type ButtonLinkProps = Omit<ButtonProps, 'as' | 'onSelect'> &
    Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'onSelect'> & {
        /** The link destination URL. */
        to?: string

        /**
         * Called when the user clicks or presses enter on this element.
         */
        onSelect?: (event: React.MouseEvent<HTMLElement> | React.KeyboardEvent<HTMLElement>) => void

        /**
         * If given, the element is treated as a toggle with the boolean indicating its state.
         * Applies `aria-pressed`.
         */
        pressed?: boolean

        disabledClassName?: string

        ['data-content']?: string

        id?: string

        disabled?: boolean

        /** Override default tab index */
        tabIndex?: number
    }

/**
 * A component that is displayed in the same way, regardless of whether it's a link (with a
 * destination URL) or a button (with a click handler).
 *
 * It is keyboard accessible: unlike `<Link>` or `<a>`, pressing the enter key triggers it.
 */
export const ButtonLink = React.forwardRef(function ButtonLink(props, reference) {
    const {
        className,
        to,
        disabled,
        disabledClassName,
        pressed,
        onSelect,
        children,
        id,
        'data-content': dataContent,
        tabIndex,
        ...rest
    } = props
    const buttonReference = useRef<HTMLAnchorElement>(null)
    const mergedbuttonReference = useMergeRefs([buttonReference, reference])

    // We need to set up a keydown listener because <a onclick> doesn't get
    // triggered by enter.
    const handleKeyDown = (event: React.KeyboardEvent<HTMLElement>): void => {
        if (disabled || !isSelectKeyPress(event)) {
            return
        }

        // Override the original key press and trigger the click event
        event.preventDefault()
        buttonReference.current?.click()
    }

    const handleClick = (event: React.MouseEvent<HTMLElement>): void => {
        // Prevent default action of reloading page because of empty href
        event.preventDefault()

        if (disabled) {
            return
        }

        onSelect?.(event)
    }

    const commonProps = {
        // `.disabled` will only be selected if the `.btn` class is applied as well
        className: classNames(className, disabled && ['disabled', disabledClassName]),
        role: typeof pressed === 'boolean' ? 'button' : undefined,
        'aria-pressed': pressed,
        tabIndex: isDefined(tabIndex) ? tabIndex : disabled ? -1 : 0,
        onClick: onSelect,
        // We need `onKeyDown` instead `onKeyPress` to make ButtonLink works inside reactstrap Dropdown
        onKeyDown: handleKeyDown,
        id,
        ref: mergedbuttonReference,
        disabled,
    }

    if (!to || disabled) {
        return (
            <Button
                {...commonProps}
                as={AnchorLink}
                to=""
                onClick={handleClick}
                onAuxClick={handleClick}
                role="button"
                {...rest}
            >
                {children}
            </Button>
        )
    }

    return (
        <Button {...commonProps} as={Link} to={to} {...rest}>
            {children}
        </Button>
    )
}) as ForwardReferenceComponent<typeof AnchorLink, ButtonLinkProps>
