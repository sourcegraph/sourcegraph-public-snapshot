import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import React, { useLayoutEffect, useMemo, useRef, useState } from 'react'
import { useLocation } from 'react-router'

import { Link, Menu, MenuButton, MenuLink, MenuList, Position } from '@sourcegraph/wildcard'

import styles from './NavDropdown.module.scss'
import navItemStyles from './NavItem.module.scss'

import { NavItem, NavLink } from '.'

export interface NavDropdownItem {
    content: React.ReactNode | string
    path: string
}

interface NavDropdownProps {
    toggleItem: NavDropdownItem & { icon: React.ComponentType<{ className?: string }> }
    // An extra item on mobile devices in the dropdown menu that serves as the "home" item instead of the toggle item.
    // It uses the path from the toggleItem.
    mobileHomeItem: Omit<NavDropdownItem, 'path'>
    // Items to display in the dropdown.
    items: NavDropdownItem[]
}

export const NavDropdown: React.FunctionComponent<NavDropdownProps> = ({ toggleItem, mobileHomeItem, items }) => {
    const location = useLocation()
    const isItemSelected = useMemo(
        () =>
            items.some(item => location.pathname.startsWith(item.path)) ||
            location.pathname.startsWith(toggleItem.path),
        [items, toggleItem, location.pathname]
    )

    const menuButtonReference = useRef<HTMLButtonElement>(null)

    const [isOverButton, setIsOverButton] = useState(false)
    const [isOverList, setIsOverList] = useState(false)

    const openMenu = (): void => {
        menuButtonReference.current!.dispatchEvent(new Event('mousedown', { bubbles: true }))
    }

    useLayoutEffect(() => {
        const isOpen = menuButtonReference.current!.hasAttribute('aria-expanded')

        if (isOpen && !isOverButton && !isOverList) {
            openMenu()

            return
        }

        if (!isOpen && (isOverButton || isOverList)) {
            openMenu()
        }
    }, [isOverButton, isOverList])

    // We render the bigger screen version (dropdown) together with the smaller screen version (list of nav items)
    // and then use CSS @media queries to toggle between them.
    return (
        <>
            <NavItem className="d-none d-md-flex">
                <Menu>
                    <div
                        className={classNames(
                            navItemStyles.link,
                            isItemSelected && navItemStyles.active,
                            'd-flex',
                            'align-items-center',
                            'p-0'
                        )}
                        onMouseEnter={() => {
                            setIsOverButton(true)
                        }}
                        onMouseLeave={() => {
                            setIsOverButton(false)
                        }}
                    >
                        <Link to={toggleItem.path}>
                            <span className={navItemStyles.linkContent}>
                                <toggleItem.icon className={classNames('icon-inline', navItemStyles.icon)} />
                                <span className={classNames(navItemStyles.text, navItemStyles.iconIncluded)}>
                                    {toggleItem.content}
                                </span>
                            </span>
                        </Link>
                        <MenuButton ref={menuButtonReference}>
                            <ChevronDownIcon className={classNames('icon-inline', navItemStyles.icon)} />
                        </MenuButton>
                    </div>

                    <MenuList
                        position={Position.bottomEnd}
                        onMouseEnter={() => {
                            setIsOverList(true)
                        }}
                        onMouseLeave={() => {
                            setIsOverList(false)
                        }}
                    >
                        {/* This link does not have a role="menuitem" set, because it breaks the keyboard navigation for the dropdown when hidden. */}
                        <MenuLink
                            as={Link}
                            key={toggleItem.path}
                            to={toggleItem.path}
                            className={styles.showOnTouchScreen}
                        >
                            {mobileHomeItem.content}
                        </MenuLink>
                        {items.map(item => (
                            <MenuLink as={Link} key={item.path} to={item.path} role="menuitem">
                                {item.content}
                            </MenuLink>
                        ))}
                    </MenuList>
                </Menu>
            </NavItem>
            {/* All nav items for smaller screens */}
            {/* Render the toggle item separately */}
            <NavItem icon={toggleItem.icon} className="d-flex d-md-none">
                <NavLink to={toggleItem.path}>{toggleItem.content}</NavLink>
            </NavItem>
            {/* Render the rest of the items and indent them to indicate a hierarchical structure */}
            {items.map(item => (
                <NavItem key={item.path} className="d-flex d-md-none">
                    <NavLink to={item.path} className="pl-2">
                        {item.content}
                    </NavLink>
                </NavItem>
            ))}
        </>
    )
}
