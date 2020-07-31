import React, { useEffect, useCallback } from 'react'
import { queryCampaigns as _queryCampaigns } from './backend'
import { RouteComponentProps } from 'react-router'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArgs,
} from '../../../../components/FilteredConnection'
import { CampaignState, IUser } from '../../../../../../shared/src/graphql/schema'
import { CampaignNode, CampaignNodeProps } from '../../list/CampaignNode'
import { TelemetryProps } from '../../../../../../shared/src/telemetry/telemetryService'
import { ListCampaign } from '../../../../graphql-operations'

interface Props extends TelemetryProps, Pick<RouteComponentProps, 'history' | 'location'> {
    authenticatedUser: Pick<IUser, 'siteAdmin'>
    queryCampaigns?: typeof _queryCampaigns
}

const FILTERS: FilteredConnectionFilter[] = [
    {
        label: 'All',
        id: 'all',
        tooltip: 'Show all campaigns',
        args: {},
    },
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
]

/**
 * A list of all campaigns on the Sourcegraph instance.
 */
export const GlobalCampaignListPage: React.FunctionComponent<Props> = ({
    queryCampaigns = _queryCampaigns,
    ...props
}) => {
    const queryConnection = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            queryCampaigns({
                first: args.first ?? null,
                // The types for FilteredConnectionQueryArgs don't allow access to the filter arguments.
                state: (args as { state: CampaignState | undefined }).state ?? null,
                viewerCanAdminister: null,
            }),
        [queryCampaigns]
    )
    useEffect(() => props.telemetryService.logViewEvent('CampaignsListPage'), [props.telemetryService])
    return (
        <>
            <div className="d-flex justify-content-between align-items-end mb-3">
                <div>
                    <h1 className="mb-2">
                        Campaigns{' '}
                        <sup>
                            <span className="badge badge-info text-uppercase">Beta</span>
                        </sup>
                    </h1>
                    <p className="mb-0">
                        Perform and track large-scale code changes.{' '}
                        <a href="https://docs.sourcegraph.com/user/campaigns">Learn how.</a>
                    </p>
                </div>
            </div>

            <div className="card mt-4 mb-4">
                <div className="card-body p-3">
                    <h3>Welcome to campaigns!</h3>
                    <p className="mb-1">
                        We're excited for you to use campaigns to remove legacy code, fix critical security issues, pay
                        down tech debt, and more. We look forward to hearing about campaigns you run inside your
                        organization. Take a look at some{' '}
                        <a href="https://docs.sourcegraph.com/user/campaigns/examples">examples in our documentation</a>
                        , and <a href="mailto:feedback@sourcegraph.com?subject=Campaigns feedback">get in touch</a> with
                        any questions or feedback!
                    </p>
                </div>
            </div>

            <FilteredConnection<ListCampaign, Omit<CampaignNodeProps, 'node'>>
                {...props}
                nodeComponent={CampaignNode}
                nodeComponentProps={{ history: props.history }}
                queryConnection={queryConnection}
                hideSearch={true}
                filters={FILTERS}
                noun="campaign"
                pluralNoun="campaigns"
                className="mb-3"
            />
        </>
    )
}
