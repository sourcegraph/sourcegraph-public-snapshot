import type { Meta, Decorator, StoryFn } from '@storybook/react'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import { GET_CHANGESETS_BY_IDS_QUERY } from '../backend'

import { ExportChangesetsModal } from './ExportChangesetsModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/ExportChangesetsModal',
    decorators: [decorator],
}

export default config

const mocks = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GET_CHANGESETS_BY_IDS_QUERY),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: {
            data: {
                getChangesetsByIDs: {
                    nodes: [
                        {
                            __typename: 'ExternalChangeset',
                            id: 'test-123',
                            title: 'Test Changeset',
                            state: 'OPEN',
                            reviewState: 'PENDING',
                            externalURL: {
                                url: 'https://github.com/sourcegraph/sourcegraph/pull/1',
                            },
                            repository: {
                                name: 'github.com/sourcegraph/sourcegraph',
                            },
                        },
                    ],
                },
            },
        },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

export const Confirmation: StoryFn = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={mocks}>
                <ExportChangesetsModal
                    {...props}
                    afterCreate={noop}
                    batchChangeID="test-123"
                    changesetIDs={['test-123']}
                    onCancel={noop}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)
