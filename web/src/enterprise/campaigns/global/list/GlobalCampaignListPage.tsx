import React from 'react'
import { queryCampaigns } from './backend'
import AddIcon from 'mdi-react/AddIcon'
import { Link } from '../../../../../../shared/src/components/Link'
import { RouteComponentProps } from 'react-router'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { ICampaign, IUser } from '../../../../../../shared/src/graphql/schema'
import { CampaignNode } from '../../list/CampaignNode'

interface Props extends Pick<RouteComponentProps, 'history' | 'location'> {
    authenticatedUser: IUser
}

/**
 * A list of all campaigns on the Sourcegraph instance.
 */
export const GlobalCampaignListPage: React.FunctionComponent<Props> = props => (
    <>
        <h1>Campaigns</h1>
        <p>Perform and track large-scale code changes</p>

        {props.authenticatedUser.siteAdmin && (
            <div className="text-right mb-1">
                <Link to="/campaigns/new" className="btn btn-primary">
                    <AddIcon className="icon-inline" /> New campaign
                </Link>
            </div>
        )}

        <FilteredConnection<
            Pick<
                ICampaign,
                'id' | 'plan' | 'closedAt' | 'name' | 'description' | 'changesets' | 'changesetPlans' | 'createdAt'
            >
        >
            {...props}
            nodeComponent={CampaignNode}
            queryConnection={queryCampaigns}
            hideSearch={true}
            noun="campaign"
            pluralNoun="campaigns"
        />
    </>
)
