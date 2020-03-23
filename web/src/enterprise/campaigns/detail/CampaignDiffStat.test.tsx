import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignDiffStat } from './CampaignDiffStat'
import { createRenderer } from 'react-test-renderer/shallow'

describe('CampaignDiffStat', () => {
    test('for campaign', () =>
        expect(
            createRenderer().render(
                <CampaignDiffStat
                    campaign={{
                        __typename: 'Campaign' as const,
                        changesets: {
                            nodes: [
                                {
                                    diff: {
                                        fileDiffs: {
                                            diffStat: {
                                                added: 10,
                                                changed: 10,
                                                deleted: 10,
                                                __typename: 'DiffStat' as const,
                                            },
                                        } as GQL.IFileDiffConnection,
                                    } as GQL.IRepositoryComparison,
                                } as GQL.IExternalChangeset,
                            ],
                        },
                        patches: {
                            nodes: [
                                {
                                    diff: {
                                        fileDiffs: {
                                            diffStat: {
                                                added: 10,
                                                changed: 10,
                                                deleted: 10,
                                                __typename: 'DiffStat' as const,
                                            },
                                        } as GQL.IPreviewFileDiffConnection,
                                    } as GQL.IPreviewRepositoryComparison,
                                } as GQL.IPatch,
                            ],
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
                        changesets: {
                            nodes: [
                                {
                                    diff: {
                                        fileDiffs: {
                                            diffStat: {
                                                added: 0,
                                                changed: 0,
                                                deleted: 0,
                                                __typename: 'DiffStat' as const,
                                            },
                                        } as GQL.IFileDiffConnection,
                                    } as GQL.IRepositoryComparison,
                                } as GQL.IExternalChangeset,
                            ],
                        },
                        patches: {
                            nodes: [
                                {
                                    diff: {
                                        fileDiffs: {
                                            diffStat: {
                                                added: 0,
                                                changed: 0,
                                                deleted: 0,
                                                __typename: 'DiffStat' as const,
                                            },
                                        } as GQL.IPreviewFileDiffConnection,
                                    } as GQL.IPreviewRepositoryComparison,
                                } as GQL.IPatch,
                            ],
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
                    campaign={{
                        __typename: 'PatchSet' as const,
                        patches: {
                            nodes: [
                                {
                                    diff: {
                                        fileDiffs: {
                                            diffStat: {
                                                added: 10,
                                                changed: 10,
                                                deleted: 10,
                                                __typename: 'DiffStat' as const,
                                            },
                                        } as GQL.IPreviewFileDiffConnection,
                                    } as GQL.IPreviewRepositoryComparison,
                                } as GQL.IPatch,
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
                    campaign={{
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
                                                __typename: 'DiffStat' as const,
                                            },
                                        } as GQL.IPreviewFileDiffConnection,
                                    } as GQL.IPreviewRepositoryComparison,
                                } as GQL.IPatch,
                            ],
                        },
                    }}
                    className="abc"
                />
            )
        ).toMatchSnapshot())
})
