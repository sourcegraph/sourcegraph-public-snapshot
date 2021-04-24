import classNames from 'classnames'
import H from 'history'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import React, { useState } from 'react'
import { LinkProps, NavLink as RouterLink } from 'react-router-dom'

import navActionStyles from './NavAction.module.scss'
import navBarStyles from './NavBar.module.scss'
import navItemStyles from './NavItem.module.scss'

const styles = { ...navBarStyles, ...navItemStyles, ...navActionStyles }
interface NavBarProps {
    children: React.ReactNode
    logo: React.ReactNode
}

interface NavGroupProps {
    children: React.ReactNode
}

interface NavItemProps {
    icon?: React.ComponentType<{ className?: string }>
    children: React.ReactNode
}

interface NavActionsProps {
    children: React.ReactNode
}

interface NavLinkProps extends NavItemProps, Pick<LinkProps<H.LocationState>, 'to'> {
    external?: boolean
}

export const NavBar = ({ children, logo }: NavBarProps): JSX.Element => (
    <nav aria-label="Main Menu" className={styles.navbar}>
        <h1 className={styles.logo}>
            <RouterLink to="/search">{logo}</RouterLink>
        </h1>
        <hr className={styles.divider} />
        {children}
    </nav>
)

export const NavGroup = ({ children }: NavGroupProps): JSX.Element => {
    const [open, setOpen] = useState(true)

    return (
        <>
            <button
                className={classNames('btn', styles.menuButton)}
                type="button"
                onClick={() => setOpen(!open)}
                aria-label="Sections Navigation"
            >
                <MenuIcon className="icon-inline" />
                {open ? <ChevronDownIcon className="icon-inline" /> : <ChevronUpIcon className="icon-inline" />}
            </button>
            <ul className={classNames(styles.list, { [styles.menuClose]: open })}>{children}</ul>
        </>
    )
}

export const NavActions: React.FunctionComponent<NavActionsProps> = ({ children }) => (
    <ul className={styles.actions}>{children}</ul>
)

export const NavAction: React.FunctionComponent<NavActionsProps> = ({ children }) => (
    <>
        {React.Children.map(children, action => (
            <li className={styles.action}>{action}</li>
        ))}
    </>
)

export const NavItem: React.FunctionComponent<NavItemProps> = ({ children, icon }) => {
    if (!children) {
        throw new Error('NavItem must be include at least one child')
    }

    return (
        <>
            {React.Children.map(children, child => (
                <li className={styles.item}>{React.cloneElement(child as React.ReactElement, { icon })}</li>
            ))}
        </>
    )
}

export const NavLink: React.FunctionComponent<NavLinkProps> = ({ icon: Icon, children, to, external }) => {
    const content = (
        <span className={styles.linkContent}>
            {Icon ? <Icon className={classNames('icon-inline', styles.icon)} /> : null}
            <span className={classNames(styles.text, styles.focusVisible, { [styles.iconIncluded]: Icon })}>
                {children}
            </span>
        </span>
    )

    return external ? (
        <a href={to as string} className={styles.link}>
            {content}
        </a>
    ) : (
        <RouterLink to={to} className={styles.link} activeClassName={styles.active}>
            {content}
        </RouterLink>
    )
}
