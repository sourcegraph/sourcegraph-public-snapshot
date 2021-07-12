import classNames from 'classnames'
import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { SidebarGroup, SidebarGroupHeader, SidebarGroupItems, SidebarNavItem } from '../../components/Sidebar'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { OrgAreaPageProps } from '../area/OrgArea'

import styles from './OrgSettingsSidebar.module.scss'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {
    className?: string
}

/**
 * Sidebar for org settings pages.
 */
export const OrgSettingsSidebar: React.FunctionComponent<Props> = ({ org, authenticatedUser, className, match }) => {
    if (!org) {
        return null
    }

    const siteAdminViewingOtherOrg = authenticatedUser && org.viewerCanAdminister && !org.viewerIsMember

    return (
        <div className={classNames(styles.orgSettingsSidebar, className)}>
            {/* Indicate when the site admin is viewing another org's settings */}
            {siteAdminViewingOtherOrg && (
                <SiteAdminAlert className="sidebar__alert">
                    Viewing settings for <strong>{org.name}</strong>
                </SiteAdminAlert>
            )}

            <SidebarGroup>
                <SidebarGroupHeader label="Organization" />
                <SidebarGroupItems>
                    <SidebarNavItem to={match.url} exact={true}>
                        Organization Settings
                    </SidebarNavItem>
                    <SidebarNavItem to={`${match.url}/profile`} exact={true}>
                        Profile
                    </SidebarNavItem>
                    <SidebarNavItem to={`${match.url}/members`} exact={true}>
                        Members
                    </SidebarNavItem>
                </SidebarGroupItems>
            </SidebarGroup>
        </div>
    )
}
