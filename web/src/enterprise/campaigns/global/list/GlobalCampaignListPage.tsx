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

/**
 * A list of all campaigns on the Sourcegraph instance.
 */
export const GlobalCampaignListPage: React.FunctionComponent<Props> = props => (
    <>
        <div className="d-flex justify-content-between align-items-end mb-3">
            <div>
                <h1 className="mb-2">Campaigns</h1>
                <p className="mb-0">Perform and track large-scale code changes</p>
            </div>
            {props.authenticatedUser.siteAdmin && (
                <Link to="/campaigns/create" className="btn btn-primary ml-3">
                    <AddIcon className="icon-inline" /> New campaign
                </Link>
            )}
        </div>

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
    </>
)
