import React, { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import { useLocation } from 'react-router'

import { Link, Menu, MenuButton, MenuLink, MenuList, EMPTY_RECTANGLE, Icon } from '@sourcegraph/wildcard'

import { NavItem, NavLink } from '.'

import styles from './NavDropdown.module.scss'
import navItemStyles from './NavItem.module.scss'

export interface NavDropdownItem {
    content: React.ReactNode | string
    path: string
}

interface NavDropdownProps {
    toggleItem: NavDropdownItem & {
        icon: React.ComponentType<{ className?: string }>
        // Alternative path to match against if item is active
        altPath?: string
    }
    // An extra item on mobile devices in the dropdown menu that serves as the "home" item instead of the toggle item.
    // It uses the path from the toggleItem.
    mobileHomeItem: Omit<NavDropdownItem, 'path'>
    // Items to display in the dropdown.
    items: NavDropdownItem[]
    // A current react router route match
    routeMatch?: string
}

export const NavDropdown: React.FunctionComponent<React.PropsWithChildren<NavDropdownProps>> = ({
    toggleItem,
    mobileHomeItem,
    items,
    routeMatch,
}) => {
    const location = useLocation()
    const isItemSelected = useMemo(
        () =>
            items.some(item => location.pathname.startsWith(item.path)) ||
            location.pathname.startsWith(toggleItem.path) ||
            routeMatch === toggleItem.altPath,
        [items, location.pathname, toggleItem.path, toggleItem.altPath, routeMatch]
    )

    const menuButtonReference = useRef<HTMLButtonElement>(null)
    const linkReference = useRef<HTMLAnchorElement>(null)

    const [isOverButton, setIsOverButton] = useState(false)
    const [isOverList, setIsOverList] = useState(false)

    // Use this func for toggling menu
    const triggerMenuButtonEvent = useCallback(() => {
        menuButtonReference.current!.dispatchEvent(new Event('mousedown', { bubbles: true }))
    }, [])

    useLayoutEffect(() => {
        const isOpen = menuButtonReference.current!.hasAttribute('aria-expanded')

        if (isOpen && !isOverButton && !isOverList) {
            triggerMenuButtonEvent()

            return
        }

        if (!isOpen && (isOverButton || isOverList)) {
            triggerMenuButtonEvent()
        }
    }, [isOverButton, isOverList, triggerMenuButtonEvent])

    useEffect(() => {
        const currentLink = linkReference.current!
        const currentMenuButton = menuButtonReference.current!

        const handleMenuButtonTouchEnd = (event: TouchEvent): void => {
            event.preventDefault()
            triggerMenuButtonEvent()
        }
        const handleLinkTouchEnd = (event: TouchEvent): void => {
            // preventDefault would help to block navigation when touching on Link
            event.preventDefault()
            triggerMenuButtonEvent()
        }

        // Have to add/remove `touchend` manually like this to prevent
        // page navigation on touch screen (onTouchEnd binding doesn't work)
        currentMenuButton.addEventListener('touchend', handleMenuButtonTouchEnd)
        currentLink.addEventListener('touchend', handleLinkTouchEnd)

        return () => {
            currentMenuButton.removeEventListener('touchend', handleMenuButtonTouchEnd)
            currentLink.removeEventListener('touchend', handleLinkTouchEnd)
        }
    }, [triggerMenuButtonEvent])

    // We render the bigger screen version (dropdown) together with the smaller screen version (list of nav items)
    // and then use CSS @media queries to toggle between them.
    return (
        <>
            {/*
                Add `position-relative` here for `absolute` position of `MenuButton` below
                => `MenuButton` won't change its height when hovering + indicator
                => `MenuList` won't change its position when opening
            */}
            <NavItem className="d-none d-md-flex position-relative">
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
                                onMouseEnter={() => setIsOverButton(true)}
                                onMouseLeave={() => setIsOverButton(false)}
                            >
                                <div
                                    className={classNames(
                                        'h-100 d-flex',
                                        navItemStyles.linkContent,
                                        styles.navDropdownWrapper
                                    )}
                                >
                                    <Link
                                        to={toggleItem.path}
                                        className={classNames(styles.navDropdownLink, navItemStyles.itemFocusable)}
                                        ref={linkReference}
                                    >
                                        <span className={navItemStyles.itemFocusableContent}>
                                            <Icon
                                                role="img"
                                                className={navItemStyles.icon}
                                                as={toggleItem.icon}
                                                aria-hidden={true}
                                            />
                                            <span
                                                className={classNames(navItemStyles.text, navItemStyles.iconIncluded)}
                                            >
                                                {toggleItem.content}
                                            </span>
                                        </span>
                                    </Link>
                                    <MenuButton
                                        className={classNames(
                                            styles.navDropdownIconButton,
                                            navItemStyles.itemFocusable
                                        )}
                                        ref={menuButtonReference}
                                        aria-label={isExpanded ? 'Hide search menu' : 'Show search menu'}
                                    >
                                        <span className={navItemStyles.itemFocusableContent}>
                                            <Icon
                                                role="img"
                                                className={navItemStyles.icon}
                                                as={isExpanded ? ChevronUpIcon : ChevronDownIcon}
                                                aria-hidden={true}
                                            />
                                        </span>
                                    </MenuButton>
                                </div>
                            </div>

                            <MenuList
                                className={styles.navDropdownContainer}
                                targetPadding={EMPTY_RECTANGLE}
                                onMouseEnter={() => setIsOverList(true)}
                                onMouseLeave={() => setIsOverList(false)}
                            >
                                <MenuLink
                                    as={Link}
                                    key={toggleItem.path}
                                    to={toggleItem.path}
                                    className={styles.showOnTouchScreen}
                                    index={-1}
                                >
                                    {mobileHomeItem.content}
                                </MenuLink>
                                {items.map(item => (
                                    <MenuLink as={Link} key={item.path} to={item.path}>
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
