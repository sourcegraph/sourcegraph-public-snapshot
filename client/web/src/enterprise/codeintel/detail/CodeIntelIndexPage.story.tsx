import { storiesOf } from '@storybook/react'
import React from 'react'
import { Observable, of } from 'rxjs'

import { LSIFUploadState } from '@sourcegraph/shared/src/graphql/schema'

import {
    ExecutionLogEntryFields,
    LsifIndexFields,
    LSIFIndexState,
    LsifIndexStepsFields,
} from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { CodeIntelIndexPage } from './CodeIntelIndexPage'

const { add } = storiesOf('web/codeintel/detail/CodeIntelIndexPage', module)
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

add('Queued', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelIndexPage
                {...props}
                fetchLsifIndex={fetch({
                    state: LSIFIndexState.QUEUED,
                    queuedAt: '2020-06-15T12:20:30+00:00',
                    startedAt: null,
                    finishedAt: null,
                    placeInQueue: 3,
                    failure: null,
                    associatedUpload: null,
                    steps,
                })}
                now={now}
            />
        )}
    </EnterpriseWebStory>
))

add('Processing', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelIndexPage
                {...props}
                fetchLsifIndex={fetch({
                    state: LSIFIndexState.PROCESSING,
                    queuedAt: '2020-06-15T12:20:30+00:00',
                    startedAt: '2020-06-15T12:25:30+00:00',
                    finishedAt: null,
                    failure: null,
                    placeInQueue: null,
                    associatedUpload: null,
                    steps,
                })}
                now={now}
            />
        )}
    </EnterpriseWebStory>
))

add('Completed', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelIndexPage
                {...props}
                fetchLsifIndex={fetch({
                    state: LSIFIndexState.COMPLETED,
                    queuedAt: '2020-06-15T12:20:30+00:00',
                    startedAt: '2020-06-15T12:25:30+00:00',
                    finishedAt: '2020-06-15T12:30:30+00:00',
                    failure: null,
                    placeInQueue: null,
                    associatedUpload: null,
                    steps,
                })}
                now={now}
            />
        )}
    </EnterpriseWebStory>
))

add('Errored', () => {
    const entry = steps.index.logEntry as ExecutionLogEntryFields
    const logEntry = { ...entry, exitCode: 127, out: `${entry.out}\nstderr: Failed to complete\n` }

    return (
        <EnterpriseWebStory>
            {props => (
                <CodeIntelIndexPage
                    {...props}
                    fetchLsifIndex={fetch({
                        state: LSIFIndexState.ERRORED,
                        queuedAt: '2020-06-15T12:20:30+00:00',
                        startedAt: '2020-06-15T12:25:30+00:00',
                        finishedAt: '2020-06-15T12:30:30+00:00',
                        failure: 'Whoops! The server encountered a boo-boo handling this input.',
                        placeInQueue: null,
                        associatedUpload: null,
                        steps: { ...steps, index: { ...steps.index, logEntry } },
                    })}
                    now={now}
                />
            )}
        </EnterpriseWebStory>
    )
})

add('Associated upload', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelIndexPage
                {...props}
                fetchLsifIndex={fetch({
                    state: LSIFIndexState.COMPLETED,
                    queuedAt: '2020-06-15T12:20:30+00:00',
                    startedAt: '2020-06-15T12:25:30+00:00',
                    finishedAt: '2020-06-15T12:30:30+00:00',
                    failure: null,
                    placeInQueue: null,
                    associatedUpload: {
                        id: '6789',
                        state: LSIFUploadState.QUEUED,
                        uploadedAt: '2020-06-15T12:28:30+00:00',
                        startedAt: null,
                        finishedAt: null,
                        placeInQueue: 5,
                    },
                    steps,
                })}
                now={now}
            />
        )}
    </EnterpriseWebStory>
))

