import { storiesOf } from '@storybook/react'
import * as H from 'history'
import { SuiteFunction } from 'mocha'
import React from 'react'
import { of } from 'rxjs'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { LsifIndexFields, LSIFIndexState } from '../../../graphql-operations'
import { SourcegraphContext } from '../../../jscontext'
import webStyles from '../../../SourcegraphWebApp.scss'
import { CodeIntelIndexPage } from './CodeIntelIndexPage'

window.context = {} as SourcegraphContext & SuiteFunction

const { add } = storiesOf('web/Codeintel administration/CodeIntelIndex', module).addDecorator(story => (
    <>
        <div className="container">{story()}</div>
        <style>{webStyles}</style>
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
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

const executionLog = {
    key: 'log',
    command: ['lsif-go', '-v'],
    startTime: '2020-06-15T15:25:00+00:00',
    exitCode: 0,
    out: 'foo\nbar\baz\n',
    durationMilliseconds: 123456,
}

const index: Omit<LsifIndexFields, 'state' | 'queuedAt' | 'startedAt' | 'finishedAt' | 'failure' | 'placeInQueue'> = {
    __typename: 'LSIFIndex',
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
}

add('Completed', () => (
    <CodeIntelIndexPage
        {...commonProps}
        fetchLsifIndex={() =>
            of({
                ...index,
                state: LSIFIndexState.COMPLETED,
                queuedAt: '2020-06-15T12:20:30+00:00',
                startedAt: '2020-06-15T12:25:30+00:00',
                finishedAt: '2020-06-15T12:30:30+00:00',
                failure: null,
                placeInQueue: null,
            })
        }
    />
))

add('Errored', () => (
    <CodeIntelIndexPage
        {...commonProps}
        fetchLsifIndex={() =>
            of({
                ...index,
                state: LSIFIndexState.ERRORED,
                queuedAt: '2020-06-15T12:20:30+00:00',
                startedAt: '2020-06-15T12:25:30+00:00',
                finishedAt: '2020-06-15T12:30:30+00:00',
                failure: 'Whoops! The server encountered a boo-boo handling this input.',
                placeInQueue: null,
            })
        }
    />
))

add('Processing', () => (
    <CodeIntelIndexPage
        {...commonProps}
        fetchLsifIndex={() =>
            of({
                ...index,
                state: LSIFIndexState.PROCESSING,
                queuedAt: '2020-06-15T12:20:30+00:00',
                startedAt: '2020-06-15T12:25:30+00:00',
                finishedAt: null,
                failure: null,
                placeInQueue: null,
            })
        }
    />
))

add('Queued', () => (
    <CodeIntelIndexPage
        {...commonProps}
        fetchLsifIndex={() =>
            of({
                ...index,
                state: LSIFIndexState.QUEUED,
                queuedAt: '2020-06-15T12:20:30+00:00',
                startedAt: null,
                finishedAt: null,
                placeInQueue: 3,
                failure: null,
            })
        }
    />
))
