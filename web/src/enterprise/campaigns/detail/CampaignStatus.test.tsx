import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignStatus } from './CampaignStatus'
import { createRenderer } from 'react-test-renderer/shallow'

const PROPS = {
    onRetry: () => undefined,
}

const CAMPAIGN: Pick<GQL.ICampaign, '__typename' | 'closedAt' | 'publishedAt' | 'changesets'> = {
    __typename: 'Campaign',
    closedAt: null,
    publishedAt: '2020-01-01',
    changesets: {
        totalCount: 0,
    } as GQL.IExternalChangesetConnection,
}

describe('CampaignStatus', () => {
    for (const viewerCanAdminister of [true, false]) {
        const campaign = { ...CAMPAIGN, viewerCanAdminister }
        describe(`viewerCanAdminister: ${String(viewerCanAdminister)}`, () => {
            test('closed campaign', () =>
                expect(
                    createRenderer().render(
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
                            onPublish={() => undefined}
                        />
                    )
                ).toMatchSnapshot())

            test('drafted campaign', () =>
                expect(
                    createRenderer().render(
                        <CampaignStatus
                            {...PROPS}
                            campaign={{
                                ...campaign,
                                publishedAt: null,
                                status: {
                                    completedCount: 1,
                                    pendingCount: 0,
                                    errors: [],
                                    state: GQL.BackgroundProcessState.COMPLETED,
                                },
                            }}
                            onPublish={() => undefined}
                        />
                    )
                ).toMatchSnapshot())

            test('drafted campaign, some published', () =>
                expect(
                    createRenderer().render(
                        <CampaignStatus
                            {...PROPS}
                            campaign={{
                                ...campaign,
                                publishedAt: null,
                                changesets: { totalCount: 1 } as GQL.IExternalChangesetConnection,
                                status: {
                                    completedCount: 1,
                                    pendingCount: 0,
                                    errors: [],
                                    state: GQL.BackgroundProcessState.COMPLETED,
                                },
                            }}
                            onPublish={() => undefined}
                        />
                    )
                ).toMatchSnapshot())

            test('campaign processing', () =>
                expect(
                    createRenderer().render(
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
                            onPublish={() => undefined}
                        />
                    )
                ).toMatchSnapshot())

            test('campaign errored', () =>
                expect(
                    createRenderer().render(
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
                            onPublish={() => undefined}
                        />
                    )
                ).toMatchSnapshot())
        })
    }
})
