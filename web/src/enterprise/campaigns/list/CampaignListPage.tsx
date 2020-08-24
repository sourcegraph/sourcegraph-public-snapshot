import React, { useEffect, useCallback, useMemo } from 'react'
import { queryCampaigns as _queryCampaigns, queryCampaignsByUser, queryCampaignsByOrg } from './backend'
import { RouteComponentProps } from 'react-router'
import { FilteredConnection, FilteredConnectionFilter } from '../../../components/FilteredConnection'
import { CampaignNode, CampaignNodeProps } from './CampaignNode'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import {
    ListCampaign,
    CampaignState,
    Scalars,
    CampaignsByUserVariables,
    CampaignsByOrgVariables,
} from '../../../graphql-operations'
import { CampaignsListBetaNotice } from './CampaignsListBetaNotice'
import PlusIcon from 'mdi-react/PlusIcon'
import { Link } from '../../../../../shared/src/components/Link'
import { PageHeader } from '../../../components/PageHeader'
import { CampaignsIcon } from '../icons'
import { Breadcrumbs, BreadcrumbSetters, BreadcrumbsProps } from '../../../components/Breadcrumbs'

interface Props
    extends TelemetryProps,
        Pick<RouteComponentProps, 'history' | 'location'>,
        BreadcrumbsProps,
        BreadcrumbSetters {
    displayNamespace?: boolean
    queryCampaigns?: typeof _queryCampaigns
}

const FILTERS: FilteredConnectionFilter[] = [
    {
        label: 'Open',
        id: 'open',
        tooltip: 'Show only campaigns that are open',
        args: { state: CampaignState.OPEN },
    },
    {
        label: 'Closed',
        id: 'closed',
        tooltip: 'Show only campaigns that are closed',
        args: { state: CampaignState.CLOSED },
    },
    {
        label: 'All',
        id: 'all',
        tooltip: 'Show all campaigns',
        args: {},
    },
]

/**
 * A list of all campaigns on the Sourcegraph instance.
 */
export const CampaignListPage: React.FunctionComponent<Props> = ({
    queryCampaigns = _queryCampaigns,
    displayNamespace = true,
    location,
    ...props
}) => {
    useEffect(() => props.telemetryService.logViewEvent('CampaignsListPage'), [props.telemetryService])
    props.useBreadcrumb(
        useMemo(
            () => ({
                key: 'Campaigns list',
                element: <>Campaigns</>,
            }),
            []
        )
    )
    return (
        <>
            <Breadcrumbs breadcrumbs={props.breadcrumbs} />
            <PageHeader
                icon={CampaignsIcon}
                title={
                    <span className="test-campaign-list-page">
                        Campaigns{' '}
                        <sup>
                            <span className="badge badge-merged text-uppercase">Beta</span>
                        </sup>
                    </span>
                }
                actions={
                    <Link to={`${location.pathname}/create`} className="btn btn-primary">
                        <PlusIcon className="icon-inline" /> New campaign
                    </Link>
                }
            />
            <CampaignsListBetaNotice />
            <FilteredConnection<ListCampaign, Omit<CampaignNodeProps, 'node'>>
                {...props}
                location={location}
                nodeComponent={CampaignNode}
                nodeComponentProps={{ history: props.history, displayNamespace }}
                queryConnection={queryCampaigns}
                hideSearch={true}
                defaultFirst={15}
                filters={FILTERS}
                noun="campaign"
                pluralNoun="campaigns"
                listComponent="div"
                listClassName="campaign-list-page__grid mb-3"
                className="mb-3"
                cursorPaging={true}
            />
        </>
    )
}

export interface UserCampaignListPageProps extends Props {
    userID: Scalars['ID']
}

/**
 * A list of all campaigns in a users namespace.
 */
export const UserCampaignListPage: React.FunctionComponent<UserCampaignListPageProps> = ({ userID, ...props }) => {
    const queryConnection = useCallback(
        (args: Partial<CampaignsByUserVariables>) =>
            queryCampaignsByUser({
                userID,
                first: args.first ?? null,
                after: args.after ?? null,
                // The types for FilteredConnectionQueryArgs don't allow access to the filter arguments.
                state: (args as { state: CampaignState | undefined }).state ?? null,
                viewerCanAdminister: null,
            }),
        [userID]
    )
    return <CampaignListPage {...props} displayNamespace={false} queryCampaigns={queryConnection} />
}

export interface OrgCampaignListPageProps extends Props {
    orgID: Scalars['ID']
}

/**
 * A list of all campaigns in an orgs namespace.
 */
export const OrgCampaignListPage: React.FunctionComponent<OrgCampaignListPageProps> = ({ orgID, ...props }) => {
    const queryConnection = useCallback(
        (args: Partial<CampaignsByOrgVariables>) =>
            queryCampaignsByOrg({
                orgID,
                first: args.first ?? null,
                after: args.after ?? null,
                // The types for FilteredConnectionQueryArgs don't allow access to the filter arguments.
                state: (args as { state: CampaignState | undefined }).state ?? null,
                viewerCanAdminister: null,
            }),
        [orgID]
    )
    return <CampaignListPage {...props} displayNamespace={false} queryCampaigns={queryConnection} />
}
