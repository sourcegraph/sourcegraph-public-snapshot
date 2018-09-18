import FeedIcon from '@sourcegraph/icons/lib/Feed'
import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import LockIcon from 'mdi-react/LockIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import {
    SIDEBAR_BUTTON_CLASS,
    SidebarGroup,
    SidebarGroupHeader,
    SidebarGroupItems,
    SidebarNavItem,
} from '../components/Sidebar'
import { NavItemDescriptor } from '../util/contributions'

export type SiteAdminSideBarItems = Record<
    'primary' | 'secondary' | 'registry' | 'auth' | 'other',
    ReadonlyArray<NavItemDescriptor>
>

export interface SiteAdminSidebarProps {
    /** The items for the side bar, by group */
    items: SiteAdminSideBarItems
    className: string
}

/**
 * Sidebar for the site admin area.
 */
export const SiteAdminSidebar: React.SFC<SiteAdminSidebarProps> = ({ className, items }) => (
    <div className={`site-admin-sidebar ${className}`}>
        <SidebarGroup>
            <SidebarGroupHeader icon={ServerIcon} label="Site admin" />
            <SidebarGroupItems>
                {items.primary.map(
                    ({ label, to, exact, condition = () => true }) =>
                        condition({}) && (
                            <SidebarNavItem to={to} exact={exact} key={label}>
                                {label}
                            </SidebarNavItem>
                        )
                )}
            </SidebarGroupItems>
        </SidebarGroup>
        <SidebarGroup>
            <SidebarGroupItems>
                {items.secondary.map(
                    ({ label, to, exact, condition = () => true }) =>
                        condition({}) && (
                            <SidebarNavItem to={to} exact={exact} key={label}>
                                {label}
                            </SidebarNavItem>
                        )
                )}
            </SidebarGroupItems>
        </SidebarGroup>
        <SidebarGroup>
            <SidebarGroupHeader icon={LockIcon} label="Auth" />
            <SidebarGroupItems>
                {items.auth.map(
                    ({ label, to, exact, condition = () => true }) =>
                        condition({}) && (
                            <SidebarNavItem to={to} exact={exact} key={label}>
                                {label}
                            </SidebarNavItem>
                        )
                )}
            </SidebarGroupItems>
        </SidebarGroup>
        <SidebarGroup>
            <SidebarGroupHeader icon={PuzzleIcon} label="Registry" />
            <SidebarGroupItems>
                {items.registry.map(
                    ({ label, to, exact, condition = () => true }) =>
                        condition({}) && (
                            <SidebarNavItem to={to} exact={exact} key={label}>
                                {label}
                            </SidebarNavItem>
                        )
                )}
            </SidebarGroupItems>
        </SidebarGroup>
        <SidebarGroup>
            <SidebarGroupItems>
                {items.other.map(
                    ({ label, to, exact, condition = () => true }) =>
                        condition({}) && (
                            <SidebarNavItem to={to} exact={exact} key={label}>
                                {label}
                            </SidebarNavItem>
                        )
                )}
            </SidebarGroupItems>
        </SidebarGroup>

        <Link to="/api/console" className={SIDEBAR_BUTTON_CLASS}>
            <FeedIcon className="icon-inline" />
            API console
        </Link>
        <a href="/-/debug/" className={SIDEBAR_BUTTON_CLASS}>
            Instrumentation
        </a>
    </div>
)
