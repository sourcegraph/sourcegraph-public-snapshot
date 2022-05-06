import { storiesOf } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'

import { WebStory } from '../../../../components/WebStory'

import { EditBatchSpecPage } from './EditBatchSpecPage'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { GET_BATCH_CHANGE_TO_EDIT } from '../../create/backend'

import { mockBatchChange } from '../batch-spec.mock'

const { add } = storiesOf('web/batches/batch-spec/edit/EditBatchSpecPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            disableSnapshot: false,
        },
    })

const FIXTURE_ORG: SettingsOrgSubject = {
    __typename: 'Org',
    name: 'sourcegraph',
    displayName: 'Sourcegraph',
    id: 'a',
    viewerCanAdminister: true,
}

const FIXTURE_USER: SettingsUserSubject = {
    __typename: 'User',
    username: 'alice',
    displayName: 'alice',
    id: 'b',
    viewerCanAdminister: true,
}

const SETTINGS_CASCADE = {
    ...EMPTY_SETTINGS_CASCADE,
    subjects: [
        { subject: FIXTURE_ORG, settings: { a: 1 }, lastID: 1 },
        { subject: FIXTURE_USER, settings: { b: 2 }, lastID: 2 },
    ],
}

const mocks = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GET_BATCH_CHANGE_TO_EDIT),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { batchChange: mockBatchChange() } },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

add('editing for the first time', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={mocks}>
                <div style={{ height: '95vh', width: '100%' }}>
                    <EditBatchSpecPage
                        {...props}
                        batchChange={{
                            name: 'doesnt-exist',
                            url: 'some-url',
                            namespace: { name: 'my-cool-org', id: 'test1234', url: 'some-url' },
                        }}
                        settingsCascade={SETTINGS_CASCADE}
                    />
                </div>
            </MockedTestProvider>
        )}
    </WebStory>
))

add('batch change not found', () => (
    <WebStory>
        {props => (
            <div style={{ height: '95vh', width: '100%' }}>
                <EditBatchSpecPage
                    {...props}
                    batchChange={{
                        name: 'doesnt-exist',
                        url: 'some-url',
                        namespace: { name: 'my-cool-org', id: 'test1234', url: 'some-url' },
                    }}
                    settingsCascade={SETTINGS_CASCADE}
                />
            </div>
        )}
    </WebStory>
))
