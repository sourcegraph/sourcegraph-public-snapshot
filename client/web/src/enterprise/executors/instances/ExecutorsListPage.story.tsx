import type { Decorator, StoryFn, Meta } from '@storybook/react'
import { subDays, subHours } from 'date-fns'
import { type Observable, of } from 'rxjs'

import { WebStory } from '../../../components/WebStory'
import { ExecutorCompatibility, type ExecutorConnectionFields } from '../../../graphql-operations'

import { ExecutorsListPage } from './ExecutorsListPage'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/instances/ExecutorsListPage',
    decorators: [decorator],
}

export default config

const listExecutorsQuery: () => Observable<ExecutorConnectionFields> = () =>
    of({
        pageInfo: { hasNextPage: false, endCursor: null },
        totalCount: 2,
        nodes: [
            {
                __typename: 'Executor',
                active: true,
                architecture: 'amd64',
                compatibility: ExecutorCompatibility.UP_TO_DATE,
                dockerVersion: '20.0.4',
                executorVersion: '4.1.0',
                firstSeenAt: subDays(new Date(), 1).toISOString(),
                gitVersion: '2.38.0',
                hostname: 'executor1.sgdev.org',
                id: 'ID1',
                igniteVersion: '0.10.5',
                lastSeenAt: new Date().toISOString(),
                os: 'linux',
                queueName: 'batches',
                srcCliVersion: '4.1.0',
                queueNames: [],
            },
            {
                __typename: 'Executor',
                active: false,
                architecture: 'amd64',
                compatibility: ExecutorCompatibility.OUTDATED,
                dockerVersion: '20.0.4',
                executorVersion: '4.1.0',
                firstSeenAt: subDays(new Date(), 1).toISOString(),
                gitVersion: '2.38.0',
                hostname: 'executor1.sgdev.org',
                id: 'ID2',
                igniteVersion: '0.10.5',
                lastSeenAt: subHours(new Date(), 5).toISOString(),
                os: 'linux',
                queueName: 'batches',
                srcCliVersion: '4.1.0',
                queueNames: [],
            },
        ],
    })

export const List: StoryFn = () => (
    <WebStory>{props => <ExecutorsListPage {...props} queryExecutors={listExecutorsQuery} />}</WebStory>
)

List.storyName = 'List of executors'

const emptyExecutorsQuery: () => Observable<ExecutorConnectionFields> = () =>
    of({
        pageInfo: { hasNextPage: false, endCursor: null },
        totalCount: 0,
        nodes: [],
    })

export const EmptyList: StoryFn = () => (
    <WebStory>{props => <ExecutorsListPage {...props} queryExecutors={emptyExecutorsQuery} />}</WebStory>
)

EmptyList.storyName = 'No executors'
