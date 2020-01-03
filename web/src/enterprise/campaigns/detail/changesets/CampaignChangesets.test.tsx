import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CampaignChangesets } from './CampaignChangesets'
import { createRenderer } from 'react-test-renderer/shallow'
import * as H from 'history'

describe('CampaignChangesets', () => {
    const history = H.createMemoryHistory()
    test('renders', () =>
        expect(
            createRenderer().render(
                <CampaignChangesets
                    changesets={{ nodes: [{ id: '0' } as GQL.IExternalChangeset] }}
                    history={history}
                    location={history.location}
                    isLightTheme={true}
                />
            )
        ).toMatchSnapshot())
})
