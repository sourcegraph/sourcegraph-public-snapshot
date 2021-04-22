import classNames from 'classnames'
import H from 'history'
import React from 'react'
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
    collapse?: boolean
}

interface NavItemProps {
    icon?: React.ReactNode
    children: React.ReactNode
}

interface NavActions {
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

export const NavGroup = ({ children, collapse }: NavGroupProps): JSX.Element => (
    <ul className={styles.list}>{children}</ul>
)

export const NavActions = ({ children }: NavActions): JSX.Element => <ul className={styles.actions}>{children}</ul>
export const NavAction = ({ children }: NavActions): JSX.Element =>
    React.Children.map(children, action => <li className={styles.action}>{action}</li>)

export const NavItem = ({ children, icon }: NavItemProps): JSX.Element => {
    if (!children) {
        throw new Error('NavItem must be include at least one child')
    }

    return React.Children.map(children, child => <li className={styles.item}>{React.cloneElement(child, { icon })}</li>)
}

export const NavLink = ({ icon: Icon, children, to, external }: NavLinkProps): JSX.Element => {
    const content = (
        <span className={styles.linkContent}>
            {Icon && <Icon className={classNames('icon-inline', styles.icon)} />}
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
