import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignStateBadge } from './CampaignStateBadge'
import * as H from 'history'

describe('CampaignStateBadge', () => {
    test('', () =>
        expect(
            createRenderer().render(
                <CampaignStateBadge
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
