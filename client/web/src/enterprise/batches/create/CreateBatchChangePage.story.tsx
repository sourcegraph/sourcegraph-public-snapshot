import { DecoratorFn, Meta, Story } from '@storybook/react'

import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'

import { AuthenticatedUser } from '../../../auth'
import { WebStory } from '../../../components/WebStory'

import { CreateBatchChangePage } from './CreateBatchChangePage'

const decorator: DecoratorFn = story => (
    <div className="p-3" style={{ height: '95vh', width: '100%' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/batches/create/CreateBatchChangePage',
    decorators: [decorator],
    parameters: {
        chromatic: {
            disableSnapshot: false,
        },
    },
}

export default config

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

export const ExperimentalExecutionDisabled: Story = () => (
    <WebStory>
        {props => (
            <CreateBatchChangePage
                {...props}
                headingElement="h1"
                authenticatedUser={mockAuthenticatedUser}
                settingsCascade={{
                    ...EMPTY_SETTINGS_CASCADE,
                    final: { experimentalFeatures: { batchChangesExecution: false } },
                }}
            />
        )}
    </WebStory>
)

ExperimentalExecutionDisabled.storyName = 'Experimental execution disabled'

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

export const ExperimentalExecutionEnabled: Story = () => (
    <WebStory>
        {props => (
            <CreateBatchChangePage
                {...props}
                headingElement="h1"
                authenticatedUser={mockAuthenticatedUser}
                settingsCascade={{
                    ...EMPTY_SETTINGS_CASCADE,
                    subjects: [
                        { subject: FIXTURE_ORG, settings: { a: 1 }, lastID: 1 },
                        { subject: FIXTURE_USER, settings: { b: 2 }, lastID: 2 },
                    ],
                }}
            />
        )}
    </WebStory>
)

ExperimentalExecutionEnabled.storyName = 'Experimental execution enabled'

export const ExperimentalExecutionEnabledFromOrgNamespace: Story = () => (
    <WebStory>
        {props => (
            <CreateBatchChangePage
                {...props}
                headingElement="h1"
                initialNamespaceID={MOCK_ORGANIZATION.id}
                authenticatedUser={mockAuthenticatedUser}
                settingsCascade={{
                    ...EMPTY_SETTINGS_CASCADE,
                    final: {
                        experimentalFeatures: { batchChangesExecution: true },
                    },
                    subjects: [
                        { subject: FIXTURE_ORG, settings: { a: 1 }, lastID: 1 },
                        { subject: FIXTURE_USER, settings: { b: 2 }, lastID: 2 },
                    ],
                }}
            />
        )}
    </WebStory>
)

ExperimentalExecutionEnabledFromOrgNamespace.storyName = 'Experimental execution enabled from org namespace'

export const ExperimentalExecutionEnabledFromUserNamespace: Story = () => (
    <WebStory>
        {props => (
            <CreateBatchChangePage
                {...props}
                headingElement="h1"
                initialNamespaceID={mockAuthenticatedUser.id}
                authenticatedUser={mockAuthenticatedUser}
                settingsCascade={{
                    ...EMPTY_SETTINGS_CASCADE,
                    final: {
                        experimentalFeatures: { batchChangesExecution: true },
                    },
                    subjects: [
                        { subject: FIXTURE_ORG, settings: { a: 1 }, lastID: 1 },
                        { subject: FIXTURE_USER, settings: { b: 2 }, lastID: 2 },
                    ],
                }}
            />
        )}
    </WebStory>
)

ExperimentalExecutionEnabledFromUserNamespace.storyName = 'Experimental execution enabled from user namespace'