const fetch = (
    index: Pick<
        LsifIndexFields,
        'state' | 'queuedAt' | 'startedAt' | 'finishedAt' | 'failure' | 'placeInQueue' | 'associatedUpload' | 'steps'
    >
): (() => Observable<LsifIndexFields>) => () =>
    of({
        __typename: 'LSIFIndex',
        id: '1234',
        projectRoot: {
            url: '',
            path: 'web',
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
        inputIndexer: 'sourcegraph/lsif-go:latest',
        ...index,
    })

const trim = (value: string) =>
    value
        .split('\n')
        .map(line => line.slice(16))
        .join('\n')
        .trim()

const steps: LsifIndexStepsFields = {
    setup: [
        {
            key: '',
            command: ['git', '-C', '/tmp/077915702', 'init'],
            out: trim(`
                stdout: Initialized empty Git repository in /tmp/077915702/.git/
            `),
            startTime: '2020-06-15T12:25:30+00:00',
            exitCode: 0,
            durationMilliseconds: 50123,
        },
        {
            key: '',
            command: [
                'git',
                '-C',
                '/tmp/077915702',
                '-c',
                'protocol.version=2',
                'fetch',
                'https://USERNAME_REMOVED:PASSWORD_REMOVED@sourcegraph.com/.USERNAME_REMOVEDs/git/github.com/kubernetes/kubernetes',
                'e1c617a88ec286f5f6cb2589d6ac562d095e1068',
            ],
            out: trim(`
                stderr: From https://sourcegraph.com/.USERNAME_REMOVEDs/git/github.com/kubernetes/kubernetes
                stderr: * branch                    e1c617a88ec286f5f6cb2589d6ac562d095e1068 -> FETCH_HEAD
            `),
            startTime: '2020-06-15T12:25:30+00:00',
            exitCode: 0,
            durationMilliseconds: 50123,
        },
        {
            key: '',
            command: ['git', '-C', '/tmp/077915702', 'checkout', 'e1c617a88ec286f5f6cb2589d6ac562d095e1068'],
            out: trim(`
                stderr: Note: switching to 'e1c617a88ec286f5f6cb2589d6ac562d095e1068'.
                stderr: You are in 'detached HEAD' state. You can look around, make experimental
                stderr: changes and commit them, and you can discard any commits you make in this
                stderr: state without impacting any branches by switching back to a branch.
                stderr: If you want to create a new branch to retain commits you create, you may
                stderr: do so (now or later) by using -c with the switch command. Example:
                stderr: git switch -c <new-branch-name>
                stderr: Or undo this operation with:
                stderr: git switch -
                stderr: Turn off this advice by setting config variable advice.detachedHead to false
                stdout: HEAD is now at e1c617a88ec Merge pull request #96874 from MikeSpreitzer/flaky/apnf-e2e-drown-test
            `),
            startTime: '2020-06-15T12:25:30+00:00',
            exitCode: 0,
            durationMilliseconds: 50123,
        },
    ],
    preIndex: [{ root: '/web', image: 'node:alpine', commands: ['yarn'], logEntry: null }],
    index: {
        indexerArgs: ['lsif-go', '--no-animation'],
        outfile: 'index.lsif',
        logEntry: {
            key: 'log',
            command: [
                'ignite',
                'exec',
                'b6683f7c-8b35-436e-8973-062db8ca37b7',
                '--',
                'docker',
                'run',
                '--rm',
                '--cpus',
                '4',
                '--memory',
                '12G',
                '-v',
                '/work:/data',
                '-w',
                '/data/web',
                'sourcegraph/lsif-go:latest',
                'lsif-go',
                '--no-animation',
            ],
            out: trim(`
                stdout: Loading packages
                stdout: Emitting documents
                stdout: Adding import definitions
                stdout: Indexing definitions
                stdout: Indexing references
                stdout: Linking items to definitions
                stdout: Emitting contains relations
            `),
            startTime: '2020-06-15T12:25:30+00:00',
            exitCode: 0,
            durationMilliseconds: 50123,
        },
    },
    upload: null,
    teardown: [],
}

const now = () => new Date('2020-06-15T15:25:00+00:00')
