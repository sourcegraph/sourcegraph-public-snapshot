import React, { useEffect, useRef, useState } from 'react'

import classNames from 'classnames'
import H from 'history'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import { LinkProps, NavLink as RouterLink } from 'react-router-dom'

import { Button, Link, Icon, Typography } from '@sourcegraph/wildcard'

import { PageRoutes } from '../../routes.constants'

import navActionStyles from './NavAction.module.scss'
import navBarStyles from './NavBar.module.scss'
import navItemStyles from './NavItem.module.scss'

interface NavBarProps {
    children: React.ReactNode
    logo: React.ReactNode
}

interface NavGroupProps {
    children: React.ReactNode
}

interface NavItemProps {
    icon?: React.ComponentType<{ className?: string }>
    className?: string
    children: React.ReactNode
}

interface NavActionsProps {
    children: React.ReactNode
}

export interface NavLinkProps extends NavItemProps, Pick<LinkProps<H.LocationState>, 'to'> {
    external?: boolean
    className?: string
    variant?: 'compact'
}

const useOutsideClickDetector = (
    reference: React.RefObject<HTMLDivElement>
): [boolean, React.Dispatch<React.SetStateAction<boolean>>] => {
    const [outsideClick, setOutsideClick] = useState(false)

    useEffect(() => {
        function handleClickOutside(event: MouseEvent): void {
            if (reference.current && !reference.current.contains(event.target as Node | null)) {
                setOutsideClick(false)
            }
        }
        document.addEventListener('mouseup', handleClickOutside)
        return () => {
            document.removeEventListener('mouseup', handleClickOutside)
        }
    }, [reference, setOutsideClick])

    return [outsideClick, setOutsideClick]
}

export const NavBar = ({ children, logo }: NavBarProps): JSX.Element => (
    <nav aria-label="Main Menu" className={navBarStyles.navbar}>
        <Typography.H1 className={navBarStyles.logo}>
            <RouterLink className="d-flex align-items-center" to={PageRoutes.Search}>
                {logo}
            </RouterLink>
        </Typography.H1>
        <hr className={navBarStyles.divider} aria-hidden={true} />
        {children}
    </nav>
)

export const NavGroup = ({ children }: NavGroupProps): JSX.Element => {
    const menuReference = useRef<HTMLDivElement>(null)
    const [open, setOpen] = useOutsideClickDetector(menuReference)

    return (
        <div className={navBarStyles.menu} ref={menuReference}>
            <Button className={navBarStyles.menuButton} onClick={() => setOpen(!open)} aria-label="Sections Navigation">
                <Icon role="img" as={MenuIcon} aria-hidden={true} />
                <Icon role="img" as={open ? ChevronUpIcon : ChevronDownIcon} aria-hidden={true} />
            </Button>
            <ul className={classNames(navBarStyles.list, { [navBarStyles.menuClose]: !open })}>{children}</ul>
        </div>
    )
}

export const NavActions: React.FunctionComponent<React.PropsWithChildren<NavActionsProps>> = ({ children }) => (
    <ul className={navActionStyles.actions}>{children}</ul>
)

export const NavAction: React.FunctionComponent<React.PropsWithChildren<NavActionsProps>> = ({ children }) => (
    <>
        {React.Children.map(children, action => (
            <li className={navActionStyles.action}>{action}</li>
        ))}
    </>
)

export const NavItem: React.FunctionComponent<React.PropsWithChildren<NavItemProps>> = ({
    children,
    className,
    icon,
}) => {
    if (!children) {
        throw new Error('NavItem must be include at least one child')
    }

    return (
        <>
            {React.Children.map(children, child => (
                <li className={classNames(navItemStyles.item, className)}>
                    {React.cloneElement(child as React.ReactElement, { icon })}
                </li>
            ))}
        </>
    )
}

export const NavLink: React.FunctionComponent<React.PropsWithChildren<NavLinkProps>> = ({
    icon: LinkIcon,
    children,
    to,
    external,
    variant,
    className,
}) => {
    const content = (
        <span className={classNames(navItemStyles.linkContent, className)}>
            {LinkIcon ? <Icon role="img" className={navItemStyles.icon} as={LinkIcon} aria-hidden={true} /> : null}
            <span
                className={classNames(navItemStyles.text, {
                    [navItemStyles.iconIncluded]: Icon,
                    [navItemStyles.isCompact]: variant === 'compact',
                })}
            >
                {children}
            </span>
        </span>
    )

    if (external) {
        return (
            <Link to={to as string} rel="noreferrer noopener" target="_blank" className={navItemStyles.link}>
                {content}
            </Link>
        )
    }

    return (
        <RouterLink to={to} className={navItemStyles.link} activeClassName={navItemStyles.active}>
            {content}
        </RouterLink>
    )
}
