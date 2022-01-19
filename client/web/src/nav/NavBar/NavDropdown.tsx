import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import React, { useMemo, useState } from 'react'
import { useHistory, useLocation } from 'react-router'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { Link } from '@sourcegraph/shared/src/components/Link'

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

const DROPDOWN_MODIFIERS = {
    flip: {
        enabled: false,
    },
    offset: {
        enabled: true,
        // Offset menu to the top so that the menu overlaps with the toggle button.
        // This prevents the menu from closing when moving mouse cursor from the button
        // to the menu.
        offset: '-10,-2',
    },
}

export const NavDropdown: React.FunctionComponent<NavDropdownProps> = ({ toggleItem, mobileHomeItem, items }) => {
    const location = useLocation()
    const history = useHistory()

    const [isDropdownOpen, setIsDropdownOpen] = useState(false)
    const toggle = (event: React.KeyboardEvent | React.MouseEvent): void => {
        // Don't toggle dropdown with Enter key; instead, navigate to the home item of the dropdown.
        // This matches the behavior of the rest of the nav items.
        if (event.type === 'keydown' && (event as React.KeyboardEvent).key === 'Enter') {
            history.push(toggleItem.path)
            return
        }
        // For all other events that toggle the dropdown (e.g. space, arrow keys, clicking outside of the dropdown), toggle as normal.
        setIsDropdownOpen(!isDropdownOpen)
    }

    const closeDropdown = (): void => setIsDropdownOpen(false)

    const isItemSelected = useMemo(
        () =>
            items.some(item => location.pathname.startsWith(item.path)) ||
            location.pathname.startsWith(toggleItem.path),
        [items, toggleItem, location.pathname]
    )

    // We render the bigger screen version (dropdown) together with the smaller screen version (list of nav items)
    // and then use CSS @media queries to toggle between them.
    return (
        <>
            {/* Dropdown nav item for bigger screens */}
            <NavItem className="d-none d-md-flex">
                <ButtonDropdown
                    isOpen={isDropdownOpen}
                    onPointerLeave={(event: React.PointerEvent) => {
                        if (event.pointerType === 'mouse') {
                            closeDropdown()
                        }
                    }}
                    toggle={toggle}
                >
                    <DropdownToggle
                        className={classNames(
                            navItemStyles.link,
                            isItemSelected && navItemStyles.active,
                            'd-flex',
                            'align-items-center',
                            'p-0'
                        )}
                        nav={true}
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
                    </DropdownToggle>
                    <DropdownMenu modifiers={DROPDOWN_MODIFIERS}>
                        <>
                            {/* This link does not have a role="menuitem" set, because it breaks the keyboard navigation for the dropdown when hidden. */}
                            <Link
                                key={toggleItem.path}
                                to={toggleItem.path}
                                className={classNames('dropdown-item', styles.showOnTouchScreen)}
                                onClick={closeDropdown}
                            >
                                {mobileHomeItem.content}
                            </Link>
                            {items.map(item => (
                                <Link
                                    key={item.path}
                                    to={item.path}
                                    className="dropdown-item"
                                    onClick={closeDropdown}
                                    role="menuitem"
                                >
                                    {item.content}
                                </Link>
                            ))}
                        </>
                    </DropdownMenu>
                </ButtonDropdown>
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
