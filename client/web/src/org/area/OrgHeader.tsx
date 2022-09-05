import React, { useEffect, useMemo } from 'react'

import { gql, useQuery } from '@apollo/client'
import { Location } from 'history'
import { match } from 'react-router'
import { NavLink, RouteComponentProps } from 'react-router-dom'

import { PageHeader, Button, Link, Icon } from '@sourcegraph/wildcard'

import { BatchChangesProps } from '../../batches'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { GetStartedInfoResult, GetStartedInfoVariables } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { NavItemWithIconDescriptor } from '../../util/contributions'
import { useEventBus } from '../emitter'
import { calculateLeftGetStartedSteps, showGetStartPage } from '../openBeta/GettingStarted'
import { OrgAvatar } from '../OrgAvatar'

import { OrgAreaPageProps } from './OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {
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
    newMembersInviteEnabled: boolean
    getStartedInfo: OrgSummary | undefined
}

export interface OrgAreaHeaderNavItem extends NavItemWithIconDescriptor<OrgAreaHeaderContext> {
    isActive?: (match: match | null, location: Location, props: OrgAreaHeaderContext) => boolean
}

const GET_STARTED_INFO_QUERY = gql`
    query GetStartedInfo($organization: ID!) {
        membersSummary: orgMembersSummary(organization: $organization) {
            membersCount
            invitesCount
        }
        extServices: externalServices(namespace: $organization) {
            totalCount
        }
    }
`

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
    newMembersInviteEnabled,
}) => {
    const emitter = useEventBus()

    useEffect(() => {
        const unsubscribe = emitter.subscribe('refreshOrgHeader', () => {
            refetch({ organization: org.id }).catch(() => eventLogger.log('OrgHeaderSummaryrefreshError'))
        })

        return () => {
            emitter.unSubscribe('refreshOrgHeader', unsubscribe)
        }
    })
    const { data, refetch } = useQuery<GetStartedInfoResult, GetStartedInfoVariables>(GET_STARTED_INFO_QUERY, {
        variables: { organization: org.id },
    })

    const [isOpenBetaEnabled] = useFeatureFlag('open-beta-enabled')

    const memoizedNavItems = useMemo(
        (): readonly OrgAreaHeaderNavItem[] => [
            {
                to: '/getstarted',
                label: 'Get started',
                dynamicLabel: ({ getStartedInfo, org }) => calculateLeftGetStartedSteps(getStartedInfo, org.name),
                isActive: (_match, location) => location.pathname.includes('getstarted'),
                condition: ({ getStartedInfo, org, isSourcegraphDotCom }) =>
                    showGetStartPage(getStartedInfo, org.name, isOpenBetaEnabled, isSourcegraphDotCom),
            },
            ...navItems,
        ],
        [navItems, isOpenBetaEnabled]
    )

    const context = {
        batchChangesEnabled,
        batchChangesExecutionEnabled,
        batchChangesWebhookLogsEnabled,
        org,
        isSourcegraphDotCom,
        newMembersInviteEnabled,
        getStartedInfo: data ? (data as OrgSummary) : undefined,
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
                                {memoizedNavItems.map(
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
                        </div>
                    </>
                )}
            </div>
        </div>
    )
}
