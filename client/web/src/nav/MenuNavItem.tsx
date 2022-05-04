import React, { ComponentType, forwardRef } from 'react'

import classNames from 'classnames'
import { noop } from 'lodash'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'

import { ForwardReferenceComponent, Menu, MenuButton, MenuItem, MenuList, Position, Icon } from '@sourcegraph/wildcard'

import styles from './MenuNavItem.module.scss'

interface MenuNavItemProps {
    children: React.ReactNode
    as?: ComponentType<React.PropsWithChildren<unknown>>
    position?: Position
    menuButtonRef?: React.Ref<HTMLButtonElement>
}

/**
 * Displays a dropdown menu in the navbar
 * displaiyng navigation links as menu items
 *
 */
export const MenuNavItem: React.FunctionComponent<React.PropsWithChildren<MenuNavItemProps>> = forwardRef(
    (props, reference) => {
        const { children, as: Component, menuButtonRef, position = Position.bottomStart } = props

        return (
            <Menu ref={reference} as={Component}>
                {({ isExpanded }) => (
                    <>
                        <MenuButton className={classNames('bg-transparent', styles.menuNavItem)} ref={menuButtonRef}>
                            <Icon as={MenuIcon} />
                            <Icon as={isExpanded ? MenuUpIcon : MenuDownIcon} />
                        </MenuButton>
                        <MenuList position={position}>
                            {React.Children.map(
                                children,
                                child => child && <MenuItem onSelect={noop}>{child}</MenuItem>
                            )}
                        </MenuList>
                    </>
                )}
            </Menu>
        )
    }
) as ForwardReferenceComponent<'div', MenuNavItemProps>
