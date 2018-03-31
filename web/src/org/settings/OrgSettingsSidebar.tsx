import CityIcon from '@sourcegraph/icons/lib/City'
import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { OrgAreaPageProps } from '../area/OrgArea'

interface Props extends OrgAreaPageProps {
    className: string
}

/**
 * Sidebar for org settings pages
 */
export const OrgSettingsSidebar: React.SFC<Props> = ({ org, authenticatedUser, className }) => {
    if (!org) {
        return null
    }

    const siteAdminViewingOtherOrg = authenticatedUser && org.viewerCanAdminister && !org.viewerIsMember

    return (
        <div className={`sidebar org-settings-sidebar ${className}`}>
            {/* Indicate when the site admin is viewing another org's settings */}
            {siteAdminViewingOtherOrg && (
                <SiteAdminAlert className="sidebar__alert">
                    Viewing settings for <strong>{org.name}</strong>
                </SiteAdminAlert>
            )}

            <ul className="sidebar__items">
                <div className="sidebar__header">
                    <div className="sidebar__header-icon">
                        <CityIcon className="icon-inline" />
                    </div>
                    <h5 className="sidebar__header-title">Settings</h5>
                </div>
                <li className="sidebar__item">
                    <NavLink
                        to={`/organizations/${org.name}/settings/profile`}
                        exact={true}
                        className="sidebar__item-link"
                        activeClassName="sidebar__item--active"
                    >
                        Profile
                    </NavLink>
                </li>
                <li className="sidebar__item">
                    <NavLink
                        to={`/organizations/${org.name}/settings/configuration`}
                        exact={true}
                        className="sidebar__item-link"
                        activeClassName="sidebar__item--active"
                    >
                        Configuration
                    </NavLink>
                </li>
            </ul>
        </div>
    )
}
