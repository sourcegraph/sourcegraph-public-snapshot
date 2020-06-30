import * as H from 'history'
import React from 'react'
import { PatchSetPatches } from './PatchSetPatches'
import { Subject, of } from 'rxjs'
import { shallow } from 'enzyme'

describe('PatchSetPatches', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    test('renders', () => {
        expect(
            shallow(
                <PatchSetPatches
                    isLightTheme={true}
                    history={history}
                    location={location}
                    patchSet={{ id: 'test' }}
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
