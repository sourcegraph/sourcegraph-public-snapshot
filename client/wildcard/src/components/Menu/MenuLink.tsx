import React from 'react'

import { MenuLink as ReachMenuLink, MenuLinkProps as ReachMenuLinkProps } from '@reach/menu-button'
import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../types'

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
export const MenuLink = React.forwardRef(({ className, disabled, ...props }, reference) => {
    const Component = disabled ? MenuDisabledLink : ReachMenuLink

    return <Component ref={reference} {...props} className={classNames(styles.dropdownItem, className)} />
}) as ForwardReferenceComponent<'a', MenuLinkProps>
