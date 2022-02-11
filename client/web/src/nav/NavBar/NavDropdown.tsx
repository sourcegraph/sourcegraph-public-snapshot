import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import React, { useMemo, useRef, useState } from 'react'
import { useHistory, useLocation } from 'react-router'

import {
    Menu,
    MenuButton,
    Link,
    MenuList,
    MenuLink,
    PopoverOpenEvent,
    PopoverOpenEventReason,
} from '@sourcegraph/wildcard'

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
    const history = useHistory()
    const menuListReference = useRef<HTMLDivElement | null>(null)

    const [isDropdownOpen, setIsDropdownOpen] = useState(false)
    const handleKeyDown = (event: React.KeyboardEvent): void => {
        // Don't toggle dropdown with Enter key; instead, navigate to the home item of the dropdown.
        // This matches the behavior of the rest of the nav items.
        if (event.type === 'keydown') {
            switch (event.key) {
                case 'Enter': {
                    history.push(toggleItem.path)
                    event.preventDefault()
                    return
                }
                case 'ArrowUp':
                case 'ArrowDown': {
                    if (!isDropdownOpen) {
                        setIsDropdownOpen(true)
                    } else {
                        menuListReference.current?.focus()
                    }
                    return
                }
            }
        }

        setIsDropdownOpen(false)
    }

    const closeDropdown = (): void => setIsDropdownOpen(false)

    const isItemSelected = useMemo(
        () =>
            items.some(item => location.pathname.startsWith(item.path)) ||
            location.pathname.startsWith(toggleItem.path),
        [items, toggleItem, location.pathname]
    )

    const handleOpenChange = (event: PopoverOpenEvent): void => {
        if (event.reason === PopoverOpenEventReason.TriggerClick) {
            history.push(toggleItem.path)
        }

        setIsDropdownOpen(event.isOpen)
    }

    // We render the bigger screen version (dropdown) together with the smaller screen version (list of nav items)
    // and then use CSS @media queries to toggle between them.
    return (
        <>
            {/* Dropdown nav item for bigger screens */}
            <NavItem className="d-none d-md-flex">
                <Menu isOpen={isDropdownOpen} onOpenChange={handleOpenChange}>
                    <MenuButton
                        className={classNames(
                            navItemStyles.link,
                            isItemSelected && navItemStyles.active,
                            'd-flex',
                            'align-items-center',
                            'p-0'
                        )}
                        onPointerEnter={(event: React.PointerEvent) => {
                            if (event.pointerType === 'mouse') {
                                setIsDropdownOpen(true)
                            }
                        }}
                        onPointerDown={(event: React.PointerEvent) => {
                            // Navigate to toggle item path on mouse click.
                            if (event.pointerType === 'mouse') {
                                history.push(toggleItem.path)
                            }
                        }}
                        onKeyDown={handleKeyDown}
                    >
                        <span className={navItemStyles.linkContent}>
                            <toggleItem.icon className={classNames('icon-inline', navItemStyles.icon)} />
                            <span className={classNames(navItemStyles.text, navItemStyles.iconIncluded)}>
                                {toggleItem.content}
                            </span>
                            {isDropdownOpen ? (
                                <ChevronUpIcon className={classNames('icon-inline', navItemStyles.icon)} />
                            ) : (
                                <ChevronDownIcon className={classNames('icon-inline', navItemStyles.icon)} />
                            )}
                        </span>
                    </MenuButton>
                    {/* MenuPoper does not have modifiers
                            Add onPointerLeave from ButtonDropdown Reactstrap library */}

                    <MenuList
                        onPointerLeave={(event: React.PointerEvent) => {
                            if (event.pointerType === 'mouse') {
                                closeDropdown()
                            }
                        }}
                        ref={menuListReference}
                    >
                        {/* This link does not have a role="menuitem" set, because it breaks the keyboard navigation for the dropdown when hidden. */}
                        <MenuLink
                            as={Link}
                            key={toggleItem.path}
                            to={toggleItem.path}
                            className={styles.showOnTouchScreen}
                            onClick={closeDropdown}
                        >
                            {mobileHomeItem.content}
                        </MenuLink>
                        {items.map(item => (
                            <MenuLink as={Link} key={item.path} to={item.path} onClick={closeDropdown} role="menuitem">
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
