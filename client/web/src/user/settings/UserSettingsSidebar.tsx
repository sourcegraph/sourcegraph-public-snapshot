import AddIcon from 'mdi-react/AddIcon'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'

import { AuthenticatedUser } from '../../auth'
import { BatchChangesProps } from '../../batches'
import { Badge, BadgeStatus } from '../../components/Badge'
import { SidebarGroup, SidebarGroupHeader, SidebarNavItem } from '../../components/Sidebar'
import { UserSettingsAreaUserFields } from '../../graphql-operations'
import { OrgAvatar } from '../../org/OrgAvatar'
import { OnboardingTourProps } from '../../search'
import { useTemporarySetting } from '../../settings/temporary/useTemporarySetting'
import { NavItemDescriptor } from '../../util/contributions'

import { UserSettingsAreaRouteContext } from './UserSettingsArea'

export interface UserSettingsSidebarItemConditionContext extends BatchChangesProps {
    user: UserSettingsAreaUserFields
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'siteAdmin' | 'tags'>
    isSourcegraphDotCom: boolean
}

type UserSettingsSidebarItem = NavItemDescriptor<UserSettingsSidebarItemConditionContext> & {
    status?: BadgeStatus
}

export type UserSettingsSidebarItems = readonly UserSettingsSidebarItem[]

export interface UserSettingsSidebarProps
    extends UserSettingsAreaRouteContext,
        BatchChangesProps,
        OnboardingTourProps,
        RouteComponentProps<{}> {
    items: UserSettingsSidebarItems
    isSourcegraphDotCom: boolean
    className?: string
}

/** Sidebar for user account pages. */
export const UserSettingsSidebar: React.FunctionComponent<UserSettingsSidebarProps> = props => {
    const [, setHasCancelledTour] = useTemporarySetting('search.onboarding.tourCancelled')

    if (!props.authenticatedUser) {
        return null
    }

    // When the site admin is viewing another user's account.
    const siteAdminViewingOtherUser = props.user.id !== props.authenticatedUser.id
    const context: UserSettingsSidebarItemConditionContext = {
        batchChangesEnabled: props.batchChangesEnabled,
        batchChangesExecutionEnabled: props.batchChangesExecutionEnabled,
        user: props.user,
        authenticatedUser: props.authenticatedUser,
        isSourcegraphDotCom: props.isSourcegraphDotCom,
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
                                {label} {status && <Badge className="ml-1" status={status} />}
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
                        >
                            <OrgAvatar org={org.name} className="d-inline-flex mr-1" /> {org.name}
                        </SidebarNavItem>
                    ))}
                    {!siteAdminViewingOtherUser &&
                        (window.context.sourcegraphDotComMode ? (
                            <SidebarNavItem to={`${props.match.path}/about-organizations`}>
                                About organizations
                            </SidebarNavItem>
                        ) : (
                            <div className="user-settings-sidebar__new-org-btn-wrapper">
                                <Link to="/organizations/new" className="btn btn-outline-secondary btn-sm">
                                    <AddIcon className="icon-inline" /> New organization
                                </Link>
                            </div>
                        ))}
                </SidebarGroup>
            )}
            <SidebarGroup>
                <SidebarGroupHeader label="Other actions" />
                {!siteAdminViewingOtherUser && <SidebarNavItem to="/api/console">API console</SidebarNavItem>}
                {props.authenticatedUser.siteAdmin && <SidebarNavItem to="/site-admin">Site admin</SidebarNavItem>}
                {props.showOnboardingTour && (
                    <button
                        type="button"
                        className="btn text-left sidebar__link--inactive d-flex w-100"
                        onClick={reEnableSearchTour}
                    >
                        Show search tour
                    </button>
                )}
            </SidebarGroup>
            <div>Version: {window.context.version}</div>
        </div>
    )
}
