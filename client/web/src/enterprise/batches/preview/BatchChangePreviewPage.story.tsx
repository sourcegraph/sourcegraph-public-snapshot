import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { addDays, subDays } from 'date-fns'
import React from 'react'
import { of, Observable } from 'rxjs'

import { WebStory } from '../../../components/WebStory'
import {
    ApplyPreviewStatsFields,
    BatchSpecApplyPreviewConnectionFields,
    BatchSpecFields,
    ChangesetApplyPreviewFields,
    ExternalServiceKind,
} from '../../../graphql-operations'

import { fetchBatchSpecById } from './backend'
import { BatchChangePreviewPage } from './BatchChangePreviewPage'
import { hiddenChangesetApplyPreviewStories } from './list/HiddenChangesetApplyPreviewNode.story'
import { visibleChangesetApplyPreviewNodeStories } from './list/VisibleChangesetApplyPreviewNode.story'

const { add } = storiesOf('web/batches/preview/BatchChangePreviewPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
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
        url: '/users/alice',
        username: 'alice',
    },
    description: {
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
        namespaceName: 'alice',
        url: '/users/alice',
    },
    supersedingBatchSpec: boolean('supersedingBatchSpec', false)
        ? {
              createdAt: subDays(new Date(), 1).toISOString(),
              applyURL: '/users/alice/batch-changes/apply/newspecid',
          }
        : null,
    viewerCanAdminister: boolean('viewerCanAdminister', true),
    viewerBatchChangesCodeHosts: {
        totalCount: 0,
        nodes: [],
    },
    originalInput: 'name: awesome-batch-change\ndescription: somestring',
    applyPreview: {
        stats: {
            archive: 18,
        },
        totalCount: 18,
    },
})

const fetchBatchSpecCreate: typeof fetchBatchSpecById = () => of(batchSpec())

const fetchBatchSpecMissingCredentials: typeof fetchBatchSpecById = () =>
    of({
        ...batchSpec(),
        viewerBatchChangesCodeHosts: {
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

const fetchBatchSpecUpdate: typeof fetchBatchSpecById = () =>
    of({
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

add('Create', () => (
    <WebStory>
        {props => (
            <BatchChangePreviewPage
                {...props}
                expandChangesetDescriptions={true}
                batchSpecID="123123"
                fetchBatchSpecById={fetchBatchSpecCreate}
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
        )}
    </WebStory>
))

add('Update', () => (
    <WebStory>
        {props => (
            <BatchChangePreviewPage
                {...props}
                expandChangesetDescriptions={true}
                batchSpecID="123123"
                fetchBatchSpecById={fetchBatchSpecUpdate}
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
        )}
    </WebStory>
))

add('Missing credentials', () => (
    <WebStory>
        {props => (
            <BatchChangePreviewPage
                {...props}
                expandChangesetDescriptions={true}
                batchSpecID="123123"
                fetchBatchSpecById={fetchBatchSpecMissingCredentials}
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
        )}
    </WebStory>
))

add('Spec file', () => (
    <WebStory initialEntries={['/users/alice/batch-changes/awesome-batch-change?tab=spec']}>
        {props => (
            <BatchChangePreviewPage
                {...props}
                expandChangesetDescriptions={true}
                batchSpecID="123123"
                fetchBatchSpecById={fetchBatchSpecCreate}
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
        )}
    </WebStory>
))

add('No changesets', () => (
    <WebStory>
        {props => (
            <BatchChangePreviewPage
                {...props}
                expandChangesetDescriptions={true}
                batchSpecID="123123"
                fetchBatchSpecById={fetchBatchSpecCreate}
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
        )}
    </WebStory>
))
