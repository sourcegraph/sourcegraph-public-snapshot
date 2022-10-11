import { DecoratorFn, Story, Meta } from '@storybook/react'
import { WildcardMockLink, MATCH_ANY_PARAMETERS } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { GlobalChangesetsStatsResult } from '../../../graphql-operations'

import {
    GLOBAL_CHANGESETS_STATS,
    BATCH_CHANGES,
    BATCH_CHANGES_BY_NAMESPACE,
    GET_LICENSE_AND_USAGE_INFO,
} from './backend'
import { BatchChangeListPage } from './BatchChangeListPage'
import {
    BATCH_CHANGES_BY_NAMESPACE_RESULT,
    BATCH_CHANGES_RESULT,
    getLicenseAndUsageInfoResult,
    NO_BATCH_CHANGES_RESULT,
} from './testData'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/list/BatchChangeListPage',
    decorators: [decorator],

    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
            disableSnapshot: false,
        },
    },
}

export default config

const statBarData: GlobalChangesetsStatsResult = {
    __typename: 'Query',
    batchChanges: { __typename: 'BatchChangeConnection', totalCount: 30 },
    globalChangesetsStats: { __typename: 'GlobalChangesetsStats', open: 7, closed: 5, merged: 21 },
}

const buildMocks = (isLicensed = true, hasBatchChanges = true, hasFilteredBatchChanges = true) =>
    new WildcardMockLink([
        {
            request: { query: getDocumentNode(BATCH_CHANGES), variables: MATCH_ANY_PARAMETERS },
            result: {
                data: hasBatchChanges && hasFilteredBatchChanges ? BATCH_CHANGES_RESULT : NO_BATCH_CHANGES_RESULT,
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: { query: getDocumentNode(GET_LICENSE_AND_USAGE_INFO), variables: MATCH_ANY_PARAMETERS },
            result: { data: getLicenseAndUsageInfoResult(isLicensed, hasBatchChanges) },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(GLOBAL_CHANGESETS_STATS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: {
                data: statBarData,
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

const MOCKS_FOR_NAMESPACE = new WildcardMockLink([
    {
        request: { query: getDocumentNode(BATCH_CHANGES_BY_NAMESPACE), variables: MATCH_ANY_PARAMETERS },
        result: { data: BATCH_CHANGES_BY_NAMESPACE_RESULT },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: { query: getDocumentNode(GET_LICENSE_AND_USAGE_INFO), variables: MATCH_ANY_PARAMETERS },
        result: { data: getLicenseAndUsageInfoResult() },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

export const ListOfBatchChanges: Story = args => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildMocks()}>
                <BatchChangeListPage
                    {...props}
                    headingElement="h1"
                    canCreate={args.canCreate}
                    settingsCascade={EMPTY_SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)
ListOfBatchChanges.argTypes = {
    canCreate: {
        name: 'can create batch changes',
        control: { type: 'boolean' },
        defaultValue: true,
    },
}

ListOfBatchChanges.storyName = 'List of batch changes'

export const ListOfBatchChangesSpecificNamespace: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={MOCKS_FOR_NAMESPACE}>
                <BatchChangeListPage
                    {...props}
                    headingElement="h1"
                    canCreate={true}
                    namespaceID="test-12345"
                    settingsCascade={EMPTY_SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ListOfBatchChangesSpecificNamespace.storyName = 'List of batch changes, for a specific namespace'

export const ListOfBatchChangesServerSideExecutionEnabled: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildMocks()}>
                <BatchChangeListPage
                    {...props}
                    headingElement="h1"
                    canCreate={true}
                    settingsCascade={{
                        ...EMPTY_SETTINGS_CASCADE,
                        final: {
                            experimentalFeatures: { batchChangesExecution: true },
                        },
                    }}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ListOfBatchChangesServerSideExecutionEnabled.storyName = 'List of batch changes, server-side execution enabled'

export const LicensingNotEnforced: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildMocks(false)}>
                <BatchChangeListPage
                    {...props}
                    headingElement="h1"
                    canCreate={true}
                    settingsCascade={EMPTY_SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

LicensingNotEnforced.storyName = 'Licensing not enforced'

export const NoBatchChanges: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildMocks(true, false)}>
                <BatchChangeListPage
                    {...props}
                    headingElement="h1"
                    canCreate={true}
                    settingsCascade={EMPTY_SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

NoBatchChanges.storyName = 'No batch changes'

export const AllBatchChangesTabEmpty: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={buildMocks(true, true, false)}>
                <BatchChangeListPage
                    {...props}
                    headingElement="h1"
                    canCreate={true}
                    openTab="batchChanges"
                    settingsCascade={EMPTY_SETTINGS_CASCADE}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

AllBatchChangesTabEmpty.storyName = 'All batch changes tab empty'
