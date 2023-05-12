import React, { useMemo } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'
import { useLocation } from 'react-router-dom'

import { Link, Menu, MenuButton, MenuLink, MenuList, EMPTY_RECTANGLE, Icon } from '@sourcegraph/wildcard'

import { NavItem, NavLink, NavLinkProps } from '.'

import styles from './NavDropdown.module.scss'
import navItemStyles from './NavItem.module.scss'

export interface NavDropdownItem {
    content: React.ReactNode | string

    /** To match against the current path to determine if the item is active */
    path: string

    target?: '_blank'
}

interface NavDropdownProps {
    toggleItem: NavDropdownItem & {
        icon: React.ComponentType<{ className?: string }>
        /** Alternative path to match against if item is active */
        altPath?: string
    } & Pick<NavLinkProps, 'variant'>
    /** The first item in the dropdown menu that serves as the "home" item.
        It uses the path from the toggleItem. */
    homeItem?: Omit<NavDropdownItem, 'path'>
    /** Items to display in the dropdown */
    items: NavDropdownItem[]
    /** A current react router route match */
    routeMatch?: string
    /** The name of the dropdown to use for accessible labels */
    name: string
}

export const NavDropdown: React.FunctionComponent<React.PropsWithChildren<NavDropdownProps>> = ({
    toggleItem,
    homeItem: homeItem,
    items,
    routeMatch,
    name,
}) => {
    const location = useLocation()
    const isItemSelected = useMemo(
        () =>
            items.some(item => location.pathname.startsWith(item.path)) ||
            location.pathname.startsWith(toggleItem.path) ||
            routeMatch === toggleItem.altPath,
        [items, location.pathname, toggleItem.path, toggleItem.altPath, routeMatch]
    )

    // We render the bigger screen version (dropdown) together with the smaller screen version (list of nav items)
    // and then use CSS @media queries to toggle between them.
    return (
        <>
            <NavItem className={classNames('d-none d-sm-flex', styles.wrapper)}>
                <Menu>
                    {({ isExpanded }) => (
                        <>
                            <div
                                className={classNames(
                                    navItemStyles.link,
                                    isItemSelected && navItemStyles.active,
                                    'd-flex',
                                    'align-items-center',
                                    'p-0'
                                )}
                                data-test-id={toggleItem.path}
                                data-test-active={isItemSelected}
                            >
                                <MenuButton
                                    className={classNames(navItemStyles.itemFocusable, styles.button)}
                                    aria-label={isExpanded ? `Hide ${name} menu` : `Show ${name} menu`}
                                >
                                    <span className={navItemStyles.itemFocusableContent}>
                                        <Icon className={navItemStyles.icon} as={toggleItem.icon} aria-hidden={true} />
                                        <span
                                            className={classNames(navItemStyles.text, navItemStyles.iconIncluded, {
                                                [navItemStyles.isCompact]: toggleItem.variant === 'compact',
                                            })}
                                        >
                                            {toggleItem.content}
                                        </span>
                                        <Icon
                                            className={navItemStyles.icon}
                                            svgPath={isExpanded ? mdiChevronUp : mdiChevronDown}
                                            aria-hidden={true}
                                        />
                                    </span>
                                </MenuButton>
                            </div>

                            <MenuList className={styles.menuList} targetPadding={EMPTY_RECTANGLE}>
                                {homeItem && (
                                    <MenuLink as={Link} key={toggleItem.path} to={toggleItem.path}>
                                        {homeItem.content}
                                    </MenuLink>
                                )}
                                {items.map(item => (
                                    <MenuLink as={Link} key={item.path} to={item.path} target={item.target}>
                                        {item.content}
                                    </MenuLink>
                                ))}
                            </MenuList>
                        </>
                    )}
                </Menu>
            </NavItem>
            {/* All nav items for smaller screens */}
            {/* Render the toggle item separately */}
            {toggleItem.path !== '#' && (
                <NavItem icon={toggleItem.icon} className="d-flex d-sm-none">
                    <NavLink to={toggleItem.path}>{toggleItem.content}</NavLink>
                </NavItem>
            )}
            {/* Render the rest of the items and indent them to indicate a hierarchical structure */}
            {items.map(item => (
                <NavItem key={item.path} className="d-flex d-sm-none">
                    <NavLink to={item.path} className="pl-2" external={item.target === '_blank'}>
                        {item.content}
                    </NavLink>
                </NavItem>
            ))}
        </>
    )
}
