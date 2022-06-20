import { storiesOf } from '@storybook/react'

import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'

import { WebStory } from '../../../components/WebStory'

import { CreateBatchChangePage } from './CreateBatchChangePage'

const { add } = storiesOf('web/batches/create/CreateBatchChangePage', module)
    .addDecorator(story => (
        <div className="p-3" style={{ height: '95vh', width: '100%' }}>
            {story()}
        </div>
    ))
    .addParameters({
        chromatic: {
            disableSnapshot: false,
        },
    })

add('experimental execution disabled', () => (
    <WebStory>
        {props => (
            <CreateBatchChangePage
                {...props}
                headingElement="h1"
                settingsCascade={{
                    ...EMPTY_SETTINGS_CASCADE,
                    final: { experimentalFeatures: { batchChangesExecution: false } },
                }}
            />
        )}
    </WebStory>
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

add('experimental execution enabled', () => (
    <WebStory>
        {props => (
            <CreateBatchChangePage
                {...props}
                headingElement="h1"
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
))

add('experimental execution enabled, from org namespace', () => (
    <WebStory>
        {props => (
            <CreateBatchChangePage
                {...props}
                headingElement="h1"
                initialNamespaceID="a"
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
))

add('experimental execution enabled, from user namespace', () => (
    <WebStory>
        {props => (
            <CreateBatchChangePage
                {...props}
                headingElement="h1"
                initialNamespaceID="b"
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
))
