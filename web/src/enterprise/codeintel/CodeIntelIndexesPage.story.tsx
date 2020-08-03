import { CodeIntelIndexesPage } from './CodeIntelIndexesPage'
import { Index } from './backend'
import { of } from 'rxjs'
import { storiesOf } from '@storybook/react'
import { SuiteFunction } from 'mocha'
import * as GQL from '../../../../shared/src/graphql/schema'
import * as H from 'history'
import React from 'react'
import webStyles from '../../SourcegraphWebApp.scss'
import { SourcegraphContext } from '../../jscontext'

window.context = {} as SourcegraphContext & SuiteFunction

const { add } = storiesOf('web/CodeIntelIndexes', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="theme-light container">{story()}</div>
    </>
))

const history = H.createMemoryHistory()

const commonProps = {
    history,
    location: history.location,
    match: {
        params: { id: '' },
        isExact: true,
        path: '',
        url: '',
    },
    now: () => new Date('2020-06-15T15:25:00+00:00'),
}

const index: Pick<Index, 'id' | 'projectRoot' | 'inputCommit'> = {
    id: '1234',
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
}

add('List', () => (
    <CodeIntelIndexesPage
        {...commonProps}
        fetchLsifIndexes={() =>
            of({
                nodes: [
                    {
                        ...index,
                        state: GQL.LSIFIndexState.COMPLETED,
                        queuedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: '2020-06-15T12:25:30+00:00',
                        finishedAt: '2020-06-15T12:30:30+00:00',
                        failure: null,
                        placeInQueue: null,
                    },
                    {
                        ...index,
                        state: GQL.LSIFIndexState.ERRORED,
                        queuedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: '2020-06-15T12:25:30+00:00',
                        finishedAt: '2020-06-15T12:30:30+00:00',
                        failure: 'Whoops! The server encountered a boo-boo handling this input.',
                        placeInQueue: null,
                    },
                    {
                        ...index,
                        state: GQL.LSIFIndexState.PROCESSING,
                        queuedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: '2020-06-15T12:25:30+00:00',
                        finishedAt: null,
                        failure: null,
                        placeInQueue: null,
                    },
                    {
                        ...index,
                        state: GQL.LSIFIndexState.QUEUED,
                        queuedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: null,
                        finishedAt: null,
                        placeInQueue: 3,
                        failure: null,
                    },
                ],
                totalCount: 8,
                pageInfo: {
                    __typename: 'PageInfo',
                    endCursor: 'fakenextpage',
                    hasNextPage: true,
                },
            })
        }
    />
))
