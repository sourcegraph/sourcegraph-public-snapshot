import { Meta, Story } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../components/WebStory'
import { ExecutionLogEntryFields, LsifIndexFields, LSIFIndexState } from '../../../graphql-operations'

import { CodeIntelIndexesPage, CodeIntelIndexesPageProps } from './CodeIntelIndexesPage'

const executionLogPrototype: ExecutionLogEntryFields = {
    key: 'log',
    command: ['lsif-go', '-v'],
    startTime: '2020-06-15T15:25:00+00:00',
    exitCode: 0,
    out: 'foo\nbar\baz\n',
    durationMilliseconds: 123456,
}

const indexPrototype: Omit<LsifIndexFields, 'id' | 'state' | 'queuedAt'> = {
    __typename: 'LSIFIndex',
    inputCommit: '',
    inputRoot: 'web/',
    inputIndexer: 'lsif-tsc',
    failure: null,
    startedAt: null,
    finishedAt: null,
    placeInQueue: null,
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
    steps: {
        setup: [executionLogPrototype],
        preIndex: [
            { root: '/', image: 'node:alpine', commands: ['yarn'], logEntry: executionLogPrototype },
            { root: '/web', image: 'node:alpine', commands: ['yarn'], logEntry: executionLogPrototype },
        ],
        index: {
            indexerArgs: ['-p', '.'],
            outfile: 'index.lsif',
            logEntry: executionLogPrototype,
        },
        upload: executionLogPrototype,
        teardown: [executionLogPrototype],
    },
    associatedUpload: null,
}

const testIndexes: LsifIndexFields[] = [
    {
        ...indexPrototype,
        id: '4',
        state: LSIFIndexState.QUEUED,
        queuedAt: '2020-06-15T12:20:30+00:00',
        placeInQueue: 1,
    },
    {
        ...indexPrototype,
        id: '3',
        state: LSIFIndexState.PROCESSING,
        queuedAt: '2020-06-15T12:20:30+00:00',
        startedAt: '2020-06-15T12:25:30+00:00',
    },
    {
        ...indexPrototype,
        id: '2',
        state: LSIFIndexState.COMPLETED,
        queuedAt: '2020-06-14T12:20:30+00:00',
        startedAt: '2020-06-14T12:25:30+00:00',
        finishedAt: '2020-06-14T12:30:30+00:00',
    },
    {
        ...indexPrototype,
        id: '1',
        state: LSIFIndexState.ERRORED,
        queuedAt: '2020-06-13T12:20:30+00:00',
        startedAt: '2020-06-13T12:25:30+00:00',
        finishedAt: '2020-06-13T12:30:30+00:00',
        failure: 'Whoops! The server encountered a boo-boo handling this input.',
    },
]

const now = () => new Date('2020-06-15T15:25:00+00:00')

const makeResponse = (indexes: LsifIndexFields[]) => ({
    nodes: indexes,
    totalCount: indexes.length,
    pageInfo: {
        __typename: 'PageInfo',
        endCursor: null,
        hasNextPage: false,
    },
})

const story: Meta = {
    title: 'web/codeintel/list/CodeIntelIndexesPage',
    decorators: [story => <div className="p-3 container">{story()}</div>],
    parameters: {
        component: CodeIntelIndexesPage,
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}
export default story

const Template: Story<CodeIntelIndexesPageProps> = args => (
    <WebStory>{props => <CodeIntelIndexesPage {...props} {...args} />}</WebStory>
)

const defaults: Partial<CodeIntelIndexesPageProps> = {
    now,
    queryLsifIndexListByRepository: () => of(makeResponse([])),
    queryLsifIndexList: () => of(makeResponse([])),
}

export const EmptyGlobalPage = Template.bind({})
EmptyGlobalPage.args = {
    ...defaults,
}

export const GlobalPage = Template.bind({})
GlobalPage.args = {
    ...defaults,
    queryLsifIndexListByRepository: () => of(makeResponse(testIndexes)),
    queryLsifIndexList: () => of(makeResponse(testIndexes)),
}

export const EmptyRepositoryPage = Template.bind({})
EmptyRepositoryPage.args = {
    ...defaults,
    repo: { id: 'sourcegraph' },
}

export const RepositoryPage = Template.bind({})
RepositoryPage.args = {
    ...defaults,
    repo: { id: 'sourcegraph' },
    queryLsifIndexListByRepository: () => of(makeResponse(testIndexes)),
    queryLsifIndexList: () => of(makeResponse(testIndexes)),
}
