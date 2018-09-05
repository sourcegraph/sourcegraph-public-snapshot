import AddIcon from '@sourcegraph/icons/lib/Add'
import FeedIcon from '@sourcegraph/icons/lib/Feed'
import SignOutIcon from '@sourcegraph/icons/lib/SignOut'
import * as React from 'react'
import { Link, NavLink, RouteComponentProps } from 'react-router-dom'
import {
    SIDEBAR_BUTTON_CLASS,
    SIDEBAR_CARD_CLASS,
    SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS,
} from '../../components/Sidebar'
import { authExp } from '../../enterprise/site-admin/SiteAdminAuthenticationProvidersPage'
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

            <div className={SIDEBAR_CARD_CLASS}>
                <div className="card-header">User account</div>
                <div className="list-group list-group-flush">
                    <NavLink
                        to={`${props.match.path}/profile`}
                        exact={true}
                        className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                    >
                        Profile
                    </NavLink>
                    {!siteAdminViewingOtherUser &&
                        !props.externalAuthEnabled && (
                            <NavLink
                                to={`${props.match.path}/account`}
                                exact={true}
                                className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                            >
                                Password
                            </NavLink>
                        )}
                    <NavLink
                        to={`${props.match.path}/emails`}
                        exact={true}
                        className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                    >
                        Emails
                    </NavLink>
                    {authExp && (
                        <NavLink
                            to={`${props.match.path}/external-accounts`}
                            exact={true}
                            className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}
                        >
                            External accounts
                        </NavLink>
                    )}
                    {window.context.accessTokensAllow !== 'none' && (
                        <NavLink to={`${props.match.path}/tokens`} className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}>
                            Access tokens
                        </NavLink>
                    )}
                </div>
            </div>

            {(props.user.organizations.nodes.length > 0 || !siteAdminViewingOtherUser) && (
                <div className={SIDEBAR_CARD_CLASS}>
                    <div className="card-header">Organizations</div>
                    <div className="list-group list-group-flush">
                        {props.user.organizations.nodes.map(org => (
                            <NavLink
                                key={org.id}
                                to={`/organizations/${org.name}/settings`}
                                className={`${SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS} text-truncate text-nowrap`}
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
