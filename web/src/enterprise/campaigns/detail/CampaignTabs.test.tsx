import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignTabs } from './CampaignTabs'
import * as H from 'history'
import { createRenderer } from 'react-test-renderer/shallow'

jest.mock('./changesets/CampaignChangesets', () => ({ CampaignChangesets: 'CampaignChangesets' }))
jest.mock('./diffs/CampaignDiffs', () => ({ CampaignDiffs: 'CampaignDiffs' }))

const history = H.createMemoryHistory()

describe('CampaignTabs', () => {
    test('renders', () =>
        expect(
            createRenderer().render(
                <CampaignTabs
                    changesets={{
                        nodes: [
                            {
                                __typename: 'ChangesetPlan' as const,
                                id: '0',
                                repository: {
                                    url: 'github.com/sourcegraph/sourcegraph',
                                    name: 'sourcegraph/sourcegraph',
                                } as GQL.IRepository,
                                diff: {
                                    __typename: 'PreviewRepositoryComparison',
                                    fileDiffs: {
                                        __typename: 'PreviewFileDiffConnection',
                                        nodes: [] as GQL.IPreviewFileDiff[],
                                        diffStat: {
                                            __typename: 'DiffStat' as const,
                                            added: 1,
                                            changed: 2,
                                            deleted: 3,
                                        },
                                    } as GQL.IPreviewFileDiffConnection,
                                } as GQL.IPreviewRepositoryComparison,
                            } as GQL.IChangesetPlan,
                        ],
                        totalCount: 1,
                    }}
                    persistLines={true}
                    history={history}
                    location={history.location}
                    isLightTheme={true}
                />
            )
        ).toMatchSnapshot())
})
