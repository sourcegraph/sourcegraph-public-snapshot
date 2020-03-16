import React from 'react'
import * as H from 'history'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignUpdateSelection } from './CampaignUpdateSelection'

describe('CampaignUpdateSelection', () => {
    test('renders', () => {
        const history = H.createMemoryHistory()
        const renderer = createRenderer()
        renderer.render(
            <CampaignUpdateSelection
                history={history}
                location={history.location}
                onSelect={() => undefined}
                className="abc"
            />
        )
        expect(renderer.getRenderOutput()).toMatchSnapshot()
    })
})
