import React, { useCallback, useState } from 'react'

import { mdiMenu } from '@mdi/js'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router-dom'

import { Button, Icon, ProductStatusBadge, ProductStatusType } from '@sourcegraph/wildcard'

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
    ...props
}) => {
    const [isMobileExpanded, setIsMobileExpanded] = useState(false)
    const collapseMobileSidebar = useCallback((): void => setIsMobileExpanded(false), [])

    const siteAdminViewingOtherOrg = authenticatedUser && org.viewerCanAdminister && !org.viewerIsMember
    const context: OrgSettingsSidebarItemConditionContext = {
        batchChangesEnabled: props.batchChangesEnabled,
        batchChangesExecutionEnabled: props.batchChangesExecutionEnabled,
        batchChangesWebhookLogsEnabled: props.batchChangesWebhookLogsEnabled,
        org,
        authenticatedUser,
        isSourcegraphDotCom: props.isSourcegraphDotCom,
    }

    return (
        <>
            <Button className="d-sm-none align-self-start mb-3" onClick={() => setIsMobileExpanded(!isMobileExpanded)}>
                <Icon aria-hidden={true} svgPath={mdiMenu} className="mr-2" />
                {isMobileExpanded ? 'Hide' : 'Show'} menu
            </Button>
            <div
                className={classNames(
                    styles.orgSettingsSidebar,
                    className,
                    'd-sm-block',
                    !isMobileExpanded && 'd-none'
                )}
            >
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
                                <SidebarNavItem
                                    key={label}
                                    to={match.path + to}
                                    exact={exact}
                                    onClick={collapseMobileSidebar}
                                >
                                    {label} {status && <ProductStatusBadge className="ml-1" status={status} />}
                                </SidebarNavItem>
                            )
                    )}
                </SidebarGroup>
            </div>
        </>
    )
}
