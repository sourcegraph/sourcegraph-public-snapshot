import { useMemo } from '@storybook/addons'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import { subDays } from 'date-fns'
import { of } from 'rxjs'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { BatchChangeByNamespaceResult, BatchChangeFields, ExternalServiceKind } from '../../../graphql-operations'

import {
    queryExternalChangesetWithFileDiffs,
    queryAllChangesetIDs as _queryAllChangesetIDs,
    BATCH_CHANGE_BY_NAMESPACE,
    BULK_OPERATIONS,
    CHANGESETS,
} from './backend'
import { BatchChangeDetailsPage } from './BatchChangeDetailsPage'
import {
    MOCK_BATCH_CHANGE,
    MOCK_BULK_OPERATIONS,
    BATCH_CHANGE_CHANGESETS_RESULT,
    EMPTY_BATCH_CHANGE_CHANGESETS_RESULT,
} from './BatchChangeDetailsPage.mock'
import { CHANGESET_COUNTS_OVER_TIME_MOCK } from './testdata'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>
const config: Meta = {
    title: 'web/batches/details/BatchChangeDetailsPage',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
            disableSnapshot: false,
        },
        controls: {
            exclude: ['url', 'supersededBatchSpec'],
        },
    },
    argTypes: {
        supersedingBatchSpec: {
            control: { type: 'boolean' },
        },
        viewerCanAdminister: {
            control: { type: 'boolean' },
            defaultValue: true,
        },
        isClosed: {
            control: { type: 'boolean' },
            defaultValue: false,
        },
    },
}

export default config

const now = new Date()

const authenticatedUser = { url: 'https://sourcegraph.com/users/this-is-a-fake-user' }

const queryAllChangesetIDs: typeof _queryAllChangesetIDs = () => of(['somev1', 'somev2'])

const queryEmptyExternalChangesetWithFileDiffs: typeof queryExternalChangesetWithFileDiffs = () =>
    of({
        diff: {
            __typename: 'PreviewRepositoryComparison',
            fileDiffs: {
                nodes: [],
                totalCount: 0,
                pageInfo: {
                    endCursor: null,
                    hasNextPage: false,
                },
            },
        },
    })

const deleteBatchChange = () => Promise.resolve(undefined)

