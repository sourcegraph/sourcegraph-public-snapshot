import { Meta, Story, DecoratorFn } from '@storybook/react'

import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'

import { WebStory } from '../../../components/WebStory'

import { ConfigurationForm } from './ConfigurationForm'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/create/ConfigurationForm',
    decorators: [decorator],
    parameters: {
        chromatic: {
            disableSnapshot: false,
        },
    },
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

export const NewBatchChange: Story = () => (
    <WebStory>{props => <ConfigurationForm {...props} settingsCascade={SETTINGS_CASCADE} />}</WebStory>
)

NewBatchChange.storyName = 'New batch change'

export const ExistingBatchChange: Story = () => (
    <WebStory>
        {props => (
            <ConfigurationForm
                {...props}
                settingsCascade={SETTINGS_CASCADE}
                isReadOnly={true}
                batchChange={{
                    name: 'My existing batch change',
                    namespace: {
                        __typename: 'Org',
                        namespaceName: 'Sourcegraph',
                        url: '/orgs/sourcegraph',
                        id: 'test1234',
                    },
                }}
            />
        )}
    </WebStory>
)

ExistingBatchChange.storyName = 'Read-only for existing batch change'
