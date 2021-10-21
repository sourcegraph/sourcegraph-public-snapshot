import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../components/WebStory'
import { RepoBatchChange, RepositoryFields } from '../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'

import {
    queryRepoBatchChanges as _queryRepoBatchChanges,
    queryRepoBatchChangeStats as _queryRepoBatchChangeStats,
} from './backend'
import { BatchChangeRepoPage } from './BatchChangeRepoPage'
import { NODES } from './testData'

const { add } = storiesOf('web/batches/repo/BatchChangeRepoPage', module)
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const repoDefaults: RepositoryFields = {
    description: 'An awesome repo!',
    defaultBranch: null,
    viewerCanAdminister: false,
    externalURLs: [],
    id: 'repoid',
    name: 'github.com/sourcegraph/awesome',
    url: 'http://test.test/awesome',
}

const queryRepoBatchChangeStats: typeof _queryRepoBatchChangeStats = () =>
    of({
        batchChangesDiffStat: {
            added: 247,
            changed: 1896,
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
            changed: 0,
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

const queryRepoBatchChanges = (nodes: RepoBatchChange[]): typeof _queryRepoBatchChanges => () =>
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

add('List of batch changes', () => (
    <WebStory initialEntries={['/github.com/sourcegraph/awesome/-/batch-changes']}>
        {props => (
            <BatchChangeRepoPage
                {...props}
                repo={repoDefaults}
                queryRepoBatchChangeStats={queryRepoBatchChangeStats}
                queryRepoBatchChanges={queryList}
                queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
            />
        )}
    </WebStory>
))

add('No batch changes', () => (
    <WebStory initialEntries={['/github.com/sourcegraph/awesome/-/batch-changes']}>
        {props => (
            <BatchChangeRepoPage
                {...props}
                repo={repoDefaults}
                queryRepoBatchChangeStats={queryEmptyRepoBatchChangeStats}
                queryRepoBatchChanges={queryNone}
            />
        )}
    </WebStory>
))
