import React from 'react'

import { MenuLink as ReachMenuLink, type MenuLinkProps as ReachMenuLinkProps } from '@reach/menu-button'
import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../types'

import { MenuDisabledLink } from './MenuDisabledItem'

import styles from './MenuItem.module.scss'

export type MenuLinkProps = ReachMenuLinkProps

/**
 * A styled link component that should be used for any items
 * that will navigate away from the Menu.
 *
 * Renders an `<a>` element by default. Can be modified using the `as` prop.
 *
 * @see — Docs https://reach.tech/menu-button#menulink
 */
export const MenuLink = React.forwardRef(function MenuLink({ className, disabled, ...props }, reference) {
    const Component = disabled ? MenuDisabledLink : ReachMenuLink

    return (
        <Component
            ref={reference}
            {...props}
            // NOTE: In Safari 16.0+, the onBlur event bubbling up to the MenuButton
            // prevents the menu from opening properly. We prevent this event from
            // bubbling up as ReachUI is managing focus state for us internally.
            onBlur={event => event.stopPropagation()}
            className={classNames(styles.dropdownItem, className)}
        />
    )
}) as ForwardReferenceComponent<'a', MenuLinkProps>
