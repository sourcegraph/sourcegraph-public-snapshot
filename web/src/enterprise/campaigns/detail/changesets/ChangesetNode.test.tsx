import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { ChangesetNode } from './ChangesetNode'
import {
    IChangesetPlan,
    ChangesetReviewState,
    ChangesetState,
    IExternalChangeset,
} from '../../../../../../shared/src/graphql/schema'

jest.mock('mdi-react/AccountCheckIcon', () => 'AccountCheckIcon')
jest.mock('mdi-react/AccountAlertIcon', () => 'AccountAlertIcon')
jest.mock('mdi-react/AccountQuestionIcon', () => 'AccountQuestionIcon')
jest.mock('mdi-react/SourceMergeIcon', () => 'SourceMergeIcon')
jest.mock('mdi-react/SourcePullIcon', () => 'SourcePullIcon')
jest.mock('mdi-react/DeleteIcon', () => 'DeleteIcon')

describe('ChangesetNode', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    test('renders a changesetplan', () => {
        expect(
            renderer
                .create(
                    <ChangesetNode
                        isLightTheme={true}
                        history={history}
                        location={location}
                        node={
                            {
                                __typename: 'ChangesetPlan',
                                diff: {
                                    fileDiffs: {
                                        __typename: 'PreviewFileDiffConnection',
                                        diffStat: {
                                            added: 100,
                                            changed: 200,
                                            deleted: 100,
                                        },
                                    },
                                },
                                repository: {
                                    __typename: 'Repository',
                                    name: 'sourcegraph',
                                    url: 'github.com/sourcegraph/sourcegraph',
                                },
                            } as IChangesetPlan
                        }
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
    test('renders an externalchangeset', () => {
        expect(
            renderer
                .create(
                    <ChangesetNode
                        isLightTheme={true}
                        history={history}
                        location={location}
                        node={
                            {
                                __typename: 'ExternalChangeset',
                                reviewState: ChangesetReviewState.PENDING,
                                state: ChangesetState.OPEN,
                                externalURL: {
                                    url: 'https://github.com/sourcegraph/sourcegraph/pull/111111',
                                },
                                title: 'Remove lodash',
                                body: 'We should remove lodash',
                                diff: {
                                    fileDiffs: {
                                        diffStat: {
                                            added: 100,
                                            changed: 200,
                                            deleted: 100,
                                        },
                                        nodes: [{ __typename: 'FileDiff' }],
                                    },
                                },
                                repository: {
                                    __typename: 'Repository',
                                    name: 'sourcegraph',
                                    url: 'github.com/sourcegraph/sourcegraph',
                                },
                            } as IExternalChangeset
                        }
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
