import { useMemo } from '@storybook/addons'
import type { Decorator, Meta, StoryFn, StoryObj } from '@storybook/react'
import { subDays } from 'date-fns'
import { of } from 'rxjs'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import {
    type BatchChangeByNamespaceResult,
    type BatchChangeFields,
    ExternalServiceKind,
} from '../../../graphql-operations'

import {
    type queryExternalChangesetWithFileDiffs,
    type queryAllChangesetIDs as _queryAllChangesetIDs,
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

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>
const config: Meta<Args> = {
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
        },
        isClosed: {
            control: { type: 'boolean' },
        },
    },
    args: {
        viewerCanAdminister: true,
        isClosed: false,
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

interface Args {
    url: string
    supersedingBatchSpec?: boolean
    currentBatchSpec?: BatchChangeFields['currentSpec']
    viewerCanAdminister: boolean
    isClosed?: boolean
}

const Template: StoryFn<Args> = ({ url, supersedingBatchSpec, currentBatchSpec, viewerCanAdminister, isClosed }) => {
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
        <WebStory path="/users/:username/batch-changes/:batchChangeName" initialEntries={[url]}>
            {props => (
                <MockedTestProvider link={mocks}>
                    <BatchChangeDetailsPage
                        {...props}
                        authenticatedUser={authenticatedUser}
                        namespaceID="namespace123"
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

type Story = StoryObj<typeof config>

export const Overview: Story = Template.bind({})
Overview.args = { url: '/users/alice/batch-changes/awesome-batch-change', supersedingBatchSpec: false }
Overview.argTypes = {
    supersedingBatchSpec: {},
}

export const BurndownChart: Story = Template.bind({})
BurndownChart.args = { url: '/users/alice/batch-changes/awesome-batch-change?tab=chart', supersedingBatchSpec: false }
BurndownChart.storyName = 'Burndown chart'
BurndownChart.argTypes = {
    supersedingBatchSpec: {},
}

export const SpecFile: Story = Template.bind({})
SpecFile.args = {
    url: '/users/alice/batch-changes/awesome-batch-change?tab=spec',
    supersedingBatchSpec: false,
    viewerCanAdminister: false,
}
SpecFile.storyName = 'Spec file'
SpecFile.argTypes = {
    supersedingBatchSpec: {},
    viewerCanAdminister: {},
}

export const Archived: Story = Template.bind({})
Archived.args = { url: '/users/alice/batch-changes/awesome-batch-change?tab=archived', supersedingBatchSpec: false }
Archived.argTypes = {
    supersedingBatchSpec: {},
}

export const BulkOperations: Story = Template.bind({})
BulkOperations.args = {
    url: '/users/alice/batch-changes/awesome-batch-change?tab=bulkoperations',
    supersedingBatchSpec: false,
}
BulkOperations.storyName = 'Bulk operations'
BulkOperations.argTypes = {
    supersedingBatchSpec: {},
}

export const SupersededBatchSpec: Story = Template.bind({})
SupersededBatchSpec.args = { url: '/users/alice/batch-changes/awesome-batch-change', supersedingBatchSpec: true }
SupersededBatchSpec.storyName = 'Superseded batch-spec'
SupersededBatchSpec.argTypes = {
    supersedingBatchSpec: {},
}

export const UnpublishableBatchSpec: Story = Template.bind({})
UnpublishableBatchSpec.args = {
    url: '/users/alice/batch-changes/awesome-batch-change',
    supersedingBatchSpec: true,
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
    supersedingBatchSpec: {},
}

export const EmptyChangesets: StoryFn = args => {
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
        <WebStory path="/:batchChangeName" initialEntries={['/awesome-batch-change']}>
            {props => (
                <MockedTestProvider link={mocks}>
                    <BatchChangeDetailsPage
                        {...props}
                        authenticatedUser={authenticatedUser}
                        namespaceID="namespace123"
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
