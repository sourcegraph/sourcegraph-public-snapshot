import classNames from 'classnames'
import { noop } from 'lodash'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { ComponentType, forwardRef, useCallback, useState } from 'react'

import { ForwardReferenceComponent, Menu, MenuButton, MenuItem, MenuList, Position } from '@sourcegraph/wildcard'

import styles from './MenuNavItem.module.scss'

interface MenuNavItemProps {
    children: React.ReactNode
    openByDefault?: boolean
    as?: ComponentType
    position?: Position
}

/**
 * Displays a dropdown menu in the navbar
 * displaiyng navigation links as menu items
 *
 */
export const MenuNavItem: React.FunctionComponent<MenuNavItemProps> = forwardRef((props, reference) => {
    const { children, openByDefault, as: Component, position = Position.bottomStart } = props

    const [isOpen, setIsOpen] = useState(() => !!openByDefault)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])

    return (
        <Menu isOpen={isOpen} onOpenChange={toggleIsOpen} ref={reference} as={Component}>
            {({ isExpanded }) => (
                <>
                    <MenuButton className={classNames('bg-transparent', styles.menuNavItem)}>
                        <MenuIcon className="icon-inline" />
                        {isExpanded ? <MenuUpIcon className="icon-inline" /> : <MenuDownIcon className="icon-inline" />}
                    </MenuButton>
                    <MenuList position={position}>
                        {React.Children.map(children, child => child && <MenuItem onSelect={noop}>{child}</MenuItem>)}
                    </MenuList>
                </>
            )}
        </Menu>
    )
}) as ForwardReferenceComponent<'div', MenuNavItemProps>
