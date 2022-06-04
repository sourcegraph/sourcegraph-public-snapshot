import { storiesOf } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import { GET_BATCH_CHANGE_TO_EDIT } from '../../create/backend'
import {
    ACTIVE_EXECUTORS_MOCK,
    mockBatchChange,
    mockBatchSpec,
    NO_ACTIVE_EXECUTORS_MOCK,
    UNSTARTED_CONNECTION_MOCKS,
    UNSTARTED_WITH_CACHE_CONNECTION_MOCKS,
} from '../batch-spec.mock'
import { insertNameIntoLibraryItem } from '../yaml-util'

import { EditBatchSpecPage } from './EditBatchSpecPage'
import goImportsSample from './library/go-imports.batch.yaml'

const { add } = storiesOf('web/batches/batch-spec/edit/EditBatchSpecPage', module).addDecorator(story => (
    <div className="p-3" style={{ height: '95vh', width: '100%' }}>
        {story()}
    </div>
))

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

const FIRST_TIME_MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GET_BATCH_CHANGE_TO_EDIT),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { batchChange: mockBatchChange() } },
        nMatches: Number.POSITIVE_INFINITY,
    },
    ACTIVE_EXECUTORS_MOCK,
    ...UNSTARTED_CONNECTION_MOCKS,
])

add('editing for the first time', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={FIRST_TIME_MOCKS}>
                <EditBatchSpecPage
                    {...props}
                    batchChange={{
                        name: 'my-batch-change',
                        namespace: 'test1234',
                    }}
                    settingsCascade={SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))

const MULTIPLE_SPEC_MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GET_BATCH_CHANGE_TO_EDIT),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: {
            data: {
                batchChange: mockBatchChange({
                    batchSpecs: {
                        nodes: [
                            mockBatchSpec({
                                id: 'new',
                                originalInput: insertNameIntoLibraryItem(goImportsSample, 'my-batch-change'),
                            }),
                            mockBatchSpec({ id: 'old1' }),
                            mockBatchSpec({ id: 'old2' }),
                        ],
                    },
                }),
            },
        },
        nMatches: Number.POSITIVE_INFINITY,
    },
    ACTIVE_EXECUTORS_MOCK,
    ...UNSTARTED_WITH_CACHE_CONNECTION_MOCKS,
])

add('editing the latest batch spec', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={MULTIPLE_SPEC_MOCKS}>
                <EditBatchSpecPage
                    {...props}
                    batchChange={{
                        name: 'my-batch-change',
                        namespace: 'test1234',
                    }}
                    settingsCascade={SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('batch change not found', () => (
    <WebStory>
        {props => (
            <EditBatchSpecPage
                {...props}
                batchChange={{
                    name: 'doesnt-exist',
                    namespace: 'test1234',
                }}
                settingsCascade={SETTINGS_CASCADE}
            />
        )}
    </WebStory>
))

const NO_EXECUTORS_MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GET_BATCH_CHANGE_TO_EDIT),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { batchChange: mockBatchChange() } },
        nMatches: Number.POSITIVE_INFINITY,
    },
    NO_ACTIVE_EXECUTORS_MOCK,
    ...UNSTARTED_CONNECTION_MOCKS,
])

add('executors not active', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={NO_EXECUTORS_MOCKS}>
                <EditBatchSpecPage
                    {...props}
                    batchChange={{
                        name: 'my-batch-change',
                        namespace: 'test1234',
                    }}
                    settingsCascade={SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))
