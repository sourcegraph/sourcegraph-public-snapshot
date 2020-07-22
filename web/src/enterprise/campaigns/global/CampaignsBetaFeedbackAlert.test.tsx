import React from 'react'
import { CampaignsBetaFeedbackAlert } from './CampaignsBetaFeedbackAlert'
import { shallow } from 'enzyme'

describe('CampaignsBetaFeedbackAlert', () => {
    test('renders', () => {
        expect(shallow(<CampaignsBetaFeedbackAlert />)).toMatchSnapshot()
    })
})
