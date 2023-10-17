import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { of } from 'rxjs'

import { mockAuthenticatedUser } from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'

import { WebStory } from '../../../components/WebStory'
import { type RepoBatchChange, type RepositoryFields, RepositoryType } from '../../../graphql-operations'
import type { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'

import type {
    queryRepoBatchChanges as _queryRepoBatchChanges,
    queryRepoBatchChangeStats as _queryRepoBatchChangeStats,
} from './backend'
import { BatchChangeRepoPage } from './BatchChangeRepoPage'
import { NODES } from './testData'

const repoDefaults: RepositoryFields = {
    description: 'An awesome repo!',
    defaultBranch: null,
    viewerCanAdminister: false,
    externalURLs: [],
    externalRepository: { serviceType: 'github', serviceID: 'https://github.com/' },
    id: 'repoid',
    name: 'github.com/sourcegraph/awesome',
    url: 'http://test.test/awesome',
    isFork: false,
    metadata: [],
    sourceType: RepositoryType.GIT_REPOSITORY,
}

const queryRepoBatchChangeStats: typeof _queryRepoBatchChangeStats = () =>
    of({
        batchChangesDiffStat: {
            added: 247,
            deleted: 990,
        },
        changesetsStats: {
            unpublished: 1,
            draft: 1,
            open: 2,
            merged: 47,
            closed: 9,
            total: 60,
        },
    })

const queryEmptyRepoBatchChangeStats: typeof _queryRepoBatchChangeStats = () =>
    of({
        batchChangesDiffStat: {
            added: 0,
            deleted: 0,
        },
        changesetsStats: {
            unpublished: 0,
            draft: 0,
            open: 0,
            merged: 0,
            closed: 0,
            total: 0,
        },
    })

const queryRepoBatchChanges =
    (nodes: RepoBatchChange[]): typeof _queryRepoBatchChanges =>
    () =>
        of({
            batchChanges: {
                totalCount: Object.values(nodes).length,
                nodes: Object.values(nodes),
                pageInfo: { endCursor: null, hasNextPage: false },
            },
        })

const queryList = queryRepoBatchChanges(NODES)
const queryNone = queryRepoBatchChanges([])

const queryEmptyExternalChangesetWithFileDiffs: typeof _queryExternalChangesetWithFileDiffs = () =>
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

const decorator: Decorator = story => <div className="p-3 container web-content">{story()}</div>

const config: Meta = {
    title: 'web/batches/repo/BatchChangeRepoPage',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
            disableSnapshot: false,
        },
    },
}

export default config

export const ListOfBatchChanges: StoryFn = () => (
    <WebStory initialEntries={['/github.com/sourcegraph/awesome/-/batch-changes']}>
        {props => (
            <BatchChangeRepoPage
                {...props}
                repo={repoDefaults}
                authenticatedUser={mockAuthenticatedUser}
                isSourcegraphDotCom={false}
                queryRepoBatchChangeStats={queryRepoBatchChangeStats}
                queryRepoBatchChanges={queryList}
                queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
            />
        )}
    </WebStory>
)

ListOfBatchChanges.storyName = 'List of batch changes'

export const NoBatchChanges: StoryFn = () => (
    <WebStory initialEntries={['/github.com/sourcegraph/awesome/-/batch-changes']}>
        {props => (
            <BatchChangeRepoPage
                {...props}
                repo={repoDefaults}
                authenticatedUser={mockAuthenticatedUser}
                isSourcegraphDotCom={false}
                queryRepoBatchChangeStats={queryEmptyRepoBatchChangeStats}
                queryRepoBatchChanges={queryNone}
            />
        )}
    </WebStory>
)

NoBatchChanges.storyName = 'No batch changes'
