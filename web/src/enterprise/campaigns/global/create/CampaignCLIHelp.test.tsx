import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignCLIHelp } from './CampaignCLIHelp'
import { of } from 'rxjs'

describe('CampaignCLIHelp', () => {
    test('renders', () => {
        const result = renderer.create(<CampaignCLIHelp isLightTheme={true} highlightCode={({ code }) => of(code)} />)
        expect(result.toJSON()).toMatchSnapshot()
    })
})
