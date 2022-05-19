import React from 'react'

import classNames from 'classnames'
import kebabCase from 'lodash/kebabCase'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import { useRouteMatch } from 'react-router-dom'

import {
    AnchorLink,
    ButtonLink,
    Icon,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    Typography,
} from '@sourcegraph/wildcard'

import styles from './Sidebar.module.scss'

/**
 * Item of `SideBarGroup`.
 */
export const SidebarNavItem: React.FunctionComponent<
    React.PropsWithChildren<{
        to: string
        className?: string
        exact?: boolean
        source?: string
    }>
> = ({ children, className, to, exact, source }) => {
    const buttonClassNames = classNames('text-left d-flex', styles.linkInactive, className)
    const routeMatch = useRouteMatch({ path: to, exact })

    if (source === 'server') {
        return (
            <ButtonLink as={AnchorLink} to={to} className={classNames(buttonClassNames, className)}>
                {children}
            </ButtonLink>
        )
    }

    return (
        <ButtonLink to={to} className={buttonClassNames} variant={routeMatch?.isExact ? 'primary' : undefined}>
            {children}
        </ButtonLink>
    )
}
/**
 *
 * Header of a `SideBarGroup`
 */
export const SidebarGroupHeader: React.FunctionComponent<React.PropsWithChildren<{ label: string }>> = ({ label }) => (
    <Typography.H3 as={Typography.H2}>{label}</Typography.H3>
)

/**
 * Sidebar with collapsible items
 */
export const SidebarCollapseItems: React.FunctionComponent<
    React.PropsWithChildren<{
        children: React.ReactNode
        icon?: React.ComponentType<React.PropsWithChildren<{ className?: string }>>
        label?: string
        openByDefault?: boolean
    }>
> = ({ children, label, icon: CollapseItemIcon, openByDefault = false }) => (
    <Collapse openByDefault={openByDefault}>
        {({ isOpen }) => (
            <>
                <CollapseHeader
                    aria-expanded={isOpen}
                    aria-controls={kebabCase(label)}
                    type="button"
                    className="bg-2 border-0 d-flex justify-content-between list-group-item-action py-2 w-100"
                >
                    <span>
                        {CollapseItemIcon && <Icon className="mr-1" as={CollapseItemIcon} />} {label}
                    </span>
                    <Icon className={styles.chevron} as={isOpen ? MenuUpIcon : MenuDownIcon} />
                </CollapseHeader>
                <CollapsePanel id={kebabCase(label)} className="border-top">
                    {children}
                </CollapsePanel>
            </>
        )}
    </Collapse>
)

interface SidebarGroupProps {
    className?: string
}

/**
 * A box of items in the side bar. Use `SideBarGroupHeader` as children.
 */
export const SidebarGroup: React.FunctionComponent<React.PropsWithChildren<SidebarGroupProps>> = ({
    children,
    className,
}) => <div className={classNames('mb-3', styles.sidebar, className)}>{children}</div>
