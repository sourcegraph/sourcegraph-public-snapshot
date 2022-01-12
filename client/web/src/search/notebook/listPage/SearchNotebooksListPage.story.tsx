import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'
import { Observable, of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../components/WebStory'
import { ListNotebooksResult } from '../../../graphql-operations'

import { SearchNotebooksListPage } from './SearchNotebooksListPage'

const { add } = storiesOf('web/search/notebook/SearchNotebooksListPage', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const now = new Date()

const fetchNotebooks = (): Observable<ListNotebooksResult['notebooks']> =>
    of({
        totalCount: 2,
        nodes: [
            {
                __typename: 'Notebook',
                id: '1',
                title: 'Notebook Title 1',
                createdAt: subDays(now, 5).toISOString(),
                updatedAt: subDays(now, 2).toISOString(),
                public: true,
                viewerCanManage: true,
                creator: { __typename: 'User', username: 'user1' },
                blocks: [
                    { __typename: 'MarkdownBlock', id: '1', markdownInput: '# Title' },
                    { __typename: 'QueryBlock', id: '2', queryInput: 'query' },
                ],
            },
            {
                __typename: 'Notebook',
                id: '2',
                title: 'Notebook Title 2',
                createdAt: subDays(now, 5).toISOString(),
                updatedAt: subDays(now, 1).toISOString(),
                public: true,
                viewerCanManage: true,
                creator: { __typename: 'User', username: 'user2' },
                blocks: [{ __typename: 'MarkdownBlock', id: '1', markdownInput: '# Title' }],
            },
        ],
        pageInfo: { hasNextPage: false, endCursor: null },
    })

add('default', () => (
    <WebStory>
        {props => (
            <SearchNotebooksListPage
                {...props}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                authenticatedUser={null}
                fetchNotebooks={fetchNotebooks}
            />
        )}
    </WebStory>
))
