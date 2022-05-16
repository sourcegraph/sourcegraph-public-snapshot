import { boolean } from '@storybook/addon-knobs'
import { useMemo } from '@storybook/addons'
import { storiesOf } from '@storybook/react'
import { addDays, subDays } from 'date-fns'
import { of, Observable } from 'rxjs'
import { WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import {
    ApplyPreviewStatsFields,
    BatchSpecApplyPreviewConnectionFields,
    BatchSpecFields,
    ChangesetApplyPreviewFields,
    ExternalServiceKind,
} from '../../../graphql-operations'

import { BATCH_SPEC_BY_ID } from './backend'
import { BatchChangePreviewPage } from './BatchChangePreviewPage'
import { hiddenChangesetApplyPreviewStories } from './list/HiddenChangesetApplyPreviewNode.story'
import { visibleChangesetApplyPreviewNodeStories } from './list/VisibleChangesetApplyPreviewNode.story'

const { add } = storiesOf('web/batches/preview/BatchChangePreviewPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
            disableSnapshot: false,
        },
    })

const nodes: ChangesetApplyPreviewFields[] = [
    ...Object.values(visibleChangesetApplyPreviewNodeStories(false)),
    ...Object.values(hiddenChangesetApplyPreviewStories),
]

const batchSpec = (): BatchSpecFields => ({
    appliesToBatchChange: null,
    createdAt: subDays(new Date(), 5).toISOString(),
    creator: {
        __typename: 'User',
        url: '/users/alice',
        username: 'alice',
    },
    description: {
        __typename: 'BatchChangeDescription',
        name: 'awesome-batch-change',
        description: 'This is the description',
    },
    diffStat: {
        __typename: 'DiffStat',
        added: 10,
        changed: 8,
        deleted: 10,
    },
    expiresAt: addDays(new Date(), 7).toISOString(),
    id: 'specid',
    namespace: {
        __typename: 'User',
        namespaceName: 'alice',
        url: '/users/alice',
    },
    supersedingBatchSpec: boolean('supersedingBatchSpec', false)
        ? {
              __typename: 'BatchSpec',
              createdAt: subDays(new Date(), 1).toISOString(),
              applyURL: '/users/alice/batch-changes/apply/newspecid',
          }
        : null,
    viewerCanAdminister: boolean('viewerCanAdminister', true),
    viewerBatchChangesCodeHosts: {
        __typename: 'BatchChangesCodeHostConnection',
        totalCount: 0,
        nodes: [],
    },
    originalInput: 'name: awesome-batch-change\ndescription: somestring',
    applyPreview: {
        __typename: 'ChangesetApplyPreviewConnection',
        stats: {
            archive: 18,
        },
        totalCount: 18,
    },
})

// This has to be a link so we can return as many mock responses are required
// for the time the storybook is open.
const batchSpecByIDLink = (spec: BatchSpecFields): WildcardMockLink =>
    new WildcardMockLink([
        {
            request: {
                query: getDocumentNode(BATCH_SPEC_BY_ID),
                variables: {
                    batchSpec: '123123',
                },
            },
            result: {
                data: {
                    node: {
                        __typename: 'BatchSpec',
                        ...spec,
                    },
                },
            },
            nMatches: Number.POSITIVE_INFINITY,
        },
    ])

const fetchBatchSpecCreate = () => batchSpecByIDLink(batchSpec())

const fetchBatchSpecMissingCredentials = () =>
    batchSpecByIDLink({
        ...batchSpec(),
        viewerBatchChangesCodeHosts: {
            __typename: 'BatchChangesCodeHostConnection',
            totalCount: 2,
            nodes: [
                {
                    externalServiceKind: ExternalServiceKind.GITHUB,
                    externalServiceURL: 'https://github.com/',
                },
                {
                    externalServiceKind: ExternalServiceKind.GITLAB,
                    externalServiceURL: 'https://gitlab.com/',
                },
            ],
        },
    })

