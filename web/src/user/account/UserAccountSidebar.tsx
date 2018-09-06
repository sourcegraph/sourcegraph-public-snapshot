import AddIcon from '@sourcegraph/icons/lib/Add'
import FeedIcon from '@sourcegraph/icons/lib/Feed'
import SignOutIcon from '@sourcegraph/icons/lib/SignOut'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import {
    SIDEBAR_BUTTON_CLASS,
    SideBarGroup,
    SideBarGroupHeader,
    SideBarGroupItems,
    SideBarNavItem,
} from '../../components/Sidebar'
import { OrgAvatar } from '../../org/OrgAvatar'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAreaPageProps } from '../area/UserArea'

interface Props extends UserAreaPageProps, RouteComponentProps<{}> {
    className?: string
    externalAuthEnabled: boolean
}

/** Sidebar for user account pages. */
export const UserAccountSidebar: React.SFC<Props> = props => {
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

            <SideBarGroup>
                <SideBarGroupHeader label="User account" />
                <SideBarGroupItems>
                    <SideBarNavItem to={`${props.match.path}/profile`} exact={true}>
                        Profile
                    </SideBarNavItem>
                    {!siteAdminViewingOtherUser &&
                        !props.externalAuthEnabled && (
                            <SideBarNavItem to={`${props.match.path}/account`} exact={true}>
                                Password
                            </SideBarNavItem>
                        )}
                    <SideBarNavItem to={`${props.match.path}/emails`} exact={true}>
                        Emails
                    </SideBarNavItem>
                    {true && (
                        <SideBarNavItem to={`${props.match.path}/external-accounts`} exact={true}>
                            External accounts
                        </SideBarNavItem>
                    )}
                    {window.context.accessTokensAllow !== 'none' && (
                        <SideBarNavItem to={`${props.match.path}/tokens`}>Access tokens</SideBarNavItem>
                    )}
                </SideBarGroupItems>
            </SideBarGroup>

            {(props.user.organizations.nodes.length > 0 || !siteAdminViewingOtherUser) && (
                <SideBarGroup>
                    <SideBarGroupHeader label="Organizations" />
                    <SideBarGroupItems>
                        {props.user.organizations.nodes.map(org => (
                            <SideBarNavItem
                                key={org.id}
                                to={`/organizations/${org.name}/settings`}
                                className="text-truncate text-nowrap"
                            >
                                <OrgAvatar org={org.name} className="d-inline-flex" /> {org.name}
                            </SideBarNavItem>
                        ))}
                    </SideBarGroupItems>
                    {!siteAdminViewingOtherUser && (
                        <div className="card-body">
                            <Link to="/organizations/new" className="btn btn-secondary btn-sm w-100">
                                <AddIcon className="icon-inline" /> New organization
                            </Link>
                        </div>
                    )}
                </SideBarGroup>
            )}
            {!siteAdminViewingOtherUser && (
                <Link to="/api/console" className={SIDEBAR_BUTTON_CLASS}>
                    <FeedIcon className="icon-inline" /> API console
                </Link>
            )}
            {!siteAdminViewingOtherUser && (
                <NavLink to={`${props.match.path}/integrations`} exact={true} className={SIDEBAR_BUTTON_CLASS}>
                    Integrations
                </NavLink>
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
                        <SignOutIcon className="icon-inline list-group-item-action-icon" /> Sign out
                    </a>
                )}
        </div>
    )
}

function logTelemetryOnSignOut(): void {
    eventLogger.log('SignOutClicked')
}
