import React from 'react'

import { Location } from 'history'
import { match } from 'react-router'
import { NavLink, RouteComponentProps } from 'react-router-dom'

import { PageHeader, Button, Link, Icon } from '@sourcegraph/wildcard'

import { BatchChangesProps } from '../../batches'
import { NavItemWithIconDescriptor } from '../../util/contributions'
import { OrgAvatar } from '../OrgAvatar'

import { OrgAreaRouteContext } from './OrgArea'

interface Props extends OrgAreaRouteContext, RouteComponentProps<{}> {
    isSourcegraphDotCom: boolean
    navItems: readonly OrgAreaHeaderNavItem[]
    className?: string
}

export interface OrgSummary {
    membersSummary: { membersCount: number; invitesCount: number }
    extServices: { totalCount: number }
}

export interface OrgAreaHeaderContext extends BatchChangesProps, Pick<Props, 'org'> {
    isSourcegraphDotCom: boolean
}

export interface OrgAreaHeaderNavItem extends NavItemWithIconDescriptor<OrgAreaHeaderContext> {
    isActive?: (match: match | null, location: Location, props: OrgAreaHeaderContext) => boolean
}

/**
 * Header for the organization area.
 */
export const OrgHeader: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    batchChangesEnabled,
    batchChangesExecutionEnabled,
    batchChangesWebhookLogsEnabled,
    org,
    navItems,
    match,
    className = '',
    isSourcegraphDotCom,
}) => {
    const context: OrgAreaHeaderContext = {
        batchChangesEnabled,
        batchChangesExecutionEnabled,
        batchChangesWebhookLogsEnabled,
        org,
        isSourcegraphDotCom,
    }

    return (
        <div className={className}>
            <div className="container">
                {org && (
                    <>
                        <PageHeader className="mb-3">
                            <PageHeader.Heading as="h2" styleAs="h1">
                                <PageHeader.Breadcrumb
                                    icon={() => <OrgAvatar org={org.name} size="lg" className="mr-3" />}
                                >
                                    <span className="align-middle">
                                        {org.displayName ? (
                                            <>
                                                {org.displayName} ({org.name})
                                            </>
                                        ) : (
                                            org.name
                                        )}
                                    </span>
                                </PageHeader.Breadcrumb>
                            </PageHeader.Heading>
                        </PageHeader>
                        <nav className="d-flex align-items-end justify-content-between" aria-label="Org">
                            <ul className="nav nav-tabs w-100">
                                {navItems.map(
                                    ({
                                        to,
                                        label,
                                        exact,
                                        icon: ItemIcon,
                                        condition = () => true,
                                        isActive,
                                        dynamicLabel,
                                    }) =>
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
                                                        {ItemIcon && <Icon as={ItemIcon} aria-hidden={true} />}{' '}
                                                        <span className="text-content" data-tab-content={label}>
                                                            {dynamicLabel ? dynamicLabel(context) : label}
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
                        </nav>
                    </>
                )}
            </div>
        </div>
    )
}
