import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import React, { useMemo, useState } from 'react'
import { useLocation } from 'react-router'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'

import { Link } from '@sourcegraph/shared/src/components/Link'

import styles from './NavDropdown.module.scss'
import navItemStyles from './NavItem.module.scss'

import { NavItem, NavLink } from '.'

interface NavDropdownItem {
    icon: React.ComponentType<{ className?: string }>
    content: JSX.Element | string
    path: string
}

interface NavDropdownProps {
    parent: NavDropdownItem
    items: NavDropdownItem[]
}

export const NavDropdown: React.FunctionComponent<NavDropdownProps> = ({ parent, items }) => {
    const [isDropdownOpen, setIsDropdownOpen] = useState(false)
    const toggle = (): void => setIsDropdownOpen(!isDropdownOpen)

    const allItems = useMemo(() => [parent].concat(items), [parent, items])

    const location = useLocation()
    const isItemSelected = useMemo(() => allItems.some(item => location.pathname.startsWith(item.path)), [
        allItems,
        location.pathname,
    ])

    // We render the bigger screen version together with the smaller screen version and then use CSS @media
    // queries to toggle between them.
    return (
        <>
            {/* Nav items for bigger screens */}
            <NavItem className={styles.hideSmDown}>
                <ButtonDropdown isOpen={isDropdownOpen} toggle={toggle}>
                    <DropdownToggle
                        className={classNames(
                            navItemStyles.link,
                            isItemSelected && navItemStyles.active,
                            styles.dropdownToggle
                        )}
                        nav={true}
                    >
                        <span className={navItemStyles.linkContent}>
                            <parent.icon className={classNames('icon-inline', navItemStyles.icon)} />
                            <span className={classNames(navItemStyles.text, navItemStyles.iconIncluded)}>
                                {parent.content}
                            </span>
                            {isDropdownOpen ? (
                                <ChevronUpIcon className={classNames('icon-inline', navItemStyles.icon)} />
                            ) : (
                                <ChevronDownIcon className={classNames('icon-inline', navItemStyles.icon)} />
                            )}
                        </span>
                    </DropdownToggle>
                    <DropdownMenu>
                        {allItems.map(item => (
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
            {/* Nav items for small screens */}
            <NavItem icon={parent.icon} className={styles.hideSmUp}>
                <NavLink to={parent.path}>{parent.content}</NavLink>
            </NavItem>
            {items.map(item => (
                <NavItem key={item.path} icon={item.icon} className={styles.hideSmUp}>
                    {/* Indent non-parent items to indicate a hierarchical structure */}
                    <NavLink to={item.path} className="pl-2">
                        {item.content}
                    </NavLink>
                </NavItem>
            ))}
        </>
    )
}
