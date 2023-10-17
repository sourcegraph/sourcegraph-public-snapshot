import React from 'react'

import {
    MenuItem as ReachMenuItem,
    MenuLink as ReachMenuLink,
    type MenuItemProps as ReachMenuItemProps,
    type MenuLinkProps as ReachMenuLinkProps,
} from '@reach/menu-button'
import { noop } from 'lodash'

import type { ForwardReferenceComponent } from '../../types'
import { AnchorLink } from '../Link/AnchorLink'

export type MenuDisabledItemProps = Omit<ReachMenuItemProps, 'onSelect' | 'disabled'>

export const MenuDisabledItem = React.forwardRef(({ children, ...props }, reference) => (
    <ReachMenuItem ref={reference} {...props} onSelect={noop} disabled={false} as={AriaDisabledDiv}>
        {children}
    </ReachMenuItem>
)) as ForwardReferenceComponent<'div', MenuDisabledItemProps>

const AriaDisabledDiv = React.forwardRef(({ children, ...props }, reference) => (
    <div ref={reference} {...props} aria-disabled="true">
        {children}
    </div>
)) as ForwardReferenceComponent<'div', MenuDisabledItemProps>

export type MenuDisabledLinkProps = Omit<ReachMenuLinkProps, 'onSelect' | 'disabled'>

export const MenuDisabledLink = React.forwardRef((props, reference) => (
    <ReachMenuLink ref={reference} {...props} onSelect={noop} disabled={false} as={AriaDisabledLink} />
)) as ForwardReferenceComponent<'a', MenuDisabledItemProps>

const AriaDisabledLink = React.forwardRef(({ children, as, onClick, ...props }, reference) => {
    const handleOnClick: React.MouseEventHandler<HTMLAnchorElement> = event => {
        event.preventDefault()

        onClick?.(event)
    }

    return (
        <AnchorLink ref={reference} {...props} onClick={handleOnClick} aria-disabled="true" to="">
            {children}
        </AnchorLink>
    )
}) as ForwardReferenceComponent<'a', MenuDisabledItemProps>
