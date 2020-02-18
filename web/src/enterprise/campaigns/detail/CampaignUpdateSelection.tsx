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
 * A list of a campaign's to choose from for an update
 */
export const CampaignUpdateSelection: React.FunctionComponent<Props> = ({ history, location, onSelect }) => (
    <>
        <PageTitle title="Update campaign" />
        <h1>Select campaign to update</h1>
        <p>Choose a campaign to update from the list below to preview which changes will be made to the code hosts:</p>
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
                selection: { buttonLabel: 'Preview', enabled: true, onSelect },
            }}
            queryConnection={queryCampaigns}
            useURLQuery={false}
            hideSearch={true}
            noun="campaign"
            pluralNoun="campaigns"
        />
    </>
)
