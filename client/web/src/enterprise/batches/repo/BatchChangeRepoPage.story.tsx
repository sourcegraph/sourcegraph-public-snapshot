import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { RepoBatchChange, RepositoryFields } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'

import { queryRepoBatchChanges as _queryRepoBatchChanges } from './backend'
import { BatchChangeRepoPage } from './BatchChangeRepoPage'
import { NODES } from './testData'

const { add } = storiesOf('web/batches/BatchChangeRepoPage', module)
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
    <EnterpriseWebStory initialEntries={['/github.com/sourcegraph/awesome/-/batch-changes']}>
        {props => (
            <BatchChangeRepoPage
                {...props}
                repo={repoDefaults}
                queryRepoBatchChanges={queryList}
                queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
            />
        )}
    </EnterpriseWebStory>
))

add('No batch changes', () => (
    <EnterpriseWebStory initialEntries={['/github.com/sourcegraph/awesome/-/batch-changes']}>
        {props => <BatchChangeRepoPage {...props} repo={repoDefaults} queryRepoBatchChanges={queryNone} />}
    </EnterpriseWebStory>
))
