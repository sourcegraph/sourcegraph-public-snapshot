import React from 'react'
import renderer from 'react-test-renderer'
import { CreateCampaign } from './CreateCampaign'

describe('CreateCampaign', () => {
    test('renders', () => {
        const result = renderer.create(<CreateCampaign />)
        expect(result.toJSON()).toMatchSnapshot()
    })
})
