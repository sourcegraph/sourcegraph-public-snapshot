import ClipboardAccountOutlineIcon from 'mdi-react/ClipboardAccountOutlineIcon'
import AccountMultipleIcon from 'mdi-react/AccountMultipleIcon'
import EarthIcon from 'mdi-react/EarthIcon'
import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'
import { SidebarGroup, SidebarGroupItems, SidebarNavItem } from '../../components/Sidebar'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { OrgAreaPageProps } from '../area/OrgArea'

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
        <div className={className}>
            {/* Indicate when the site admin is viewing another org's settings */}
            {siteAdminViewingOtherOrg && (
                <SiteAdminAlert className="sidebar__alert">
                    Viewing settings for <strong>{org.name}</strong>
                </SiteAdminAlert>
            )}

            <SidebarGroup>
                <SidebarGroupItems>
                    <SidebarNavItem icon={EarthIcon} to={match.url} exact={true}>
                        Organization Settings
                    </SidebarNavItem>
                    <SidebarNavItem icon={ClipboardAccountOutlineIcon} to={`${match.url}/profile`} exact={true}>
                        Profile
                    </SidebarNavItem>
                    <SidebarNavItem icon={AccountMultipleIcon} to={`${match.url}/members`} exact={true}>
                        Members
                    </SidebarNavItem>
                </SidebarGroupItems>
            </SidebarGroup>
        </div>
    )
}
