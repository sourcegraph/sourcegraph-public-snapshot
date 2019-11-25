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
import { setLinkComponent } from '../../../../../../shared/src/components/Link'

describe('ChangesetNode', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    beforeEach(() => {
        setLinkComponent((props: any) => <a {...props} />)
        afterAll(() => setLinkComponent(null as any)) // reset global env for other tests
    })
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
