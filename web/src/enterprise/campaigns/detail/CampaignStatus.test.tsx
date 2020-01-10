import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignStatus } from './CampaignStatus'
import { createRenderer } from 'react-test-renderer/shallow'

const PROPS = {
    onRetry: () => undefined,
}

const CAMPAIGN: Pick<GQL.ICampaign, '__typename' | 'closedAt' | 'publishedAt'> = {
    __typename: 'Campaign',
    closedAt: null,
    publishedAt: null,
}

const CAMPAIGN_PLAN: Pick<GQL.ICampaignPlan, '__typename'> = {
    __typename: 'CampaignPlan',
}

describe('CampaignStatus', () => {
    test('closed campaign', () =>
        expect(
            createRenderer().render(
                <CampaignStatus
                    {...PROPS}
                    campaign={{ ...CAMPAIGN, closedAt: '2020-01-01' }}
                    status={{
                        completedCount: 1,
                        pendingCount: 0,
                        errors: [],
                        state: GQL.BackgroundProcessState.COMPLETED,
                    }}
                />
            )
        ).toMatchSnapshot())

    test('campaign processing', () =>
        expect(
            createRenderer().render(
                <CampaignStatus
                    {...PROPS}
                    campaign={CAMPAIGN}
                    status={{
                        completedCount: 3,
                        pendingCount: 3,
                        errors: ['a', 'b'],
                        state: GQL.BackgroundProcessState.PROCESSING,
                    }}
                />
            )
        ).toMatchSnapshot())

    test('campaign plan processing', () =>
        expect(
            createRenderer().render(
                <CampaignStatus
                    {...PROPS}
                    campaign={CAMPAIGN_PLAN}
                    status={{
                        completedCount: 3,
                        pendingCount: 3,
                        errors: ['a', 'b'],
                        state: GQL.BackgroundProcessState.PROCESSING,
                    }}
                />
            )
        ).toMatchSnapshot())

    test('campaign errored', () =>
        expect(
            createRenderer().render(
                <CampaignStatus
                    {...PROPS}
                    campaign={CAMPAIGN}
                    status={{
                        completedCount: 3,
                        pendingCount: 0,
                        errors: ['a', 'b'],
                        state: GQL.BackgroundProcessState.ERRORED,
                    }}
                />
            )
        ).toMatchSnapshot())
})
