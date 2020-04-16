import React from 'react'
import { queryCampaigns } from './backend'
import AddIcon from 'mdi-react/AddIcon'
import { Link } from '../../../../../../shared/src/components/Link'
import { RouteComponentProps } from 'react-router'
import { FilteredConnection, FilteredConnectionFilter } from '../../../../components/FilteredConnection'
import { IUser, CampaignState } from '../../../../../../shared/src/graphql/schema'
import { CampaignNode, CampaignNodeCampaign } from '../../list/CampaignNode'

interface Props extends Pick<RouteComponentProps, 'history' | 'location'> {
    authenticatedUser: IUser
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

// HELP! Need to get the total number of campaigns to use below. This was just my workaround.
const totalCount = 0

/**
 * A list of all campaigns on the Sourcegraph instance.
 */
export const GlobalCampaignListPage: React.FunctionComponent<Props> = props => (
    <>
        <div className="d-flex justify-content-between align-items-end mb-3">
            <div>
                <h1 className="mb-2">
                    Campaigns <span className="badge badge-info badge-outline">Beta</span>
                </h1>
                <p className="mb-0">
                    Perform and track large-scale code changes.{' '}
                    <a href="https://docs.sourcegraph.com/user/campaigns">Learn how.</a>
                </p>
            </div>
            {props.authenticatedUser.siteAdmin && (
                <Link to="/campaigns/create" className="btn btn-primary ml-3">
                    <AddIcon className="icon-inline" /> New campaign
                </Link>
            )}
        </div>

        {totalCount > 0 ? (
            <FilteredConnection<CampaignNodeCampaign>
                {...props}
                nodeComponent={CampaignNode}
                queryConnection={queryCampaigns}
                hideSearch={true}
                filters={FILTERS}
                noun="campaign"
                pluralNoun="campaigns"
                className="mb-3"
            />
        ) : (
            <div className="card mt-4 mb-4">
                <div className="card-body p-3">
                    <h3>
                        Welcome to Campaigns <span className="badge badge-info badge-outline">Beta</span>!
                    </h3>
                    <p className="mb-1">
                        We're excited for you to get started using Campaigns to remove legacy code, fix critical
                        security issues, pay down tech debt, and more! Take a look at some{' '}
                        <a href="https://docs.sourcegraph.com/user/campaigns/examples">examples in our documentation</a>
                        , and don't hesitate to reach out with any questions. We look forward to hearing about campaigns
                        you run inside your organization!
                    </p>
                </div>
            </div>
        )}
    </>
)
