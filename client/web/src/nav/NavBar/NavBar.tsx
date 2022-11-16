import React, { useEffect, useRef, useState, forwardRef } from 'react'

import { mdiChevronDown, mdiChevronUp, mdiMenu } from '@mdi/js'
import classNames from 'classnames'
import H from 'history'
import { LinkProps, NavLink as RouterLink } from 'react-router-dom'

import { Button, Link, Icon, H1, ForwardReferenceComponent } from '@sourcegraph/wildcard'

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
    className?: string
}

export interface NavLinkProps extends NavItemProps, Pick<LinkProps<H.LocationState>, 'to'> {
    external?: boolean
    className?: string
    variant?: 'compact'
}

const useOnClickDetector = (
    reference: React.RefObject<HTMLDivElement>
): [boolean, React.Dispatch<React.SetStateAction<boolean>>] => {
    const [onClick, setOnClick] = useState(false)

    useEffect(() => {
        function handleToggleOpen(): void {
            if (reference.current) {
                setOnClick(false)
            }
        }
        document.addEventListener('mouseup', handleToggleOpen)
        return () => {
            document.removeEventListener('mouseup', handleToggleOpen)
        }
    }, [reference, setOnClick])

    return [onClick, setOnClick]
}

export const NavBar = forwardRef(
    ({ children, logo }, reference): JSX.Element => (
        <nav aria-label="Main" className={navBarStyles.navbar} ref={reference}>
            <H1 className={navBarStyles.logo}>
                <RouterLink className="d-flex align-items-center" to={PageRoutes.Search}>
                    {logo}
                </RouterLink>
            </H1>
            <hr className={navBarStyles.divider} aria-hidden={true} />
            {children}
        </nav>
    )
) as ForwardReferenceComponent<'div', NavBarProps>

export const NavGroup = ({ children }: NavGroupProps): JSX.Element => {
    const menuReference = useRef<HTMLDivElement>(null)
    const [open, setOpen] = useOnClickDetector(menuReference)

    return (
        <div className={navBarStyles.menu} ref={menuReference}>
            <Button className={navBarStyles.menuButton} onClick={() => setOpen(!open)} aria-label="Sections Navigation">
                <Icon aria-hidden={true} svgPath={mdiMenu} />
                <Icon svgPath={open ? mdiChevronUp : mdiChevronDown} aria-hidden={true} />
            </Button>
            <ul className={classNames(navBarStyles.list, { [navBarStyles.menuClose]: !open })}>{children}</ul>
        </div>
    )
}

export const NavActions: React.FunctionComponent<React.PropsWithChildren<NavActionsProps>> = ({ children }) => (
    <ul className={navActionStyles.actions}>{children}</ul>
)

export const NavAction: React.FunctionComponent<React.PropsWithChildren<NavActionsProps>> = ({
    children,
    className,
}) => (
    <>
        {React.Children.map(children, action => (
            <li className={classNames(navActionStyles.action, className)}>{action}</li>
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
            {LinkIcon ? <Icon className={navItemStyles.icon} as={LinkIcon} aria-hidden={true} /> : null}
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
