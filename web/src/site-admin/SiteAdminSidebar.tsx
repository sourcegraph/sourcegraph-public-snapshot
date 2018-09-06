import FeedIcon from '@sourcegraph/icons/lib/Feed'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as React from 'react'
import { Link } from 'react-router-dom'
import {
    SIDEBAR_BUTTON_CLASS,
    SideBarGroup,
    SideBarGroupHeader,
    SideBarGroupItems,
    SideBarNavItem,
} from '../components/Sidebar'
import { USE_PLATFORM } from '../extensions/environment/ExtensionsEnvironment'

export interface SiteAdminSideBarItem {
    /** The text of the item */
    label: string
    /** The link destination */
    to: string
}

export type SiteAdminSideBarItems = Record<
    'primary' | 'secondary' | 'registry' | 'auth' | 'other',
    ReadonlyArray<SiteAdminSideBarItem>
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
        <SideBarGroup>
            <SideBarGroupHeader icon={ServerIcon} label="Site admin" />
            <SideBarGroupItems>
                {items.primary.map(({ label, to }) => (
                    <SideBarNavItem to={to} exact={true} key={label}>
                        {label}
                    </SideBarNavItem>
                ))}
            </SideBarGroupItems>
        </SideBarGroup>
        <SideBarGroup>
            {items.secondary.map(({ label, to }) => (
                <SideBarNavItem to={to} exact={true}>
                    {label}
                </SideBarNavItem>
            ))}
        </SideBarGroup>
        <SideBarGroup>
            <SideBarGroupHeader icon={LockIcon} label="Auth" />
            <SideBarGroupItems>
                {items.auth.map(({ label, to }) => (
                    <SideBarNavItem to={to} exact={true}>
                        {label}
                    </SideBarNavItem>
                ))}
            </SideBarGroupItems>
        </SideBarGroup>
        {USE_PLATFORM && (
            <SideBarGroup>
                <SideBarGroupHeader icon={PuzzleIcon} label="Registry" />
                <SideBarGroupItems>
                    {items.registry.map(({ label, to }) => (
                        <SideBarNavItem to={to} exact={true}>
                            {label}
                        </SideBarNavItem>
                    ))}
                </SideBarGroupItems>
            </SideBarGroup>
        )}
        <SideBarGroup>
            <SideBarGroupItems>
                {items.other.map(({ label, to }) => (
                    <SideBarNavItem to={to} exact={true}>
                        {label}
                    </SideBarNavItem>
                ))}
            </SideBarGroupItems>
        </SideBarGroup>

        <Link to="/api/console" className={SIDEBAR_BUTTON_CLASS}>
            <FeedIcon className="icon-inline" />
            API console
        </Link>
        <a href="/-/debug/" className={SIDEBAR_BUTTON_CLASS}>
            Instrumentation
        </a>
    </div>
)
