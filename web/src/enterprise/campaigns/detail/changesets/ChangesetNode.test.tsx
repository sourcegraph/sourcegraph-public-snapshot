import * as H from 'history'
import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { ChangesetNode } from './ChangesetNode'
import {
    IChangesetPlan,
    ChangesetReviewState,
    ChangesetState,
    IExternalChangeset,
    ChangesetCheckState,
} from '../../../../../../shared/src/graphql/schema'
import { Subject } from 'rxjs'

jest.mock('mdi-react/AccountCheckIcon', () => 'AccountCheckIcon')
jest.mock('mdi-react/AccountAlertIcon', () => 'AccountAlertIcon')
jest.mock('mdi-react/AccountQuestionIcon', () => 'AccountQuestionIcon')
jest.mock('mdi-react/SourceMergeIcon', () => 'SourceMergeIcon')
jest.mock('mdi-react/SourcePullIcon', () => 'SourcePullIcon')
jest.mock('mdi-react/DeleteIcon', () => 'DeleteIcon')

describe('ChangesetNode', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    const renderChangesetPlan = (enablePublishing: boolean): void => {
        const renderer = createRenderer()
        renderer.render(
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
                campaignUpdates={new Subject<void>()}
                enablePublishing={enablePublishing}
                _now={new Date('2019-12-15')}
            />
        )
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    }
    test('renders a changesetplan with publishing disabled', () => {
        renderChangesetPlan(false)
    })
    test('renders a changesetplan with publishing enabled', () => {
        renderChangesetPlan(true)
    })
    test('renders an externalchangeset', () => {
        const renderer = createRenderer()
        renderer.render(
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
                        checkState: ChangesetCheckState.FAILED,
                        externalID: '123',
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
                        labels: [
                            {
                                __typename: 'ChangesetLabel',
                                color: '93ba13',
                                description: 'Something is broken',
                                text: 'bug',
                            },
                        ],
                        repository: {
                            __typename: 'Repository',
                            name: 'sourcegraph',
                            url: 'github.com/sourcegraph/sourcegraph',
                        },
                        updatedAt: new Date('2020-01-01').toISOString(),
                    } as IExternalChangeset
                }
                campaignUpdates={new Subject<void>()}
                enablePublishing={false}
                _now={new Date('2019-12-15')}
            />
        )
        const result = renderer.getRenderOutput()
        expect(result.props).toMatchSnapshot()
    })
})