const fetchBatchSpecUpdate = () =>
    batchSpecByIDLink({
        ...batchSpec(),
        appliesToBatchChange: {
            id: 'somebatch',
            name: 'awesome-batch-change',
            url: '/users/alice/batch-changes/awesome-batch-change',
        },
    })

const queryApplyPreviewStats = (): Observable<ApplyPreviewStatsFields['stats']> =>
    of({
        close: 10,
        detach: 10,
        import: 10,
        publish: 10,
        publishDraft: 10,
        push: 10,
        reopen: 10,
        undraft: 10,
        update: 10,
        archive: 18,
        added: 5,
        modified: 10,
        removed: 3,
    })

const queryChangesetApplyPreview = (): Observable<BatchSpecApplyPreviewConnectionFields> =>
    of({
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: nodes.length,
        nodes,
    })

const queryEmptyChangesetApplyPreview = (): Observable<BatchSpecApplyPreviewConnectionFields> =>
    of({
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: 0,
        nodes: [],
    })

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

add('Create', () => {
    const link = useMemo(() => fetchBatchSpecCreate(), [])
    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={link}>
                    <BatchChangePreviewPage
                        {...props}
                        expandChangesetDescriptions={true}
                        batchSpecID="123123"
                        queryChangesetApplyPreview={queryChangesetApplyPreview}
                        queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                        queryApplyPreviewStats={queryApplyPreviewStats}
                        authenticatedUser={{
                            url: '/users/alice',
                            displayName: 'Alice',
                            username: 'alice',
                            email: 'alice@email.test',
                        }}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

add('Update', () => {
    const link = useMemo(() => fetchBatchSpecUpdate(), [])
    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={link}>
                    <BatchChangePreviewPage
                        {...props}
                        expandChangesetDescriptions={true}
                        batchSpecID="123123"
                        queryChangesetApplyPreview={queryChangesetApplyPreview}
                        queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                        queryApplyPreviewStats={queryApplyPreviewStats}
                        authenticatedUser={{
                            url: '/users/alice',
                            displayName: 'Alice',
                            username: 'alice',
                            email: 'alice@email.test',
                        }}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

add('Missing credentials', () => {
    const link = useMemo(() => fetchBatchSpecMissingCredentials(), [])
    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={link}>
                    <BatchChangePreviewPage
                        {...props}
                        expandChangesetDescriptions={true}
                        batchSpecID="123123"
                        queryChangesetApplyPreview={queryChangesetApplyPreview}
                        queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                        queryApplyPreviewStats={queryApplyPreviewStats}
                        authenticatedUser={{
                            url: '/users/alice',
                            displayName: 'Alice',
                            username: 'alice',
                            email: 'alice@email.test',
                        }}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

add('Spec file', () => {
    const link = useMemo(() => fetchBatchSpecCreate(), [])
    return (
        <WebStory initialEntries={['/users/alice/batch-changes/awesome-batch-change?tab=spec']}>
            {props => (
                <MockedTestProvider link={link}>
                    <BatchChangePreviewPage
                        {...props}
                        expandChangesetDescriptions={true}
                        batchSpecID="123123"
                        queryChangesetApplyPreview={queryChangesetApplyPreview}
                        queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                        queryApplyPreviewStats={queryApplyPreviewStats}
                        authenticatedUser={{
                            url: '/users/alice',
                            displayName: 'Alice',
                            username: 'alice',
                            email: 'alice@email.test',
                        }}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})

add('No changesets', () => {
    const link = useMemo(() => fetchBatchSpecCreate(), [])
    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={link}>
                    <BatchChangePreviewPage
                        {...props}
                        expandChangesetDescriptions={true}
                        batchSpecID="123123"
                        queryChangesetApplyPreview={queryEmptyChangesetApplyPreview}
                        queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                        queryApplyPreviewStats={queryApplyPreviewStats}
                        authenticatedUser={{
                            url: '/users/alice',
                            displayName: 'Alice',
                            username: 'alice',
                            email: 'alice@email.test',
                        }}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
})
