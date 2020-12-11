import { storiesOf } from '@storybook/react'
import React from 'react'
import { Observable, of } from 'rxjs'
import { LsifIndexFields, LSIFIndexState } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { IndexConnection } from './backend'
import { CodeIntelIndexesPage } from './CodeIntelIndexesPage'

const { add } = storiesOf('web/codeintel/list/CodeIntelIndexesPage', module)
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

add('Page', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelIndexesPage
                {...props}
                now={now}
                fetchLsifIndexes={fetch(
                    {
                        id: '1',
                        state: LSIFIndexState.QUEUED,
                        queuedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: null,
                        finishedAt: null,
                        placeInQueue: 3,
                        failure: null,
                    },
                    {
                        id: '2',
                        state: LSIFIndexState.PROCESSING,
                        queuedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: '2020-06-15T12:25:30+00:00',
                        finishedAt: null,
                        failure: null,
                        placeInQueue: null,
                    },
                    {
                        id: '3',
                        state: LSIFIndexState.COMPLETED,
                        queuedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: '2020-06-15T12:25:30+00:00',
                        finishedAt: '2020-06-15T12:30:30+00:00',
                        failure: null,
                        placeInQueue: null,
                    },
                    {
                        id: '4',
                        state: LSIFIndexState.ERRORED,
                        queuedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: '2020-06-15T12:25:30+00:00',
                        finishedAt: '2020-06-15T12:30:30+00:00',
                        failure: 'Whoops! The server encountered a boo-boo handling this input.',
                        placeInQueue: null,
                    }
                )}
            />
        )}
    </EnterpriseWebStory>
))

const fetch = (
    ...indexes: Omit<
        LsifIndexFields,
        '__typename' | 'projectRoot' | 'inputCommit' | 'inputRoot' | 'inputIndexer' | 'steps'
    >[]
): (() => Observable<IndexConnection>) => () =>
    of({
        nodes: indexes.map(index => ({
            __typename: 'LSIFIndex',
            projectRoot: {
                url: '',
                path: 'web/',
                repository: {
                    url: '',
                    name: 'github.com/sourcegraph/sourcegraph',
                },
                commit: {
                    url: '',
                    oid: '9ea5e9f0e0344f8197622df6b36faf48ccd02570',
                    abbreviatedOID: '9ea5e9f',
                },
            },
            inputCommit: '9ea5e9f0e0344f8197622df6b36faf48ccd02570',
            inputRoot: 'web/',
            inputIndexer: 'lsif-tsc',
            steps: {
                setup: [executionLog],
                preIndex: [
                    { root: '/', image: 'node:alpine', commands: ['yarn'], logEntry: executionLog },
                    { root: '/web', image: 'node:alpine', commands: ['yarn'], logEntry: executionLog },
                ],
                index: {
                    indexerArgs: ['-p', '.'],
                    outfile: 'index.lsif',
                    logEntry: executionLog,
                },
                upload: executionLog,
                teardown: [executionLog],
            },
            ...index,
        })),
        totalCount: 10,
        pageInfo: {
            __typename: 'PageInfo',
            endCursor: 'fakenextpage',
            hasNextPage: true,
        },
    })

const now = () => new Date('2020-06-15T15:25:00+00:00')

const executionLog = {
    key: 'log',
    command: ['lsif-go', '-v'],
    startTime: '2020-06-15T15:25:00+00:00',
    exitCode: 0,
    out: 'foo\nbar\baz\n',
    durationMilliseconds: 123456,
}
