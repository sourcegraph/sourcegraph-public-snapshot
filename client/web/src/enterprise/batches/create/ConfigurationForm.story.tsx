import { storiesOf } from '@storybook/react'

import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'

import { WebStory } from '../../../components/WebStory'

import { ConfigurationForm } from './ConfigurationForm'

const { add } = storiesOf('web/batches/create/ConfigurationForm', module)
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

add('new batch change', () => (
    <WebStory>{props => <ConfigurationForm {...props} settingsCascade={SETTINGS_CASCADE} />}</WebStory>
))

add('read-only for existing batch change', () => (
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
))
