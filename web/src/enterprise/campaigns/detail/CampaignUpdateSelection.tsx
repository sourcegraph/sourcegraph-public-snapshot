import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../components/PageTitle'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../components/FilteredConnection'
import { CampaignNode, CampaignNodeProps, CampaignNodeCampaign } from '../list/CampaignNode'
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
export const CampaignUpdateSelection: React.FunctionComponent<Props> = ({ history, location, onSelect }) => {
    // Only query open campaigns, that are non-manual. Those are the only ones that can be updated with a patch set.
    const queryConnection = React.useCallback(
        (args: FilteredConnectionQueryArgs) =>
            queryCampaigns({ ...args, state: GQL.CampaignState.OPEN, hasPatchSet: true }),
        []
    )
    return (
        <>
            <PageTitle title="Update campaign" />
            <h1>Select campaign to update</h1>
            <p>
                Choose the campaign you want to update and preview which changes will be made to the changesets on the
                code hosts:
            </p>
            <FilteredConnection<CampaignNodeCampaign, Omit<CampaignNodeProps, 'node'>>
                history={history}
                location={location}
                nodeComponent={CampaignNode}
                nodeComponentProps={{
                    selection: { buttonLabel: 'Preview', enabled: true, onSelect },
                }}
                queryConnection={queryConnection}
                useURLQuery={false}
                hideSearch={true}
                noun="campaign"
                pluralNoun="campaigns"
            />
        </>
    )
}
