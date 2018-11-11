import AddIcon from 'mdi-react/AddIcon'
import ConsoleIcon from 'mdi-react/ConsoleIcon'
import LogoutIcon from 'mdi-react/LogoutIcon'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import * as GQL from '../../backend/graphqlschema'
import {
    SIDEBAR_BUTTON_CLASS,
    SidebarGroup,
    SidebarGroupHeader,
    SidebarGroupItems,
    SidebarNavItem,
} from '../../components/Sidebar'
import { OrgAvatar } from '../../org/OrgAvatar'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { eventLogger } from '../../tracking/eventLogger'
import { NavItemDescriptor } from '../../util/contributions'
import { UserAreaRouteContext } from '../area/UserArea'

export interface UserAccountSidebarItemConditionContext {
    authProviders: GQL.IAuthProvider[]
}

export type UserAccountSidebarItems = Record<
    'account',
    ReadonlyArray<NavItemDescriptor<UserAccountSidebarItemConditionContext>>
>

export interface UserAccountSidebarProps extends UserAreaRouteContext, RouteComponentProps<{}> {
    items: UserAccountSidebarItems
    authProviders: GQL.IAuthProvider[]
    className?: string
}

/** Sidebar for user account pages. */
export const UserAccountSidebar: React.SFC<UserAccountSidebarProps> = props => {
    if (!props.authenticatedUser) {
        return null
    }

    // When the site admin is viewing another user's account.
    const siteAdminViewingOtherUser = props.user.id !== props.authenticatedUser.id

    return (
        <div className={`user-account-sidebar ${props.className || ''}`}>
            {/* Indicate when the site admin is viewing another user's account */}
            {siteAdminViewingOtherUser && (
                <SiteAdminAlert className="sidebar__alert">
                    Viewing account for <strong>{props.user.username}</strong>
                </SiteAdminAlert>
            )}

            <SidebarGroup>
                <SidebarGroupHeader label="User account" />
                <SidebarGroupItems>
                    {props.items.account.map(
                        ({ label, to, exact, condition = () => true }) =>
                            condition({ authProviders: props.authProviders }) && (
                                <SidebarNavItem key={label} to={props.match.path + to} exact={exact}>
                                    {label}
                                </SidebarNavItem>
                            )
                    )}
                </SidebarGroupItems>
            </SidebarGroup>

            {(props.user.organizations.nodes.length > 0 || !siteAdminViewingOtherUser) && (
                <SidebarGroup>
                    <SidebarGroupHeader label="Organizations" />
                    <SidebarGroupItems>
                        {props.user.organizations.nodes.map(org => (
                            <SidebarNavItem
                                key={org.id}
                                to={`/organizations/${org.name}/settings`}
                                className="text-truncate text-nowrap"
                            >
                                <OrgAvatar org={org.name} className="d-inline-flex" /> {org.name}
                            </SidebarNavItem>
                        ))}
                    </SidebarGroupItems>
                    {!siteAdminViewingOtherUser && (
                        <div className="card-body">
                            <Link to="/organizations/new" className="btn btn-secondary btn-sm w-100">
                                <AddIcon className="icon-inline" /> New organization
                            </Link>
                        </div>
                    )}
                </SidebarGroup>
            )}
            {!siteAdminViewingOtherUser && (
                <Link to="/api/console" className={SIDEBAR_BUTTON_CLASS}>
                    <ConsoleIcon className="icon-inline" /> API console
                </Link>
            )}
            {props.authenticatedUser.siteAdmin && (
                <Link to="/site-admin" className={SIDEBAR_BUTTON_CLASS}>
                    Site admin
                </Link>
            )}
            {!siteAdminViewingOtherUser &&
                props.authenticatedUser.session &&
                props.authenticatedUser.session.canSignOut && (
                    <a href="/-/sign-out" className={SIDEBAR_BUTTON_CLASS} onClick={logTelemetryOnSignOut}>
                        <LogoutIcon className="icon-inline list-group-item-action-icon" /> Sign out
                    </a>
                )}
        </div>
    )
}

function logTelemetryOnSignOut(): void {
    eventLogger.log('SignOutClicked')
}
