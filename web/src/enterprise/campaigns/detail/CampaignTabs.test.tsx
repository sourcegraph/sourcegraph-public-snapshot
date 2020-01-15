import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignTabs } from './CampaignTabs'
import * as H from 'history'
import { createRenderer } from 'react-test-renderer/shallow'
import { Subject } from 'rxjs'

jest.mock('./changesets/CampaignChangesets', () => ({ CampaignChangesets: 'CampaignChangesets' }))
jest.mock('./diffs/CampaignDiffs', () => ({ CampaignDiffs: 'CampaignDiffs' }))

const history = H.createMemoryHistory()

describe('CampaignTabs', () => {
    test('renders', () =>
        expect(
            createRenderer().render(
                <CampaignTabs
                    campaign={
                        {
                            id: '0',
                            changesets: {
                                nodes: [] as GQL.IExternalChangeset[],
                                totalCount: 0,
                            } as GQL.IExternalChangesetConnection,
                            changesetPlans: {
                                nodes: [] as GQL.IChangesetPlan[],
                                totalCount: 0,
                            } as GQL.IChangesetPlanConnection,
                        } as GQL.ICampaign
                    }
                    persistLines={true}
                    history={history}
                    location={history.location}
                    isLightTheme={true}
                    campaignUpdates={new Subject<void>()}
                    changesetUpdates={new Subject<void>()}
                />
            )
        ).toMatchSnapshot())
})
