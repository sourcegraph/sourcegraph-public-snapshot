import AddIcon from '@sourcegraph/icons/lib/Add'
import FeedIcon from '@sourcegraph/icons/lib/Feed'
import SignOutIcon from '@sourcegraph/icons/lib/SignOut'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import { OrgAvatar } from '../../org/OrgAvatar'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { authExp } from '../../site-admin/SiteAdminAuthenticationProvidersPage'
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

            <div className="card mb-3">
                <div className="card-header">User account</div>
                <div className="list-group list-group-flush">
                    <NavLink
                        to={`${props.match.path}/profile`}
                        exact={true}
                        className="list-group-item list-group-item-action"
                    >
                        Profile
                    </NavLink>
                    {!siteAdminViewingOtherUser &&
                        !props.externalAuthEnabled && (
                            <NavLink
                                to={`${props.match.path}/account`}
                                exact={true}
                                className="list-group-item list-group-item-action"
                            >
                                Password
                            </NavLink>
                        )}
                    <NavLink
                        to={`${props.match.path}/emails`}
                        exact={true}
                        className="list-group-item list-group-item-action"
                    >
                        Emails
                    </NavLink>
                    {authExp && (
                        <NavLink
                            to={`${props.match.path}/external-accounts`}
                            exact={true}
                            className="list-group-item list-group-item-action"
                        >
                            External accounts
                        </NavLink>
                    )}
                    {window.context.accessTokensEnabled && (
                        <NavLink to={`${props.match.path}/tokens`} className="list-group-item list-group-item-action">
                            Access tokens
                        </NavLink>
                    )}
                </div>
            </div>

            {(props.user.organizations.nodes.length > 0 || !siteAdminViewingOtherUser) && (
                <div className="card mb-3">
                    <div className="card-header">Organizations</div>
                    <div className="list-group list-group-flush">
                        {props.user.organizations.nodes.map(org => (
                            <NavLink
                                key={org.id}
                                to={`/organizations/${org.name}/settings`}
                                className="list-group-item list-group-item-action text-truncate text-nowrap"
                            >
                                <OrgAvatar org={org.name} className="d-inline-flex" /> {org.name}
                            </NavLink>
                        ))}
                    </div>
                    {!siteAdminViewingOtherUser && (
                        <div className="card-body">
                            <Link to="/organizations/new" className="btn btn-secondary btn-sm w-100">
                                <AddIcon className="icon-inline" /> New organization
                            </Link>
                        </div>
                    )}
                </div>
            )}
            {!siteAdminViewingOtherUser && (
                <Link to="/api/console" className="btn btn-secondary d-block w-100 my-2">
                    <FeedIcon className="icon-inline" /> API console
                </Link>
            )}
            {!siteAdminViewingOtherUser && (
                <NavLink
                    to={`${props.match.path}/integrations`}
                    exact={true}
                    className="btn btn-secondary d-block w-100 my-2"
                >
                    Integrations
                </NavLink>
            )}
            {props.authenticatedUser.siteAdmin && (
                <Link to="/site-admin" className="btn btn-secondary d-block w-100 my-2">
                    Site admin
                </Link>
            )}
            {!siteAdminViewingOtherUser &&
                props.authenticatedUser.session &&
                props.authenticatedUser.session.canSignOut && (
                    <a
                        href="/-/sign-out"
                        className="btn btn-secondary d-block w-100 my-2"
                        onClick={logTelemetryOnSignOut}
                    >
                        <SignOutIcon className="icon-inline list-group-item-action-icon" /> Sign out
                    </a>
                )}
        </div>
    )
}

function logTelemetryOnSignOut(): void {
    eventLogger.log('SignOutClicked')
}
