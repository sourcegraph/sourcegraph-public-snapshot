import { DecoratorFn, Meta, Story } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { GET_LICENSE_AND_USAGE_INFO } from '../list/backend'
import { getLicenseAndUsageInfoResult } from '../list/testData'

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

const buildMocks = (isLicensed = true, hasBatchChanges = true) =>
    new WildcardMockLink([
        {
            request: { query: getDocumentNode(GET_LICENSE_AND_USAGE_INFO), variables: MATCH_ANY_PARAMETERS },
            result: { data: getLicenseAndUsageInfoResult(isLicensed, hasBatchChanges) },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

export const NewBatchChange: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildMocks()}>
                <ConfigurationForm {...props} settingsCascade={SETTINGS_CASCADE} />
            </MockedTestProvider>
        )}
    </WebStory>
)

NewBatchChange.storyName = 'New batch change'

export const ExistingBatchChange: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildMocks()}>
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
            </MockedTestProvider>
        )}
    </WebStory>
)

ExistingBatchChange.storyName = 'Read-only for existing batch change'

export const LicenseAlert: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildMocks(false)}>
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
            </MockedTestProvider>
        )}
    </WebStory>
)

LicenseAlert.storyName = 'License alert'
