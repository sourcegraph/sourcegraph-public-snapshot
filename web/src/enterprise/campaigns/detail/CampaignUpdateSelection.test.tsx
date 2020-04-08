import React from 'react'
import * as H from 'history'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignUpdateSelection } from './CampaignUpdateSelection'

describe('CampaignUpdateSelection', () => {
    test('renders with URL param', () => {
        const history = H.createMemoryHistory({ initialEntries: ['/campaigns/update?patchSet=123'] })
        const renderer = createRenderer()
        renderer.render(<CampaignUpdateSelection history={history} location={history.location} className="abc" />)
        expect(renderer.getRenderOutput()).toMatchSnapshot()
    })
    test('redirects without URL param', () => {
        const history = H.createMemoryHistory({ initialEntries: ['/campaigns/update'] })
        const renderer = createRenderer()
        renderer.render(<CampaignUpdateSelection history={history} location={history.location} className="abc" />)
        expect(renderer.getRenderOutput()).toMatchSnapshot()
    })
})
