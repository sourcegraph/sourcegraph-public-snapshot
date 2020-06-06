import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignDescription } from './CampaignDescription'
import * as H from 'history'

describe('CampaignDescription', () => {
    test('', () =>
        expect(
            createRenderer().render(
                <CampaignDescription
                    campaign={{
                        author: { avatarURL: null, username: 'alice' },
                        createdAt: '2020-01-01',
                        description: '**a** b',
                    }}
                    history={H.createMemoryHistory()}
                />
            )
        ).toMatchSnapshot())
})
