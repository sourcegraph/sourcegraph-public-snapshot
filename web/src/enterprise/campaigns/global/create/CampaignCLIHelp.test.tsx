import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignCLIHelp } from './CampaignCLIHelp'
import { registerHighlightContributions } from '../../../../../../shared/src/highlight/contributions'

// This is idempotent, so calling it in multiple tests is not a problem.
registerHighlightContributions()

describe('CampaignCLIHelp', () => {
    test('renders', () => {
        const result = renderer.create(<CampaignCLIHelp className="test" />)
        expect(result.toJSON()).toMatchSnapshot()
    })
})
