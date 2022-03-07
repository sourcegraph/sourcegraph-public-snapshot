import { Location } from 'history'
import React from 'react'
import { match } from 'react-router'
import { NavLink, RouteComponentProps } from 'react-router-dom'

import { PageHeader, Button, Link } from '@sourcegraph/wildcard'

import { BatchChangesProps } from '../../batches'
import { NavItemWithIconDescriptor } from '../../util/contributions'
import { OrgAvatar } from '../OrgAvatar'

import { OrgAreaPageProps } from './OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {
    isSourcegraphDotCom: boolean
    navItems: readonly OrgAreaHeaderNavItem[]
    className?: string
}

export interface OrgAreaHeaderContext extends BatchChangesProps, Pick<Props, 'org'> {
    isSourcegraphDotCom: boolean
    newMembersInviteEnabled: boolean
}

export interface OrgAreaHeaderNavItem extends NavItemWithIconDescriptor<OrgAreaHeaderContext> {
    isActive?: (match: match | null, location: Location, props: OrgAreaHeaderContext) => boolean
}

/**
 * Header for the organization area.
 */
export const OrgHeader: React.FunctionComponent<Props> = ({
    batchChangesEnabled,
    batchChangesExecutionEnabled,
    batchChangesWebhookLogsEnabled,
    org,
    navItems,
    match,
    className = '',
    isSourcegraphDotCom,
    newMembersInviteEnabled,
}) => {
    const context = {
        batchChangesEnabled,
        batchChangesExecutionEnabled,
        batchChangesWebhookLogsEnabled,
        org,
        isSourcegraphDotCom,
        newMembersInviteEnabled,
    }

    return (
        <div className={className}>
            <div className="container">
                {org && (
                    <>
                        <PageHeader
                            path={[
                                {
                                    icon: () => <OrgAvatar org={org.name} size="lg" className="mr-3" />,
                                    text: (
                                        <span className="align-middle">
                                            {org.displayName ? (
                                                <>
                                                    {org.displayName} ({org.name})
                                                </>
                                            ) : (
                                                org.name
                                            )}
                                        </span>
                                    ),
                                },
                            ]}
                            className="mb-3"
                        />
                        <div className="d-flex align-items-end justify-content-between">
                            <ul className="nav nav-tabs w-100">
                                {navItems.map(
                                    ({ to, label, exact, icon: Icon, condition = () => true, isActive }) =>
                                        condition(context) && (
                                            <li key={label} className="nav-item">
                                                <NavLink
                                                    to={match.url + to}
                                                    className="nav-link"
                                                    activeClassName="active"
                                                    exact={exact}
                                                    isActive={
                                                        isActive
                                                            ? (match, location) => isActive(match, location, context)
                                                            : undefined
                                                    }
                                                >
                                                    <span>
                                                        {Icon && <Icon className="icon-inline" />}{' '}
                                                        <span className="text-content" data-tab-content={label}>
                                                            {label}
                                                        </span>
                                                    </span>
                                                </NavLink>
                                            </li>
                                        )
                                )}
                            </ul>
                            <div className="flex-1" />
                            {org.viewerPendingInvitation?.respondURL && (
                                <div className="pb-1">
                                    <small className="mr-2">Join organization:</small>
                                    <Button
                                        to={org.viewerPendingInvitation.respondURL}
                                        variant="success"
                                        size="sm"
                                        as={Link}
                                    >
                                        View invitation
                                    </Button>
                                </div>
                            )}
                        </div>
                    </>
                )}
            </div>
        </div>
    )
}
