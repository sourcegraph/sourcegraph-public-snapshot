import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import React, { useMemo, useState } from 'react'
import { useLocation } from 'react-router'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { Link } from '@sourcegraph/shared/src/components/Link'

import navItemStyles from './NavItem.module.scss'

import { NavItem, NavLink } from '.'

interface NavDropdownItem {
    icon: React.ComponentType<{ className?: string }>
    content: React.ReactNode | string
    path: string
}

interface NavDropdownProps {
    // Items to display in the dropdown. The first item will be used as the dropdown toggle.
    items: NavDropdownItem[]
}

export const NavDropdown: React.FunctionComponent<NavDropdownProps> = ({ items }) => {
    const [isDropdownOpen, setIsDropdownOpen] = useState(false)
    const toggle = (): void => setIsDropdownOpen(!isDropdownOpen)

    const toggleItem = items[0]
    const location = useLocation()
    const isItemSelected = useMemo(() => items.some(item => location.pathname.startsWith(item.path)), [
        items,
        location.pathname,
    ])

    // We render the bigger screen version (dropdown) together with the smaller screen version (list of nav items)
    // and then use CSS @media queries to toggle between them.
    return (
        <>
            {/* Dropdown nav item for bigger screens */}
            <NavItem className="d-none d-md-flex">
                <ButtonDropdown isOpen={isDropdownOpen} toggle={toggle}>
                    <DropdownToggle
                        className={classNames(
                            navItemStyles.link,
                            isItemSelected && navItemStyles.active,
                            'd-flex',
                            'align-items-center',
                            'p-0'
                        )}
                        nav={true}
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
                    <DropdownMenu>
                        {items.map(item => (
                            <Link
                                key={item.path}
                                to={item.path}
                                className="dropdown-item"
                                onClick={() => setIsDropdownOpen(false)}
                            >
                                <item.icon className="icon-inline" /> {item.content}
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
            {items.slice(1).map(item => (
                <NavItem key={item.path} icon={item.icon} className="d-flex d-md-none">
                    <NavLink to={item.path} className="pl-2">
                        {item.content}
                    </NavLink>
                </NavItem>
            ))}
        </>
    )
}