const Template: Story<{
    url: string
    supersedingBatchSpec?: boolean
    currentBatchSpec?: BatchChangeFields['currentSpec']
    viewerCanAdminister: boolean
    isClosed?: boolean
}> = ({ url, supersedingBatchSpec, currentBatchSpec, viewerCanAdminister, isClosed }) => {
    const batchChange: BatchChangeFields = useMemo(() => {
        const currentSpec = currentBatchSpec ?? MOCK_BATCH_CHANGE.currentSpec

        return {
            ...MOCK_BATCH_CHANGE,
            currentSpec: {
                ...currentSpec,
                supersedingBatchSpec: supersedingBatchSpec
                    ? {
                          __typename: 'BatchSpec',
                          createdAt: subDays(new Date(), 1).toISOString(),
                          applyURL: '/users/alice/batch-changes/apply/newspecid',
                      }
                    : null,
            },
            viewerCanAdminister,
            closedAt: isClosed ? subDays(now, 1).toISOString() : null,
        }
    }, [currentBatchSpec, supersedingBatchSpec, viewerCanAdminister, isClosed])

    const data: BatchChangeByNamespaceResult = { batchChange }

    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(BATCH_CHANGE_BY_NAMESPACE),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(BULK_OPERATIONS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: MOCK_BULK_OPERATIONS },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(CHANGESETS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: { node: BATCH_CHANGE_CHANGESETS_RESULT } },
            nMatches: Number.POSITIVE_INFINITY,
        },
        CHANGESET_COUNTS_OVER_TIME_MOCK,
    ])

    return (
        <WebStory initialEntries={[url]}>
            {props => (
                <MockedTestProvider link={mocks}>
                    <BatchChangeDetailsPage
                        {...props}
                        authenticatedUser={authenticatedUser}
                        namespaceID="namespace123"
                        batchChangeName="awesome-batch-change"
                        queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                        deleteBatchChange={deleteBatchChange}
                        queryAllChangesetIDs={queryAllChangesetIDs}
                        settingsCascade={EMPTY_SETTINGS_CASCADE}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

export const Overview = Template.bind({})
Overview.args = { url: '/users/alice/batch-changes/awesome-batch-change' }
Overview.argTypes = {
    supersedingBatchSpec: {
        defaultValue: false,
    },
}

export const BurndownChart = Template.bind({})
BurndownChart.args = { url: '/users/alice/batch-changes/awesome-batch-change?tab=chart' }
BurndownChart.storyName = 'Burndown chart'
BurndownChart.argTypes = {
    supersedingBatchSpec: {
        defaultValue: false,
    },
}

export const SpecFile = Template.bind({})
SpecFile.args = { url: '/users/alice/batch-changes/awesome-batch-change?tab=spec' }
SpecFile.storyName = 'Spec file'
SpecFile.argTypes = {
    supersedingBatchSpec: {
        defaultValue: false,
    },
    viewerCanAdminister: {
        defaultValue: false,
    },
}

export const Archived = Template.bind({})
Archived.args = { url: '/users/alice/batch-changes/awesome-batch-change?tab=archived' }
Archived.argTypes = {
    supersedingBatchSpec: {
        defaultValue: false,
    },
}

export const BulkOperations = Template.bind({})
BulkOperations.args = { url: '/users/alice/batch-changes/awesome-batch-change?tab=bulkoperations' }
BulkOperations.storyName = 'Bulk operations'
BulkOperations.argTypes = {
    supersedingBatchSpec: {
        defaultValue: false,
    },
}

export const SupersededBatchSpec = Template.bind({})
SupersededBatchSpec.args = { url: '/users/alice/batch-changes/awesome-batch-change', supersedingBatchSpec: true }
SupersededBatchSpec.storyName = 'Superseded batch-spec'
SupersededBatchSpec.argTypes = {
    supersedingBatchSpec: {
        defaultValue: true,
    },
}

export const UnpublishableBatchSpec = Template.bind({})
UnpublishableBatchSpec.args = {
    url: '/users/alice/batch-changes/awesome-batch-change',
    currentBatchSpec: {
        ...MOCK_BATCH_CHANGE.currentSpec,
        viewerBatchChangesCodeHosts: {
            __typename: 'BatchChangesCodeHostConnection',
            totalCount: 1,
            nodes: [
                {
                    externalServiceURL: 'https://github.com/',
                    externalServiceKind: ExternalServiceKind.GITHUB,
                },
            ],
        },
    },
}
UnpublishableBatchSpec.storyName = 'Batch spec with unpublishable changesets'
UnpublishableBatchSpec.argTypes = {
    supersedingBatchSpec: {
        defaultValue: true,
    },
}

export const EmptyChangesets: Story = args => {
    const mocks = new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(BATCH_CHANGE_BY_NAMESPACE),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: { batchChange: MOCK_BATCH_CHANGE } },
            nMatches: Number.POSITIVE_INFINITY,
        },
        {
            request: {
                query: getDocumentNode(CHANGESETS),
                variables: MATCH_ANY_PARAMETERS,
            },
            result: { data: { node: EMPTY_BATCH_CHANGE_CHANGESETS_RESULT } },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={mocks}>
                    <BatchChangeDetailsPage
                        {...props}
                        authenticatedUser={authenticatedUser}
                        namespaceID="namespace123"
                        batchChangeName="awesome-batch-change"
                        queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                        deleteBatchChange={deleteBatchChange}
                        settingsCascade={EMPTY_SETTINGS_CASCADE}
                        {...args}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}
EmptyChangesets.parameters = {
    controls: { hideNoControlsWarning: true, exclude: ['supersedingBatchSpec', 'viewerCanAdminister', 'isClosed'] },
}

EmptyChangesets.storyName = 'Empty changesets'
