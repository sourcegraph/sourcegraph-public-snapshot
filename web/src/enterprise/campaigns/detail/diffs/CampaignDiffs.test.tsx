import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { CampaignDiffs } from './CampaignDiffs'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { of, Subject } from 'rxjs'

jest.mock('mdi-react/SourcePullIcon', () => 'SourcePullIcon')

describe('CampaignDiffs', () => {
    test('renders', () => {
        const history = H.createMemoryHistory({ keyLength: 0 })
        const location = H.createLocation('/campaigns/new')
        expect(
            renderer
                .create(
                    <CampaignDiffs
                        isLightTheme={true}
                        history={history}
                        location={location}
                        persistLines={true}
                        queryChangesetsConnection={() =>
                            of({
                                __typename: 'ChangesetPlanConnection' as const,
                                totalCount: 1,
                                nodes: [
                                    {
                                        __typename: 'ChangesetPlan' as const,
                                        repository: {
                                            url: 'github.com/sourcegraph/sourcegraph',
                                            name: 'sourcegraph/sourcegraph',
                                        } as GQL.IRepository,
                                        diff: {
                                            __typename: 'PreviewRepositoryComparison',
                                            fileDiffs: {
                                                __typename: 'PreviewFileDiffConnection',
                                                nodes: [] as GQL.IPreviewFileDiff[],
                                            } as GQL.IPreviewFileDiffConnection,
                                        } as GQL.IPreviewRepositoryComparison,
                                    } as GQL.IChangesetPlan,
                                ],
                            } as GQL.IChangesetPlanConnection)
                        }
                        changesetUpdates={new Subject<void>()}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
