import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignStatusBadge } from './CampaignStatusBadge'

describe('CampaignStatusBadge', () => {
    test('open', () =>
        expect(createRenderer().render(<CampaignStatusBadge campaign={{ closedAt: null }} />)).toMatchSnapshot())

    test('closed', () =>
        expect(
            createRenderer().render(<CampaignStatusBadge campaign={{ closedAt: '2020-01-01' }} />)
        ).toMatchSnapshot())
})
