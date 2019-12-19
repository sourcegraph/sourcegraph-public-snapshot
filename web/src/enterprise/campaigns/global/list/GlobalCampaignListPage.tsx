import React from 'react'
import { queryCampaigns } from './backend'
import AddIcon from 'mdi-react/AddIcon'
import { Link } from '../../../../../../shared/src/components/Link'
import { RouteComponentProps } from 'react-router'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { ICampaign } from '../../../../../../shared/src/graphql/schema'
import { CampaignNode } from '../../list/CampaignNode'

interface Props extends Pick<RouteComponentProps, 'history' | 'location'> {}

/**
 * A list of all campaigns on the Sourcegraph instance.
 */
export const GlobalCampaignListPage: React.FunctionComponent<Props> = props => (
    <>
        <h1>Campaigns</h1>
        <p>Perform and track large-scale code changes</p>

        <div className="text-right mb-1">
            <Link to="/campaigns/new" className="btn btn-primary">
                <AddIcon className="icon-inline" /> New campaign
            </Link>
        </div>

        <FilteredConnection<ICampaign>
            {...props}
            nodeComponent={CampaignNode}
            queryConnection={queryCampaigns}
            hideSearch={true}
            noun="campaign"
            pluralNoun="campaigns"
        />
    </>
)
