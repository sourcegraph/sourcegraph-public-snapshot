import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignUpdateSelection } from './CampaignUpdateSelection'

describe('CampaignUpdateSelection', () => {
    test('renders', () => {
        const renderer = createRenderer()
        renderer.render(
            <CampaignUpdateSelection
                history={undefined as any}
                location={undefined as any}
                onSelect={() => undefined}
                className="abc"
            />
        )
        expect(renderer.getRenderOutput()).toMatchSnapshot()
    })
})
