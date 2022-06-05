import React from 'react'

import { MenuItem as ReachMenuItem, MenuItemProps as ReachMenuItemProps } from '@reach/menu-button'
import classNames from 'classnames'
import { noop } from 'lodash'

import { ForwardReferenceComponent } from '../../types'

import styles from './MenuItem.module.scss'

export type MenuDisabledItemProps = Omit<ReachMenuItemProps, 'onSelect' | 'disabled'>

/**
 * A `disabled` styled item within a `<Menu />` component.
 * This MenuItem does nothing on select and is styled as `disabled` MenuItem (included `aria-disabled=true`)
 * but is still focusable
 *
 * @see — Docs https://reach.tech/menu-button#menuitem
 */
export const MenuDisabledItem = React.forwardRef(({ children, className, ...props }, reference) => (
    <ReachMenuItem
        ref={reference}
        {...props}
        onSelect={noop}
        disabled={false}
        as={AriaDisabledDiv}
        className={classNames('dropdown-item', styles.item, className)}
    >
        {children}
    </ReachMenuItem>
)) as ForwardReferenceComponent<'div', MenuDisabledItemProps>

const AriaDisabledDiv = React.forwardRef(({ children, ...props }, reference) => (
    <div ref={reference} {...props} aria-disabled="true">
        {children}
    </div>
)) as ForwardReferenceComponent<'div', MenuDisabledItemProps>
