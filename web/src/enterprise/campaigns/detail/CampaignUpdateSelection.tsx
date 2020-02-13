import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../components/PageTitle'
import { FilteredConnection } from '../../../components/FilteredConnection'
import { CampaignNode, CampaignNodeProps } from '../list/CampaignNode'
import { queryCampaigns } from '../global/list/backend'

interface Props {
    onSelect: (campaign: Pick<GQL.ICampaign, 'id'>) => void
    history: H.History
    location: H.Location
    className?: string
}

/**
 * A list of a campaign's changesets changed over a new plan
 */
export const CampaignUpdateSelection: React.FunctionComponent<Props> = ({ history, location, onSelect }) => (
    <>
        <PageTitle title="Update campaign" />
        <h1>Update campaign</h1>
        <p>Select the campaign to update from the list of campaigns below:</p>
        <FilteredConnection<
            Pick<
                GQL.ICampaign,
                'id' | 'closedAt' | 'name' | 'description' | 'changesets' | 'changesetPlans' | 'createdAt'
            >,
            Omit<CampaignNodeProps, 'node'>
        >
            history={history}
            location={location}
            nodeComponent={CampaignNode}
            nodeComponentProps={{
                selection: { buttonLabel: 'Select', enabled: true, onSelect },
            }}
            queryConnection={queryCampaigns}
            useURLQuery={false}
            hideSearch={true}
            noun="campaign"
            pluralNoun="campaigns"
        />
    </>
)
