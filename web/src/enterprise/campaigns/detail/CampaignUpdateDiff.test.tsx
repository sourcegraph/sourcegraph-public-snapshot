import React from 'react'
import { CampaignUpdateDiff } from './CampaignUpdateDiff'
import { mount } from 'enzyme'

describe('CampaignUpdateDiff', () => {
    test('renders', () => {
        expect(
            mount(
                <CampaignUpdateDiff
                    isLightTheme={true}
                    campaign={{
                        id: 'somecampaign',
                        changesets: { totalCount: 1 },
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot()
    })
})
