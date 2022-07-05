import React from 'react'

import { MenuItem as ReachMenuItem, MenuItemProps as ReachMenuItemProps } from '@reach/menu-button'
import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../types'

import { MenuDisabledItem } from './MenuDisabledItem'

import styles from './MenuItem.module.scss'

export type MenuItemProps = ReachMenuItemProps

/**
 * A styled item within a `<Menu />` component.
 * This should be selectable by the user and should be used
 * to ensure each item is accessible.
 *
 * @see — Docs https://reach.tech/menu-button#menuitem
 */
export const MenuItem = React.forwardRef(({ children, className, disabled, ...props }, reference) => {
    const Component = disabled ? MenuDisabledItem : ReachMenuItem

    return (
        <Component ref={reference} {...props} className={classNames(styles.dropdownItem, className)}>
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<'div', MenuItemProps>
