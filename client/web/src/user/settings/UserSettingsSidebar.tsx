import * as React from 'react'

import AddIcon from 'mdi-react/AddIcon'
import { RouteComponentProps } from 'react-router-dom'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { ProductStatusBadge, Button, Link, Icon, ProductStatusType } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BatchChangesProps } from '../../batches'
import { SidebarGroup, SidebarGroupHeader, SidebarNavItem } from '../../components/Sidebar'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { UserSettingsAreaUserFields } from '../../graphql-operations'
import { OrgAvatar } from '../../org/OrgAvatar'
import { useExperimentalFeatures } from '../../stores'
import { NavItemDescriptor } from '../../util/contributions'

import { UserSettingsAreaRouteContext } from './UserSettingsArea'

import styles from './UserSettingsSidebar.module.scss'

export interface UserSettingsSidebarItemConditionContext extends BatchChangesProps {
    user: UserSettingsAreaUserFields
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'siteAdmin' | 'tags'>
    isSourcegraphDotCom: boolean
    openBetaEnabled: boolean
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
    const [, setHasCancelledTour] = useTemporarySetting('search.onboarding.tourCancelled')
    const showOnboardingTour = useExperimentalFeatures(features => features.showOnboardingTour ?? false)
    const [isOpenBetaEnabled] = useFeatureFlag('open-beta-enabled')

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
        openBetaEnabled: isOpenBetaEnabled,
    }

    function reEnableSearchTour(): void {
        setHasCancelledTour(false)
    }

    return (
        <div className={props.className}>
            <SidebarGroup>
                <SidebarGroupHeader label="Account" />
                {props.items.map(
                    ({ label, to, exact, status, condition = () => true }) =>
                        condition(context) && (
                            <SidebarNavItem key={label} to={props.match.path + to} exact={exact}>
                                {label} {status && <ProductStatusBadge className="ml-1" status={status} />}
                            </SidebarNavItem>
                        )
                )}
            </SidebarGroup>
            {!isOpenBetaEnabled && (props.user.organizations.nodes.length > 0 || !siteAdminViewingOtherUser) && (
                <SidebarGroup>
                    <SidebarGroupHeader label="Your organizations" />
                    {props.user.organizations.nodes.map(org => (
                        <SidebarNavItem
                            key={org.id}
                            to={`/organizations/${org.name}/settings`}
                            className="text-truncate text-nowrap align-items-center"
                        >
                            <OrgAvatar org={org.name} className="d-inline-flex mr-1" /> {org.name}
                        </SidebarNavItem>
                    ))}
                    {!siteAdminViewingOtherUser &&
                        (window.context.sourcegraphDotComMode &&
                        !props.authenticatedUser?.tags?.includes('CreateOrg') ? (
                            <SidebarNavItem to={`${props.match.path}/about-organizations`}>
                                About organizations
                            </SidebarNavItem>
                        ) : (
                            <div className={styles.newOrgBtnWrapper}>
                                <Button to="/organizations/new" variant="secondary" outline={true} size="sm" as={Link}>
                                    <Icon as={AddIcon} /> New organization
                                </Button>
                            </div>
                        ))}
                </SidebarGroup>
            )}
            <SidebarGroup>
                <SidebarGroupHeader label="Other actions" />
                {!siteAdminViewingOtherUser && <SidebarNavItem to="/api/console">API console</SidebarNavItem>}
                {props.authenticatedUser.siteAdmin && <SidebarNavItem to="/site-admin">Site admin</SidebarNavItem>}
                {showOnboardingTour && (
                    <Button className="text-left sidebar__link--inactive d-flex w-100" onClick={reEnableSearchTour}>
                        Show search tour
                    </Button>
                )}
            </SidebarGroup>
            <div>Version: {window.context.version}</div>
        </div>
    )
}
