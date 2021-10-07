import { storiesOf } from '@storybook/react'
import React from 'react'

import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { CreateBatchChangePage } from './CreateBatchChangePage'

const { add } = storiesOf('web/batches/CreateBatchChangePage', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
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

add('experimental execution disabled', () => (
    <EnterpriseWebStory>
        {props => <CreateBatchChangePage headingElement="h1" {...props} settingsCascade={EMPTY_SETTINGS_CASCADE} />}
    </EnterpriseWebStory>
))

add('experimental execution enabled', () => (
    <EnterpriseWebStory>
        {props => (
            <CreateBatchChangePage
                headingElement="h1"
                {...props}
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
    </EnterpriseWebStory>
))
