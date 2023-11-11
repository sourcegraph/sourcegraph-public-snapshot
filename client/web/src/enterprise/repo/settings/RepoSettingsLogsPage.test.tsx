import type { MockedResponse } from '@apollo/client/testing'
import { render, waitFor, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { describe, expect, test } from 'vitest'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import {
    ExternalServiceKind,
    type SettingsAreaRepositoryFields,
    type RepositoryRecordedCommandsResult,
} from '../../../graphql-operations'

import { REPOSITORY_RECORDED_COMMANDS_QUERY, REPOSITORY_RECORDED_COMMANDS_LIMIT } from './backend'
import { RepoSettingsLogsPage } from './RepoSettingsLogsPage'

const repositoryID = 'UmVwb3NpdG9yeToyNw=='
const mockRepo: SettingsAreaRepositoryFields = {
    id: repositoryID,
    name: 'gerrit.sgdev.org/a/gabe/repo',
    url: '/gerrit.sgdev.org/a/gabe/repo',
    isPrivate: false,
    viewerCanAdminister: true,
    mirrorInfo: {
        remoteURL: 'https://gerrit.sgdev.org/a/gabe/repo',
        cloneInProgress: false,
        cloneProgress: '',
        cloned: true,
        updatedAt: '2023-07-31T10:24:00Z',
        isCorrupted: false,
        corruptionLogs: [],
        lastError: '',
        lastSyncOutput: '',
        updateSchedule: {
            due: '2023-07-31T18:31:07Z',
            index: 29,
            total: 41,
        },
        updateQueue: {
            updating: true,
            index: 0,
            total: 1,
        },
        shard: '127.0.0.1:3501',
    },
    externalServices: {
        nodes: [
            {
                id: 'RXh0ZXJuYWxTZXJ2aWNlOjU=',
                kind: ExternalServiceKind.GERRIT,
                displayName: 'GERRIT #1',
                supportsRepoExclusion: false,
            },
        ],
    },
    metadata: [],
}

describe('RepoSettingsLogsPage', () => {
    test('should render correctly when there are recorded commands', async () => {
        const mockRecordedCommandsQuery: MockedResponse<RepositoryRecordedCommandsResult> = {
            request: {
                query: getDocumentNode(REPOSITORY_RECORDED_COMMANDS_QUERY),
                variables: {
                    id: repositoryID,
                    offset: 0,
                    limit: REPOSITORY_RECORDED_COMMANDS_LIMIT,
                },
            },
            result: {
                data: {
                    node: {
                        __typename: 'Repository',
                        isRecordingEnabled: true,
                        recordedCommands: {
                            nodes: [
                                {
                                    path: '/usr/bin/git',
                                    start: '2023-07-31T11:18:36Z',
                                    duration: 0.010911709,
                                    command: 'git cat-file -t -- 4b825dc642cb6eb9a060e54bf8d69288fbee4904',
                                    dir: '/Users/randomuser/.sourcegraph/repos_1/gerrit.sgdev.org/a/gabe/repo/.git',
                                    __typename: 'RecordedCommand',
                                    isSuccess: false,
                                    output: '',
                                },
                                {
                                    path: '/usr/bin/git',
                                    start: '2023-07-31T11:18:05Z',
                                    duration: 0.010100125,
                                    command: 'git config --get sourcegraph.type',
                                    dir: '/Users/randomuser/.sourcegraph/repos_1/gerrit.sgdev.org/a/gabe/repo/.git',
                                    __typename: 'RecordedCommand',
                                    isSuccess: true,
                                    output: 'git',
                                },
                            ],
                            totalCount: 2,
                            pageInfo: {
                                hasNextPage: false,
                            },
                        },
                    },
                },
            },
        }

        const cmp = render(
            <BrowserRouter>
                <MockedTestProvider mocks={[mockRecordedCommandsQuery]}>
                    <RepoSettingsLogsPage repo={mockRepo} />
                </MockedTestProvider>
            </BrowserRouter>
        )

        await waitFor(() => {
            expect(screen.queryByRole('img', { name: /loading/i })).not.toBeInTheDocument()
        })
        expect(cmp.asFragment()).toMatchSnapshot()
    })

    test('should render correctly when there are no recorded commands', async () => {
        const mockRecordedCommandsQuery: MockedResponse<RepositoryRecordedCommandsResult> = {
            delay: 0,
            request: {
                query: getDocumentNode(REPOSITORY_RECORDED_COMMANDS_QUERY),
                variables: {
                    id: repositoryID,
                    offset: 0,
                    limit: REPOSITORY_RECORDED_COMMANDS_LIMIT,
                },
            },
            result: {
                data: {
                    node: {
                        __typename: 'Repository',
                        isRecordingEnabled: true,
                        recordedCommands: {
                            nodes: [],
                            totalCount: 0,
                            pageInfo: {
                                hasNextPage: false,
                            },
                        },
                    },
                },
            },
        }

        const cmp = render(
            <BrowserRouter>
                <MockedTestProvider mocks={[mockRecordedCommandsQuery]}>
                    <RepoSettingsLogsPage repo={mockRepo} />
                </MockedTestProvider>
            </BrowserRouter>
        )
        await waitFor(() => {
            expect(screen.queryByRole('img', { name: /loading/i })).not.toBeInTheDocument()
        })

        expect(cmp.asFragment()).toMatchSnapshot()
    })

    test('should render a warning when recording is disabled', async () => {
        const mockRecordedCommandsQuery: MockedResponse<RepositoryRecordedCommandsResult> = {
            delay: 0,
            request: {
                query: getDocumentNode(REPOSITORY_RECORDED_COMMANDS_QUERY),
                variables: {
                    id: repositoryID,
                    offset: 0,
                    limit: REPOSITORY_RECORDED_COMMANDS_LIMIT,
                },
            },
            result: {
                data: {
                    node: {
                        __typename: 'Repository',
                        isRecordingEnabled: false,
                        recordedCommands: {
                            nodes: [],
                            totalCount: 0,
                            pageInfo: {
                                hasNextPage: false,
                            },
                        },
                    },
                },
            },
        }

        const cmp = render(
            <BrowserRouter>
                <MockedTestProvider mocks={[mockRecordedCommandsQuery]}>
                    <RepoSettingsLogsPage repo={mockRepo} />
                </MockedTestProvider>
            </BrowserRouter>
        )
        await waitFor(() => {
            expect(screen.queryByRole('img', { name: /loading/i })).not.toBeInTheDocument()
        })

        expect(cmp.asFragment()).toMatchSnapshot()
    })
})
