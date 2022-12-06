import React, { useState, useCallback } from 'react'

import { mdiMenu, mdiPlus } from '@mdi/js'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router-dom'

import { ProductStatusBadge, Button, Link, Icon, ProductStatusType } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BatchChangesProps } from '../../batches'
import { SidebarGroup, SidebarGroupHeader, SidebarNavItem } from '../../components/Sidebar'
import { UserSettingsAreaUserFields } from '../../graphql-operations'
import { OrgAvatar } from '../../org/OrgAvatar'
import { NavItemDescriptor } from '../../util/contributions'

import { UserSettingsAreaRouteContext } from './UserSettingsArea'

import styles from './UserSettingsSidebar.module.scss'

export interface UserSettingsSidebarItemConditionContext extends BatchChangesProps {
    user: UserSettingsAreaUserFields
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'siteAdmin' | 'tags'>
    isSourcegraphDotCom: boolean
}

type UserSettingsSidebarItem = NavItemDescriptor<UserSettingsSidebarItemConditionContext> & {
    status?: ProductStatusType
}

export type UserSettingsSidebarItems = readonly UserSettingsSidebarItem[]

export interface UserSettingsSidebarProps
    extends UserSettingsAreaRouteContext,
        BatchChangesProps,
        RouteComponentProps<{}> {
    items: UserSettingsSidebarItems
    isSourcegraphDotCom: boolean
    className?: string
}

/** Sidebar for user account pages. */
export const UserSettingsSidebar: React.FunctionComponent<
    React.PropsWithChildren<UserSettingsSidebarProps>
> = props => {
    const [isMobileExpanded, setIsMobileExpanded] = useState(false)
    const collapseMobileSidebar = useCallback((): void => setIsMobileExpanded(false), [])

    if (!props.authenticatedUser) {
        return null
    }

    // When the site admin is viewing another user's account.
    const siteAdminViewingOtherUser = props.user.id !== props.authenticatedUser.id
    const context: UserSettingsSidebarItemConditionContext = {
        batchChangesEnabled: props.batchChangesEnabled,
        batchChangesExecutionEnabled: props.batchChangesExecutionEnabled,
        batchChangesWebhookLogsEnabled: props.batchChangesWebhookLogsEnabled,
        user: props.user,
        authenticatedUser: props.authenticatedUser,
        isSourcegraphDotCom: props.isSourcegraphDotCom,
    }

    return (
        <>
            <Button className="d-sm-none align-self-start mb-3" onClick={() => setIsMobileExpanded(!isMobileExpanded)}>
                <Icon aria-hidden={true} svgPath={mdiMenu} className="mr-2" />
                {isMobileExpanded ? 'Hide' : 'Show'} menu
            </Button>
            <div className={classNames(props.className, 'd-sm-block', !isMobileExpanded && 'd-none')}>
                <SidebarGroup>
                    <SidebarGroupHeader label="Account" />
                    {props.items.map(
                        ({ label, to, exact, status, condition = () => true }) =>
                            condition(context) && (
                                <SidebarNavItem
                                    key={label}
                                    to={props.match.path + to}
                                    exact={exact}
                                    onClick={collapseMobileSidebar}
                                >
                                    {label} {status && <ProductStatusBadge className="ml-1" status={status} />}
                                </SidebarNavItem>
                            )
                    )}
                </SidebarGroup>
                {(props.user.organizations.nodes.length > 0 || !siteAdminViewingOtherUser) && (
                    <SidebarGroup>
                        <SidebarGroupHeader label="Your organizations" />
                        {props.user.organizations.nodes.map(org => (
                            <SidebarNavItem
                                key={org.id}
                                to={`/organizations/${org.name}/settings`}
                                className="text-truncate text-nowrap align-items-center"
                                onClick={collapseMobileSidebar}
                            >
                                <OrgAvatar org={org.name} className="d-inline-flex mr-1" /> {org.name}
                            </SidebarNavItem>
                        ))}
                        {!siteAdminViewingOtherUser &&
                            (window.context.sourcegraphDotComMode &&
                            !props.authenticatedUser?.tags?.includes('CreateOrg') ? (
                                <SidebarNavItem
                                    to={`${props.match.path}/about-organizations`}
                                    onClick={collapseMobileSidebar}
                                >
                                    About organizations
                                </SidebarNavItem>
                            ) : (
                                <div className={styles.newOrgBtnWrapper}>
                                    <Button
                                        to="/organizations/new"
                                        variant="secondary"
                                        outline={true}
                                        size="sm"
                                        as={Link}
                                        onClick={collapseMobileSidebar}
                                    >
                                        <Icon aria-hidden={true} svgPath={mdiPlus} /> New organization
                                    </Button>
                                </div>
                            ))}
                    </SidebarGroup>
                )}
                <SidebarGroup>
                    <SidebarGroupHeader label="Other actions" />
                    {!siteAdminViewingOtherUser && (
                        <SidebarNavItem to="/api/console" onClick={collapseMobileSidebar}>
                            API console
                        </SidebarNavItem>
                    )}
                    {props.authenticatedUser.siteAdmin && (
                        <SidebarNavItem to="/site-admin" onClick={collapseMobileSidebar}>
                            Site admin
                        </SidebarNavItem>
                    )}
                </SidebarGroup>
                <div>Version: {window.context.version}</div>
            </div>
        </>
    )
}
