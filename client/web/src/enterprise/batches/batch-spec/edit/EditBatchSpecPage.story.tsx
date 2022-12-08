import { DecoratorFn, Meta, Story } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { AuthenticatedUser } from '../../../../auth'
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

const decorator: DecoratorFn = story => (
    <div className="p-3" style={{ height: '95vh', width: '100%' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/batches/batch-spec/edit/EditBatchSpecPage',
    decorators: [decorator],
}

export default config

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
        result: {
            data: {
                batchChange: mockBatchChange({
                    batchSpecs: {
                        nodes: [
                            mockBatchSpec({
                                originalInput: insertNameIntoLibraryItem(goImportsSample, 'my-batch-change'),
                            }),
                        ],
                    },
                }),
            },
        },
        nMatches: Number.POSITIVE_INFINITY,
    },
    ACTIVE_EXECUTORS_MOCK,
    ...UNSTARTED_CONNECTION_MOCKS,
])

const MOCK_ORGANIZATION = {
    __typename: 'Org',
    name: 'acme-corp',
    displayName: 'ACME Corporation',
    id: 'acme-corp-id',
}

const mockAuthenticatedUser = {
    __typename: 'User',
    username: 'alice',
    displayName: 'alice',
    id: 'b',
    organizations: {
        nodes: [MOCK_ORGANIZATION],
    },
} as AuthenticatedUser

export const EditFirstTime: Story = () => (
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
                    authenticatedUser={mockAuthenticatedUser}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

EditFirstTime.storyName = 'editing for the first time'

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

export const EditLatestBatchSpec: Story = () => (
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
                    authenticatedUser={mockAuthenticatedUser}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

EditLatestBatchSpec.storyName = 'editing the latest batch spec'

const NOT_FOUND_MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GET_BATCH_CHANGE_TO_EDIT),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { batchChange: null } },
        nMatches: Number.POSITIVE_INFINITY,
    },
    ACTIVE_EXECUTORS_MOCK,
    ...UNSTARTED_CONNECTION_MOCKS,
])

export const BatchChangeNotFound: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={NOT_FOUND_MOCKS}>
                <EditBatchSpecPage
                    {...props}
                    batchChange={{
                        name: 'doesnt-exist',
                        namespace: 'test1234',
                    }}
                    settingsCascade={SETTINGS_CASCADE}
                    authenticatedUser={mockAuthenticatedUser}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

const INVALID_SPEC_MOCKS = new WildcardMockLink([
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

BatchChangeNotFound.storyName = 'batch change not found'

export const InvalidBatchSpec: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={INVALID_SPEC_MOCKS}>
                <EditBatchSpecPage
                    {...props}
                    batchChange={{
                        name: 'hello-world',
                        namespace: 'test1234',
                    }}
                    settingsCascade={SETTINGS_CASCADE}
                    authenticatedUser={mockAuthenticatedUser}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

InvalidBatchSpec.storyName = 'invalid batch spec'

const NO_EXECUTORS_MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GET_BATCH_CHANGE_TO_EDIT),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { batchChange: mockBatchChange({ name: 'hello-world' }) } },
        nMatches: Number.POSITIVE_INFINITY,
    },
    NO_ACTIVE_EXECUTORS_MOCK,
    ...UNSTARTED_CONNECTION_MOCKS,
])

export const ExecutorsNotActive: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={NO_EXECUTORS_MOCKS}>
                <EditBatchSpecPage
                    {...props}
                    batchChange={{
                        name: 'hello-world',
                        namespace: 'test1234',
                    }}
                    settingsCascade={SETTINGS_CASCADE}
                    authenticatedUser={mockAuthenticatedUser}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ExecutorsNotActive.storyName = 'executors not active'
