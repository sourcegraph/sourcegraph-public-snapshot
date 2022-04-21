import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import { Observable, of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'
import { ListNotebooksResult } from '../../graphql-operations'

import { NotebooksListPage } from './NotebooksListPage'

const { add } = storiesOf('web/search/notebooks/listPage/NotebooksListPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({ chromatic: { disableSnapshots: false } })

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
                viewerHasStarred: true,
                stars: { totalCount: 123 },
                creator: { __typename: 'User', username: 'user1' },
                updater: { __typename: 'User', username: 'user1' },
                namespace: { __typename: 'User', namespaceName: 'user1', id: '1' },
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
                viewerHasStarred: true,
                stars: { totalCount: 123 },
                creator: { __typename: 'User', username: 'user2' },
                updater: { __typename: 'User', username: 'user2' },
                namespace: { __typename: 'User', namespaceName: 'user2', id: '2' },
                blocks: [{ __typename: 'MarkdownBlock', id: '1', markdownInput: '# Title' }],
            },
        ],
        pageInfo: { hasNextPage: false, endCursor: null },
    })

add('default', () => (
    <WebStory>
        {props => (
            <NotebooksListPage
                {...props}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                authenticatedUser={null}
                fetchNotebooks={fetchNotebooks}
            />
        )}
    </WebStory>
))
