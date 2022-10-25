import * as React from 'react'

import classNames from 'classnames'
import { RouteComponentProps } from 'react-router-dom'

import { ProductStatusBadge, ProductStatusType } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BatchChangesProps } from '../../batches'
import { SidebarGroup, SidebarGroupHeader, SidebarNavItem } from '../../components/Sidebar'
import { OrgAreaOrganizationFields } from '../../graphql-operations'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { NavItemDescriptor } from '../../util/contributions'

import { OrgSettingsAreaRouteContext } from './OrgSettingsArea'

import styles from './OrgSettingsSidebar.module.scss'

export interface OrgSettingsSidebarItemConditionContext extends BatchChangesProps {
    org: OrgAreaOrganizationFields
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'siteAdmin' | 'tags'>
    isSourcegraphDotCom: boolean
    newMembersInviteEnabled: boolean
}

type OrgSettingsSidebarItem = NavItemDescriptor<OrgSettingsSidebarItemConditionContext> & {
    status?: ProductStatusType
}

export type OrgSettingsSidebarItems = readonly OrgSettingsSidebarItem[]

export interface OrgSettingsSidebarProps
    extends OrgSettingsAreaRouteContext,
        BatchChangesProps,
        RouteComponentProps<{}> {
    items: OrgSettingsSidebarItems
    isSourcegraphDotCom: boolean
    className?: string
}

/**
 * Sidebar for org settings pages.
 */
export const OrgSettingsSidebar: React.FunctionComponent<React.PropsWithChildren<OrgSettingsSidebarProps>> = ({
    org,
    authenticatedUser,
    className,
    match,
    newMembersInviteEnabled,
    ...props
}) => {
    const siteAdminViewingOtherOrg = authenticatedUser && org.viewerCanAdminister && !org.viewerIsMember
    const context: OrgSettingsSidebarItemConditionContext = {
        batchChangesEnabled: props.batchChangesEnabled,
        batchChangesExecutionEnabled: props.batchChangesExecutionEnabled,
        batchChangesWebhookLogsEnabled: props.batchChangesWebhookLogsEnabled,
        org,
        authenticatedUser,
        isSourcegraphDotCom: props.isSourcegraphDotCom,
        newMembersInviteEnabled,
    }

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
                {props.items.map(
                    ({ label, to, exact, status, condition = () => true }) =>
                        condition(context) && (
                            <SidebarNavItem key={label} to={match.path + to} exact={exact}>
                                {label} {status && <ProductStatusBadge className="ml-1" status={status} />}
                            </SidebarNavItem>
                        )
                )}
            </SidebarGroup>
        </div>
    )
}
