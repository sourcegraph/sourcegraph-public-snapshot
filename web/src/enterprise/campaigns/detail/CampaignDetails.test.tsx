import React from 'react'
import { CampaignDetails } from './CampaignDetails'
import * as H from 'history'
import { createRenderer } from 'react-test-renderer/shallow'

jest.mock('../../../settings/MonacoSettingsEditor', () => ({ MonacoSettingsEditor: 'MonacoSettingsEditor' }))

describe('CampaignDetails', () => {
    test('creation form', () => {
        const history = H.createMemoryHistory()
        expect(
            createRenderer().render(
                <CampaignDetails
                    campaignID={undefined}
                    history={history}
                    location={history.location}
                    authenticatedUser={{ username: 'alice', avatarURL: null }}
                    isLightTheme={true}
                />
            )
        ).toMatchSnapshot()
    })
})
