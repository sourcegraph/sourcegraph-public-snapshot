import React, { type FC, type PropsWithChildren } from 'react'

import { mdiMenuDown, mdiMenuUp } from '@mdi/js'
import classNames from 'classnames'
import kebabCase from 'lodash/kebabCase'
import { useMatch } from 'react-router-dom'

import {
    AnchorLink,
    ButtonLink,
    Icon,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    H2,
    H3,
    type ForwardReferenceComponent,
} from '@sourcegraph/wildcard'

import styles from './Sidebar.module.scss'

interface SidebarNavItemProps {
    to: string
    className?: string
    source?: string
    onClick?: () => void
    exact?: boolean
}

/**
 * Item of `SideBarGroup`.
 */
export const SidebarNavItem: FC<PropsWithChildren<SidebarNavItemProps>> = ({
    children,
    className,
    to,
    source,
    onClick,
    exact = false,
}) => {
    // Match nested routes too.
    const routeMatch = useMatch(to + (exact ? '' : '/*'))
    const buttonClassNames = classNames('text-left d-flex', styles.linkInactive, className)

    if (source === 'server') {
        return (
            <ButtonLink
                as={AnchorLink}
                to={to}
                variant={routeMatch ? 'primary' : undefined}
                className={classNames(buttonClassNames, className)}
                onClick={onClick}
            >
                {children}
            </ButtonLink>
        )
    }

    return (
        <ButtonLink to={to} className={buttonClassNames} variant={routeMatch ? 'primary' : undefined} onClick={onClick}>
            {children}
        </ButtonLink>
    )
}
/**
 *
 * Header of a `SideBarGroup`
 */
export const SidebarGroupHeader: React.FunctionComponent<React.PropsWithChildren<{ label: string }>> = ({ label }) => (
    <H3 as={H2}>{label}</H3>
)

interface SidebarCollapseItemsProps {
    children: React.ReactNode
    icon?: React.ComponentType<React.PropsWithChildren<{ className?: string }>>
    label?: string
    openByDefault?: boolean
}

const SidebarCollapseHeader = React.forwardRef(function SidebarCollapseHeader(props, reference) {
    const { label, 'aria-expanded': isOpen, className, icon: CollapseItemIcon, ...rest } = props

    return (
        <button
            aria-expanded={isOpen}
            aria-controls={kebabCase(label)}
            type="button"
            className={classNames(
                className,
                'bg-2 border-0 d-flex justify-content-between list-group-item-action py-2 w-100'
            )}
            ref={reference}
            {...rest}
        >
            <span>
                {CollapseItemIcon && <Icon className="mr-1" as={CollapseItemIcon} aria-hidden={true} />} {label}
            </span>
            <Icon aria-hidden={true} className={styles.chevron} svgPath={isOpen ? mdiMenuUp : mdiMenuDown} />
        </button>
    )
}) as ForwardReferenceComponent<'button', Pick<SidebarCollapseItemsProps, 'label' | 'icon'>>

/**
 * Sidebar with collapsible items
 */
export const SidebarCollapseItems: React.FunctionComponent<React.PropsWithChildren<SidebarCollapseItemsProps>> = ({
    children,
    label,
    openByDefault = false,
    ...rest
}) => (
    <Collapse openByDefault={openByDefault}>
        {/* Using `{({ isOpen }) => (<>...</>)` as children of `Collapse` will cause all contents inside `Collapse` rerender */}
        {/* It caused losing focusing state issue https://github.com/sourcegraph/sourcegraph/issues/35866 */}
        {/* Using `as` in `CollapseHeader` and getting `isOpen` from `aria-expanded` as an alternative */}
        <CollapseHeader as={SidebarCollapseHeader} label={label} {...rest} />
        <CollapsePanel id={kebabCase(label)} className="border-top">
            {children}
        </CollapsePanel>
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
