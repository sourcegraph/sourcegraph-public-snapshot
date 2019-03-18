import * as React from 'react'
import { NavLink, RouteComponentProps } from 'react-router-dom'
import { SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS } from '../../components/Sidebar'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { OrgAreaPageProps } from '../area/OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {
    className?: string
}

/**
 * Sidebar for org settings pages
 */
export const OrgAccountSidebar: React.SFC<Props> = ({ org, authenticatedUser, className, match }) => {
    if (!org) {
        return null
    }

    const siteAdminViewingOtherOrg = authenticatedUser && org.viewerCanAdminister && !org.viewerIsMember

    return (
        <div className={`org-account-sidebar ${className || ''}`}>
            {/* Indicate when the site admin is viewing another org's settings */}
            {siteAdminViewingOtherOrg && (
                <SiteAdminAlert className="sidebar__alert">
                    Viewing settings for <strong>{org.name}</strong>
                </SiteAdminAlert>
            )}

            <div className="card">
                <div className="card-header">Organization</div>
                <div className="list-group list-group-flush">
                    <NavLink to={`${match.url}/profile`} exact={true} className={SIDEBAR_LIST_GROUP_ITEM_ACTION_CLASS}>
                        Profile
                    </NavLink>
                </div>
            </div>
        </div>
    )
}
