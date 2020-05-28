import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignCliHelp } from './CampaignCliHelp'
import { registerHighlightContributions } from '../../../../../../shared/src/highlight/contributions'

// This is idempotent, so calling it in multiple tests is not a problem.
registerHighlightContributions()

describe('CampaignCliHelp', () => {
    test('renders', () => {
        const result = renderer.create(<CampaignCliHelp className="test" />)
        expect(result.toJSON()).toMatchSnapshot()
    })
})
