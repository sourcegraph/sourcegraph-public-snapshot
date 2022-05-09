import * as React from 'react'

import classNames from 'classnames'
import { RouteComponentProps } from 'react-router-dom'

import { SidebarGroup, SidebarGroupHeader, SidebarNavItem } from '../../components/Sidebar'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { OrgAreaPageProps } from '../area/OrgArea'

import styles from './OrgMembersSidebar.module.scss'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {
    className?: string
}

/**
 * Sidebar for org members pages.
 */
export const OrgMembersSidebar: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    org,
    authenticatedUser,
    className,
    match,
}) => {
    if (!org) {
        return null
    }

    const siteAdminViewingOtherOrg = authenticatedUser && org.viewerCanAdminister && !org.viewerIsMember

    return (
        <div className={classNames(styles.orgMembersSidebar, className)}>
            {/* Indicate when the site admin is viewing another org's members */}
            {siteAdminViewingOtherOrg && (
                <SiteAdminAlert className="sidebar__alert">
                    Viewing members for <strong>{org.name}</strong>
                </SiteAdminAlert>
            )}

            <SidebarGroup>
                <SidebarGroupHeader label="Organization" />
                <SidebarNavItem to={match.url} exact={true}>
                    Members
                </SidebarNavItem>
                <SidebarNavItem to={`${match.url}/pending-invites`} exact={true}>
                    Pending invites
                </SidebarNavItem>
            </SidebarGroup>
        </div>
    )
}
