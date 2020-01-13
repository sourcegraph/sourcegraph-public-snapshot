import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CampaignChangesets } from './CampaignChangesets'
import { createRenderer } from 'react-test-renderer/shallow'
import * as H from 'history'
import { of } from 'rxjs'

describe('CampaignChangesets', () => {
    const history = H.createMemoryHistory()
    test('renders', () =>
        expect(
            createRenderer().render(
                <CampaignChangesets
                    queryChangesetsConnection={() =>
                        of({ nodes: [{ id: '0' } as GQL.IExternalChangeset] } as GQL.IExternalChangesetConnection)
                    }
                    history={history}
                    location={history.location}
                    isLightTheme={true}
                />
            )
        ).toMatchSnapshot())
})
