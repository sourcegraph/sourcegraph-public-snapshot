import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import React, { useMemo, useState, useCallback } from 'react'
import { useHistory, useLocation } from 'react-router'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { hasTouchScreen } from '@sourcegraph/shared/src/util/mobileDetection'

import navItemStyles from './NavItem.module.scss'

import { NavItem, NavLink } from '.'

interface NavDropdownItem {
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
    const isMobile = useMemo(() => hasTouchScreen(), [])
    const location = useLocation()
    const history = useHistory()

    const [isDropdownOpen, setIsDropdownOpen] = useState(false)
    const toggle = useCallback(
        (event: React.KeyboardEvent | React.MouseEvent) => {
            const isClick = event.type === 'click'
            const isEnter = event.type === 'keydown' && (event as React.KeyboardEvent).key === 'Enter'
            if (!isMobile && (isClick || isEnter)) {
                history.push(toggleItem.path)
                return
            }
            setIsDropdownOpen(!isDropdownOpen)
        },
        [history, isDropdownOpen, isMobile, toggleItem.path]
    )

    const isItemSelected = useMemo(
        () =>
            items.some(item => location.pathname.startsWith(item.path)) ||
            location.pathname.startsWith(toggleItem.path),
        [items, toggleItem, location.pathname]
    )

    // Add mobileHomeItem to dropdown items on mobile screens
    const dropdownItems = useMemo(
        () => (isMobile ? [{ ...mobileHomeItem, path: toggleItem.path }] : []).concat(items),
        [isMobile, toggleItem, mobileHomeItem, items]
    )

    // We render the bigger screen version (dropdown) together with the smaller screen version (list of nav items)
    // and then use CSS @media queries to toggle between them.
    return (
        <>
            {/* Dropdown nav item for bigger screens */}
            <NavItem className="d-none d-md-flex">
                <ButtonDropdown
                    isOpen={isDropdownOpen}
                    onMouseLeave={() => !isMobile && setIsDropdownOpen(false)}
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
                        onMouseEnter={() => !isMobile && setIsDropdownOpen(true)}
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
                    <DropdownMenu
                        modifiers={{
                            flip: {
                                enabled: false,
                            },
                            offset: {
                                enabled: true,
                                // Offset menu to the top so that the menu overlaps with the toggle button.
                                // This prevents the menu from closing when moving mouse cursor from the button
                                // to the menu.
                                offset: '0,-2',
                            },
                        }}
                    >
                        {dropdownItems.map(item => (
                            <Link
                                key={item.path}
                                to={item.path}
                                className="dropdown-item"
                                onClick={() => setIsDropdownOpen(false)}
                                role="menuitem"
                            >
                                {item.content}
                            </Link>
                        ))}
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
