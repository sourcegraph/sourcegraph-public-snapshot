import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { FileDiffTab } from './FileDiffTab'
import * as GQL from '../../../../../shared/src/graphql/schema'

describe('FileDiffTab', () => {
    test('renders the form', () => {
        const history = H.createMemoryHistory({ keyLength: 0 })
        const location = H.createLocation('/campaigns/new')
        expect(
            renderer
                .create(
                    <FileDiffTab
                        isLightTheme={true}
                        history={history}
                        location={location}
                        nodes={[
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
                        ]}
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
