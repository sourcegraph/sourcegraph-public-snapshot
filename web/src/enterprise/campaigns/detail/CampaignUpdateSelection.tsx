import * as H from 'history'
import React, { useCallback, useMemo } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../components/PageTitle'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../components/FilteredConnection'
import { CampaignNode, CampaignNodeProps, CampaignNodeCampaign } from '../list/CampaignNode'
import { queryCampaigns } from '../global/list/backend'
import { Redirect } from 'react-router'

interface Props {
    history: H.History
    location: H.Location
    className?: string
}

/**
 * A list of a campaign's to choose from for an update
 */
export const CampaignUpdateSelection: React.FunctionComponent<Props> = ({ history, location }) => {
    const patchSetID = useMemo(() => new URLSearchParams(location.search).get('patchSet'), [location.search])
    const selectCampaign = useCallback(
        (campaign: Pick<GQL.ICampaign, 'id'>) => history.push(`/campaigns/${campaign.id}?patchSet=${patchSetID!}`),
        [history, patchSetID]
    )
    // Only query open campaigns, that are non-manual. Those are the only ones that can be updated with a patch set.
    const queryConnection = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            queryCampaigns({ ...args, state: GQL.CampaignState.OPEN, viewerCanAdminister: true }),
        []
    )
    if (!patchSetID) {
        return <Redirect to="/campaigns" />
    }
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
                    selection: { buttonLabel: 'Preview', enabled: true, onSelect: selectCampaign },
                    history,
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
