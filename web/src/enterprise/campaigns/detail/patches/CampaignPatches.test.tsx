import * as H from 'history'
import React from 'react'
import { CampaignPatches } from './CampaignPatches'
import { Subject, of } from 'rxjs'
import { shallow } from 'enzyme'

describe('CampaignPatches', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    test('renders', () => {
        expect(
            shallow(
                <CampaignPatches
                    isLightTheme={true}
                    history={history}
                    location={location}
                    campaign={{ id: 'test' }}
                    campaignUpdates={new Subject<void>()}
                    changesetUpdates={new Subject<void>()}
                    enablePublishing={false}
                    queryPatches={() =>
                        of({
                            __typename: 'PatchConnection',
                            nodes: [],
                            totalCount: 0,
                            pageInfo: { __typename: 'PageInfo', endCursor: null, hasNextPage: false },
                        })
                    }
                />
            )
        ).toMatchSnapshot()
    })
})
