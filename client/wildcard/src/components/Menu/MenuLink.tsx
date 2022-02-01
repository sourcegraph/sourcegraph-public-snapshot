import { MenuLink as ReachMenuLink, MenuLinkProps as ReachMenuLinkProps } from '@reach/menu-button'
import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

export type MenuLinkProps = ReachMenuLinkProps

/**
 * A styled link component that should be used for any items
 * that will navigate away from the Menu.
 *
 * Renders an `<a>` element by default. Can be modified using the `as` prop.
 *
 * @see — Docs https://reach.tech/menu-button#menulink
 */
export const MenuLink = React.forwardRef(({ className, ...props }, reference) => (
    <ReachMenuLink ref={reference} {...props} className={classNames('dropdown-item', className)} />
)) as ForwardReferenceComponent<'a', MenuLinkProps>
