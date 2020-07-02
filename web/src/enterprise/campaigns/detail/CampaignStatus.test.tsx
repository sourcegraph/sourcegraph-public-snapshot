import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignStatus } from './CampaignStatus'
import { createMemoryHistory } from 'history'
import { shallow } from 'enzyme'

const PROPS = {
    afterRetry: () => undefined,
    afterPublish: () => undefined,
    history: createMemoryHistory(),
}

const CAMPAIGN: Pick<GQL.ICampaign, 'id' | 'closedAt' | 'viewerCanAdminister'> & {
    changesets: Pick<GQL.ICampaign['changesets'], 'totalCount'>
} = {
    id: 'Q2FtcGFpZ246MQ==',
    closedAt: null,
    viewerCanAdminister: true,
    changesets: {
        totalCount: 0,
    },
}

describe('CampaignStatus', () => {
    for (const viewerCanAdminister of [true, false]) {
        const campaign = { ...CAMPAIGN, viewerCanAdminister }
        describe(`viewerCanAdminister: ${String(viewerCanAdminister)}`, () => {
            test('closed campaign', () =>
                expect(
                    shallow(
                        <CampaignStatus
                            {...PROPS}
                            campaign={{
                                ...campaign,
                                closedAt: '2020-01-01',
                                status: {
                                    completedCount: 1,
                                    pendingCount: 0,
                                    errors: [],
                                    state: GQL.BackgroundProcessState.COMPLETED,
                                },
                            }}
                        />
                    )
                ).toMatchSnapshot())

            test('drafted campaign', () =>
                expect(
                    shallow(
                        <CampaignStatus
                            {...PROPS}
                            campaign={{
                                ...campaign,
                                status: {
                                    completedCount: 1,
                                    pendingCount: 0,
                                    errors: [],
                                    state: GQL.BackgroundProcessState.COMPLETED,
                                },
                            }}
                        />
                    )
                ).toMatchSnapshot())

            test('drafted campaign, some published', () =>
                expect(
                    shallow(
                        <CampaignStatus
                            {...PROPS}
                            campaign={{
                                ...campaign,
                                changesets: { totalCount: 1 },
                                status: {
                                    completedCount: 1,
                                    pendingCount: 0,
                                    errors: [],
                                    state: GQL.BackgroundProcessState.COMPLETED,
                                },
                            }}
                        />
                    )
                ).toMatchSnapshot())

            test('campaign processing', () =>
                expect(
                    shallow(
                        <CampaignStatus
                            {...PROPS}
                            campaign={{
                                ...campaign,
                                status: {
                                    completedCount: 3,
                                    pendingCount: 3,
                                    errors: ['a', 'b'],
                                    state: GQL.BackgroundProcessState.PROCESSING,
                                },
                            }}
                        />
                    )
                ).toMatchSnapshot())

            test('campaign errored', () =>
                expect(
                    shallow(
                        <CampaignStatus
                            {...PROPS}
                            campaign={{
                                ...campaign,
                                status: {
                                    completedCount: 3,
                                    pendingCount: 0,
                                    errors: ['a', 'b'],
                                    state: GQL.BackgroundProcessState.ERRORED,
                                },
                            }}
                        />
                    )
                ).toMatchSnapshot())
        })
    }
})
