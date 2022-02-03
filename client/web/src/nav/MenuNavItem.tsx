import classNames from 'classnames'
import { noop } from 'lodash'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { ComponentType, forwardRef } from 'react'

import { ForwardReferenceComponent, Menu, MenuButton, MenuItem, MenuList } from '@sourcegraph/wildcard'

import styles from './MenuNavItem.module.scss'

interface MenuNavItemProps {
    children: React.ReactNode
    as?: ComponentType
}

/**
 * Displays a dropdown menu in the navbar
 * displaiyng navigation links as menu items
 *
 */
export const MenuNavItem: React.FunctionComponent<MenuNavItemProps> = forwardRef((props, reference) => {
    const { children, as: Component } = props

    return (
        <Menu ref={reference} as={Component}>
            {({ isExpanded }) => (
                <>
                    <MenuButton className={classNames('bg-transparent', styles.menuNavItem)}>
                        <MenuIcon className="icon-inline" />
                        {isExpanded ? <MenuUpIcon className="icon-inline" /> : <MenuDownIcon className="icon-inline" />}
                    </MenuButton>
                    <MenuList>
                        {React.Children.map(children, child => child && <MenuItem onSelect={noop}>{child}</MenuItem>)}
                    </MenuList>
                </>
            )}
        </Menu>
    )
}) as ForwardReferenceComponent<'div', MenuNavItemProps>
