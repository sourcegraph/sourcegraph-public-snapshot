import React from 'react'
import { CampaignDiffStat } from './CampaignDiffStat'
import { createRenderer } from 'react-test-renderer/shallow'

describe('CampaignDiffStat', () => {
    test('for campaign', () =>
        expect(
            createRenderer().render(
                <CampaignDiffStat
                    campaign={{
                        __typename: 'Campaign' as const,
                        diffStat: {
                            added: 888,
                            deleted: 777,
                            changed: 999,
                        },
                    }}
                    className="abc"
                />
            )
        ).toMatchSnapshot())
    test('hidden for empty campaign', () =>
        expect(
            createRenderer().render(
                <CampaignDiffStat
                    campaign={{
                        __typename: 'Campaign' as const,
                        diffStat: {
                            added: 0,
                            deleted: 0,
                            changed: 0,
                        },
                    }}
                    className="abc"
                />
            )
        ).toMatchSnapshot())
    test('for patch set', () =>
        expect(
            createRenderer().render(
                <CampaignDiffStat
                    patchSet={{
                        __typename: 'PatchSet' as const,
                        patches: {
                            nodes: [
                                {
                                    diff: {
                                        fileDiffs: {
                                            diffStat: {
                                                added: 888,
                                                changed: 777,
                                                deleted: 999,
                                            },
                                        },
                                    },
                                },
                            ],
                        },
                    }}
                    className="abc"
                />
            )
        ).toMatchSnapshot())
    test('hidden for empty patch set', () =>
        expect(
            createRenderer().render(
                <CampaignDiffStat
                    patchSet={{
                        __typename: 'PatchSet' as const,
                        patches: {
                            nodes: [
                                {
                                    diff: {
                                        fileDiffs: {
                                            diffStat: {
                                                added: 0,
                                                changed: 0,
                                                deleted: 0,
                                            },
                                        },
                                    },
                                },
                            ],
                        },
                    }}
                    className="abc"
                />
            )
        ).toMatchSnapshot())
})
