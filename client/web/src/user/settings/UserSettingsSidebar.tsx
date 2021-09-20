import AddIcon from 'mdi-react/AddIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'

import { AuthenticatedUser } from '../../auth'
import { BatchChangesProps } from '../../batches'
import { Badge, BadgeStatus } from '../../components/Badge'
import { SidebarGroup, SidebarGroupHeader, SidebarNavItem } from '../../components/Sidebar'
import { UserAreaUserFields } from '../../graphql-operations'
import { OrgAvatar } from '../../org/OrgAvatar'
import { OnboardingTourProps } from '../../search'
import { HAS_CANCELLED_TOUR_KEY, HAS_COMPLETED_TOUR_KEY } from '../../search/input/SearchOnboardingTour'
import { NavItemDescriptor } from '../../util/contributions'
import { UserAreaRouteContext } from '../area/UserArea'

export interface UserSettingsSidebarItemConditionContext extends BatchChangesProps {
    user: UserAreaUserFields
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'siteAdmin' | 'tags'>
    isSourcegraphDotCom: boolean
}

type UserSettingsSidebarItem = NavItemDescriptor<UserSettingsSidebarItemConditionContext> & {
    status?: BadgeStatus
}

export type UserSettingsSidebarItems = readonly UserSettingsSidebarItem[]

export interface UserSettingsSidebarProps
    extends UserAreaRouteContext,
        BatchChangesProps,
        OnboardingTourProps,
        RouteComponentProps<{}> {
    items: UserSettingsSidebarItems
    isSourcegraphDotCom: boolean
    className?: string
}

function reEnableSearchTour(): void {
    localStorage.setItem(HAS_CANCELLED_TOUR_KEY, 'false')
    localStorage.setItem(HAS_COMPLETED_TOUR_KEY, 'false')
}

/** Sidebar for user account pages. */
export const UserSettingsSidebar: React.FunctionComponent<UserSettingsSidebarProps> = props => {
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

    return (
        <div className={props.className}>
            <SidebarGroup>
                <SidebarGroupHeader label="User account" />
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
                    <SidebarGroupHeader label="Organizations" />
                    {props.user.organizations.nodes.map(org => (
                        <SidebarNavItem
                            key={org.id}
                            to={`/organizations/${org.name}/settings`}
                            className="text-truncate text-nowrap align-items-center"
                        >
                            <OrgAvatar org={org.name} className="d-inline-flex mr-1" /> {org.name}
                        </SidebarNavItem>
                    ))}
                    {!siteAdminViewingOtherUser && (
                        <div className="user-settings-sidebar__new-org-btn-wrapper">
                            {!window.context.sourcegraphDotComMode ? (
                                <Link to="/organizations/new" className="btn btn-outline-secondary btn-sm">
                                    <AddIcon className="icon-inline" /> New organization
                                </Link>
                            ) : (
                                <a
                                    href="https://docs.sourcegraph.com/code_search/explanations/sourcegraph_cloud"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    Learn More <ExternalLinkIcon className="icon-inline" />
                                </a>
                            )}
                        </div>
                    )}
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
