import * as H from 'history'
import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignPatches } from './CampaignPatches'
import { Subject, of } from 'rxjs'

describe('CampaignPatches', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    test('renders', () => {
        const renderer = createRenderer()
        renderer.render(
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
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    })
})
