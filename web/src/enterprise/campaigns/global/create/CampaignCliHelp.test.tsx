import React from 'react'
import { CampaignCliHelp } from './CampaignCliHelp'
import { registerHighlightContributions } from '../../../../../../shared/src/highlight/contributions'
import { mount } from 'enzyme'

// This is idempotent, so calling it in multiple tests is not a problem.
registerHighlightContributions()

describe('CampaignCliHelp', () => {
    test('renders', () => {
        expect(mount(<CampaignCliHelp className="test" />)).toMatchSnapshot()
    })
})
