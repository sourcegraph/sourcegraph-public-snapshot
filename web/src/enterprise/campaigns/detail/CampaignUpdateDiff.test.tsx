import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { of } from 'rxjs'
import { CampaignUpdateDiff } from './CampaignUpdateDiff'

describe('CampaignUpdateDiff', () => {
    test('renders', () => {
        const history = H.createMemoryHistory({ keyLength: 0 })
        const location = H.createLocation('/campaigns/Q2FtcGFpZ25QbGFuOjE4Mw%3D%3D?plan=Q2FtcGFpZ25QbGFuOjE4Mw%3D%3D')
        expect(
            renderer
                .create(
                    <CampaignUpdateDiff
                        isLightTheme={true}
                        history={history}
                        location={location}
                        campaign={{
                            id: 'somecampaign',
                            publishedAt: null,
                            changesets: { totalCount: 1 },
                            changesetPlans: { totalCount: 1 },
                        }}
                        campaignPlan={{ id: 'someothercampaign', changesetPlans: { totalCount: 1 } }}
                        _queryChangesets={() =>
                            of({
                                nodes: [{ __typename: 'ExternalChangeset', repository: { id: 'match1' } }],
                            }) as any
                        }
                        _queryChangesetPlans={() =>
                            of({
                                nodes: [{ __typename: 'ChangesetPlan', repository: { id: 'match1' } }],
                            }) as any
                        }
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
