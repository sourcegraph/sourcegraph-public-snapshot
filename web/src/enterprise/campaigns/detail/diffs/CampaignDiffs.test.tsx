import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CampaignDiffs } from './CampaignDiffs'
import { createRenderer } from 'react-test-renderer/shallow'
import * as H from 'history'

describe('CampaignDiffs', () => {
    const history = H.createMemoryHistory()
    test('renders', () =>
        expect(
            createRenderer().render(
                <CampaignDiffs
                    changesets={{ nodes: [{ id: '0' } as GQL.IExternalChangeset] }}
                    persistLines={true}
                    history={history}
                    location={history.location}
                    isLightTheme={true}
                />
            )
        ).toMatchSnapshot())
})
