import FeedIcon from '@sourcegraph/icons/lib/Feed'
import LockIcon from '@sourcegraph/icons/lib/Lock'
import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { SIDEBAR_BUTTON_CLASS, SideBarGroup, SideBarGroupHeader, SideBarNavItem } from '../components/Sidebar'
import { USE_PLATFORM } from '../extensions/environment/ExtensionsEnvironment'

export interface SiteAdminNavbarItem {
    label: string
    link: {
        to: string
        exact: boolean
    }
}

export interface SiteAdminNavbarGroup {
    title: {
        label: string
        icon: Icon
    }
    items: ReadonlyArray<SiteAdminNavbarItem>
}

export interface SiteAdminSidebarProps {
    history: H.History
    location: H.Location
    className: string
    user: GQL.IUser
}

/**
 * Sidebar for the site admin area.
 */
export const SiteAdminSidebar: React.SFC<SiteAdminSidebarProps> = ({ className }) => (
    <div className={`site-admin-sidebar ${className}`}>
        <SideBarGroup>
            <SideBarGroupHeader icon={ServerIcon} label="Site admin" />
            <SideBarNavItem to="/site-admin" exact={true}>
                Overview
            </SideBarNavItem>
            <SideBarNavItem to="/site-admin/configuration" exact={true}>
                Configuration
            </SideBarNavItem>
            <SideBarNavItem to="/site-admin/repositories" exact={true}>
                Repositories
            </SideBarNavItem>
        </SideBarGroup>
        <SideBarGroup>
            <SideBarNavItem to="/site-admin/users" exact={true}>
                Users
            </SideBarNavItem>
            <SideBarNavItem to="/site-admin/organizations" exact={true}>
                Organizations
            </SideBarNavItem>
            <SideBarNavItem to="/site-admin/global-settings" exact={true}>
                Global settings
            </SideBarNavItem>
            <SideBarNavItem to="/site-admin/code-intelligence" exact={true}>
                Code intelligence
            </SideBarNavItem>
        </SideBarGroup>
        <SideBarGroup>
            <SideBarGroupHeader icon={LockIcon} label="Auth" />
            <SideBarNavItem to="/site-admin/auth/providers" exact={true}>
                Providers
            </SideBarNavItem>
            <SideBarNavItem to="/site-admin/auth/external-accounts" exact={true}>
                External accounts
            </SideBarNavItem>
            <SideBarNavItem to="/site-admin/tokens" exact={true}>
                Access tokens
            </SideBarNavItem>
        </SideBarGroup>
        {USE_PLATFORM && (
            <SideBarGroup>
                <SideBarGroupHeader icon={PuzzleIcon} label="Registry" />
                <SideBarNavItem to="/site-admin/registry/extensions" exact={true}>
                    Extensions
                </SideBarNavItem>
            </SideBarGroup>
        )}
        <SideBarGroup>
            <SideBarNavItem to="/site-admin/updates" exact={true}>
                Updates
            </SideBarNavItem>
            <SideBarNavItem to="/site-admin/analytics" exact={true}>
                Analytics
            </SideBarNavItem>
            <SideBarNavItem to="/site-admin/surveys" exact={true}>
                User surveys
            </SideBarNavItem>
            <SideBarNavItem to="/site-admin/pings" exact={true}>
                Pings
            </SideBarNavItem>
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
