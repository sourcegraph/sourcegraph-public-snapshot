import React from 'react'
import { CampaignPreview } from './CampaignPreview'
import { createRenderer } from 'react-test-renderer/shallow'

describe('CampaignPreview', () => {
    test('renders', () => expect(createRenderer().render(<CampaignPreview />)).toMatchSnapshot())
})
