import { MenuLink as ReachMenuLink, MenuLinkProps as ReachMenuLinkProps } from '@reach/menu-button'
import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

export type MenuLinkProps = ReachMenuLinkProps

export const MenuLink = React.forwardRef(({ className, ...props }, reference) => (
    <ReachMenuLink ref={reference} {...props} className={classNames('dropdown-item', className)} />
)) as ForwardReferenceComponent<'a', MenuLinkProps>
