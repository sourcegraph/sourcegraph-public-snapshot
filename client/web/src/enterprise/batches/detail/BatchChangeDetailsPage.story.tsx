import { boolean } from '@storybook/addon-knobs'
import { useMemo } from '@storybook/addons'
import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import { of } from 'rxjs'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { BatchChangeByNamespaceResult, BatchChangeFields } from '../../../graphql-operations'

import {
    queryExternalChangesetWithFileDiffs,
    queryChangesetCountsOverTime as _queryChangesetCountsOverTime,
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

const { add } = storiesOf('web/batches/details/BatchChangeDetailsPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
            disableSnapshot: false,
        },
    })

const now = new Date()

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

const queryChangesetCountsOverTime: typeof _queryChangesetCountsOverTime = () =>
    of([
        {
            date: subDays(new Date('2020-08-10'), 5).toISOString(),
            closed: 0,
            merged: 0,
            openPending: 5,
            total: 10,
            draft: 5,
            openChangesRequested: 0,
            openApproved: 0,
        },
        {
            date: subDays(new Date('2020-08-10'), 4).toISOString(),
            closed: 0,
            merged: 0,
            openPending: 4,
            total: 10,
            draft: 3,
            openChangesRequested: 0,
            openApproved: 3,
        },
        {
            date: subDays(new Date('2020-08-10'), 3).toISOString(),
            closed: 0,
            merged: 2,
            openPending: 5,
            total: 10,
            draft: 0,
            openChangesRequested: 0,
            openApproved: 3,
        },
        {
            date: subDays(new Date('2020-08-10'), 2).toISOString(),
            closed: 0,
            merged: 3,
            openPending: 3,
            total: 10,
            draft: 0,
            openChangesRequested: 1,
            openApproved: 3,
        },
        {
            date: subDays(new Date('2020-08-10'), 1).toISOString(),
            closed: 1,
            merged: 5,
            openPending: 2,
            total: 10,
            draft: 0,
            openChangesRequested: 0,
            openApproved: 2,
        },
        {
            date: new Date('2020-08-10').toISOString(),
            closed: 1,
            merged: 5,
            openPending: 0,
            total: 10,
            draft: 0,
            openChangesRequested: 0,
            openApproved: 4,
        },
    ])

const deleteBatchChange = () => Promise.resolve(undefined)

const stories: Record<string, { url: string; supersededBatchSpec?: boolean }> = {
    Overview: { url: '/users/alice/batch-changes/awesome-batch-change' },
    'Burndown chart': { url: '/users/alice/batch-changes/awesome-batch-change?tab=chart' },
    'Spec file': { url: '/users/alice/batch-changes/awesome-batch-change?tab=spec' },
    Archived: { url: '/users/alice/batch-changes/awesome-batch-change?tab=archived' },
    'Bulk operations': { url: '/users/alice/batch-changes/awesome-batch-change?tab=bulkoperations' },
    'Superseded batch-spec': { url: '/users/alice/batch-changes/awesome-batch-change', supersededBatchSpec: true },
}

for (const [name, { url, supersededBatchSpec }] of Object.entries(stories)) {
    add(name, () => {
        const supersedingBatchSpec = boolean('supersedingBatchSpec', !!supersededBatchSpec)
        const viewerCanAdminister = boolean('viewerCanAdminister', true)
        const isClosed = boolean('isClosed', false)
        const batchChange: BatchChangeFields = useMemo(
            () => ({
                ...MOCK_BATCH_CHANGE,
                currentSpec: {
                    ...MOCK_BATCH_CHANGE.currentSpec,
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
            }),
            [supersedingBatchSpec, viewerCanAdminister, isClosed]
        )

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
        ])

        return (
            <WebStory initialEntries={[url]}>
                {props => (
                    <MockedTestProvider link={mocks}>
                        <BatchChangeDetailsPage
                            {...props}
                            namespaceID="namespace123"
                            batchChangeName="awesome-batch-change"
                            queryChangesetCountsOverTime={queryChangesetCountsOverTime}
                            queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                            deleteBatchChange={deleteBatchChange}
                            queryAllChangesetIDs={queryAllChangesetIDs}
                            extensionsController={{} as any}
                            platformContext={{} as any}
                            settingsCascade={EMPTY_SETTINGS_CASCADE}
                        />
                    </MockedTestProvider>
                )}
            </WebStory>
        )
    })
}

add('Empty changesets', () => {
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
                        namespaceID="namespace123"
                        batchChangeName="awesome-batch-change"
                        queryChangesetCountsOverTime={queryChangesetCountsOverTime}
                        queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                        deleteBatchChange={deleteBatchChange}
                        extensionsController={{} as any}
                        platformContext={{} as any}
                        settingsCascade={EMPTY_SETTINGS_CASCADE}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})
