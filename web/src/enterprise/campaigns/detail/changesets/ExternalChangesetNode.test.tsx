import * as H from 'history'
import React from 'react'
import { ExternalChangesetNode } from './ExternalChangesetNode'
import {
    ChangesetReviewState,
    ChangesetCheckState,
    ChangesetExternalState,
    ChangesetReconcilerState,
    ChangesetPublicationState,
} from '../../../../../../shared/src/graphql/schema'
import { Subject } from 'rxjs'
import { shallow } from 'enzyme'

jest.mock('mdi-react/AccountCheckIcon', () => 'AccountCheckIcon')
jest.mock('mdi-react/AccountAlertIcon', () => 'AccountAlertIcon')
jest.mock('mdi-react/AccountQuestionIcon', () => 'AccountQuestionIcon')
jest.mock('mdi-react/SourceMergeIcon', () => 'SourceMergeIcon')
jest.mock('mdi-react/SourcePullIcon', () => 'SourcePullIcon')
jest.mock('mdi-react/DeleteIcon', () => 'DeleteIcon')

describe('ExternalChangesetNode', () => {
    const history = H.createMemoryHistory({ keyLength: 0 })
    const location = H.createLocation('/campaigns')
    test('renders an externalchangeset', () => {
        expect(
            shallow(
                <ExternalChangesetNode
                    isLightTheme={true}
                    history={history}
                    location={location}
                    viewerCanAdminister={true}
                    node={{
                        __typename: 'ExternalChangeset',
                        id: 'TestExternalChangeset',
                        reviewState: ChangesetReviewState.PENDING,
                        publicationState: ChangesetPublicationState.PUBLISHED,
                        reconcilerState: ChangesetReconcilerState.COMPLETED,
                        externalState: ChangesetExternalState.OPEN,
                        externalURL: {
                            url: 'https://github.com/sourcegraph/sourcegraph/pull/111111',
                        },
                        title: 'Remove lodash',
                        body: 'We should remove lodash',
                        checkState: ChangesetCheckState.FAILED,
                        externalID: '123',
                        diffStat: {
                            added: 100,
                            changed: 200,
                            deleted: 100,
                        },
                        labels: [
                            {
                                color: '93ba13',
                                description: 'Something is broken',
                                text: 'bug',
                            },
                        ],
                        repository: {
                            name: 'sourcegraph',
                            url: 'github.com/sourcegraph/sourcegraph',
                            id: 'TestRepository',
                        },
                        createdAt: new Date('2020-01-01').toISOString(),
                        updatedAt: new Date('2020-01-01').toISOString(),
                        nextSyncAt: null,
                    }}
                    campaignUpdates={new Subject<void>()}
                />
            )
        ).toMatchSnapshot()
    })
})
