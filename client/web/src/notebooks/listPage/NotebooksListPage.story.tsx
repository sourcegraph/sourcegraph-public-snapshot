import type { DecoratorFn, Story, Meta } from '@storybook/react'
import { subDays } from 'date-fns'
import { type Observable, of } from 'rxjs'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'
import type { ListNotebooksResult } from '../../graphql-operations'

import { NotebooksListPage } from './NotebooksListPage'

const now = new Date()

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/search/notebooks/listPage/NotebooksListPage',
    parameters: {
        chromatic: { disableSnapshots: false },
    },
    decorators: [decorator],
}

export default config

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

export const Default: Story = () => (
    <WebStory>
        {props => (
            <NotebooksListPage
                {...props}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                authenticatedUser={null}
                fetchNotebooks={fetchNotebooks}
            />
        )}
    </WebStory>
)
