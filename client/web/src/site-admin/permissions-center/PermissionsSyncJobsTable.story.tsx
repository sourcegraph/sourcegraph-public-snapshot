import type { DecoratorFn, Meta, Story } from '@storybook/react'
import { addMinutes, formatRFC3339, subMinutes } from 'date-fns'
import { WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import {
    CodeHostStatus,
    ExternalServiceKind,
    PermissionsSyncJobPriority,
    PermissionsSyncJobReason,
    PermissionsSyncJobReasonGroup,
    PermissionsSyncJobState,
} from '@sourcegraph/shared/src/graphql-operations'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'
import type { PermissionsSyncJob } from '../../graphql-operations'

import { PERMISSIONS_SYNC_JOBS_QUERY, PERMISSIONS_SYNC_JOBS_STATS } from './backend'
import { PermissionsSyncJobsTable } from './PermissionsSyncJobsTable'

const decorator: DecoratorFn = Story => <Story />

const config: Meta = {
    title: 'web/src/site-admin/permissions-center/PermissionsSyncJobsTable',
    decorators: [decorator],
}

export default config

const TIMESTAMP_MOCK = subMinutes(Date.now(), 5)
const JOBS_MOCK_DATA = getSyncJobs()
const MANUAL_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(job => job.reason.group === PermissionsSyncJobReasonGroup.MANUAL)
const SCHEDULE_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(
    job => job.reason.group === PermissionsSyncJobReasonGroup.SCHEDULE
)
const SG_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(job => job.reason.group === PermissionsSyncJobReasonGroup.SOURCEGRAPH)
const WEBHOOK_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(job => job.reason.group === PermissionsSyncJobReasonGroup.WEBHOOK)

const CANCELED_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(job => job.state === PermissionsSyncJobState.CANCELED)
const COMPLETED_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(
    job => job.state === PermissionsSyncJobState.COMPLETED && !job.partialSuccess
)
const PARTIAL_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(job => job.partialSuccess)
const ERRORED_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(job => job.state === PermissionsSyncJobState.ERRORED)
const FAILED_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(job => job.state === PermissionsSyncJobState.FAILED)
const PROCESSING_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(job => job.state === PermissionsSyncJobState.PROCESSING)
const QUEUED_JOBS_MOCK_DATA = JOBS_MOCK_DATA.filter(job => job.state === PermissionsSyncJobState.QUEUED)

export const SixSyncJobsFound: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        generateResponse(null, null, JOBS_MOCK_DATA, 20),
                        generateResponse(null, PermissionsSyncJobReasonGroup.MANUAL, MANUAL_JOBS_MOCK_DATA, 6),
                        generateResponse(null, PermissionsSyncJobReasonGroup.SCHEDULE, SCHEDULE_JOBS_MOCK_DATA, 6),
                        generateResponse(null, PermissionsSyncJobReasonGroup.SOURCEGRAPH, SG_JOBS_MOCK_DATA, 3),
                        generateResponse(null, PermissionsSyncJobReasonGroup.WEBHOOK, WEBHOOK_JOBS_MOCK_DATA, 8),
                        generateResponse(PermissionsSyncJobState.CANCELED, null, CANCELED_JOBS_MOCK_DATA, 4),
                        generateResponse(PermissionsSyncJobState.COMPLETED, null, COMPLETED_JOBS_MOCK_DATA, 2),
                        generateResponse(PermissionsSyncJobState.ERRORED, null, ERRORED_JOBS_MOCK_DATA, 3),
                        generateResponse(PermissionsSyncJobState.FAILED, null, FAILED_JOBS_MOCK_DATA, 3),
                        generateResponse(PermissionsSyncJobState.PROCESSING, null, PROCESSING_JOBS_MOCK_DATA, 3),
                        generateResponse(PermissionsSyncJobState.QUEUED, null, QUEUED_JOBS_MOCK_DATA, 3),
                        generateResponse(null, null, PARTIAL_JOBS_MOCK_DATA, 2, true),
                        {
                            request: {
                                query: getDocumentNode(PERMISSIONS_SYNC_JOBS_STATS),
                                variables: {},
                            },
                            result: {
                                data: {
                                    permissionsSyncingStats: {
                                        queueSize: 1337,
                                        usersWithLatestJobFailing: 228101,
                                        reposWithLatestJobFailing: 3,
                                        usersWithNoPermissions: 4,
                                        reposWithNoPermissions: 5,
                                        usersWithStalePermissions: 6,
                                        reposWithStalePermissions: 42,
                                    },
                                },
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                    ])
                }
            >
                <PermissionsSyncJobsTable
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

SixSyncJobsFound.storyName = 'Six sync jobs'

interface repo {
    __typename: 'Repository'
    id: string
    name: string
    url: string
    externalRepository: {
        serviceType: ExternalServiceKind
        serviceID: string
    }
}

interface user {
    __typename: 'User'
    id: string
    username: string
    displayName: string | null
    email: string
    avatarURL: string | null
}

type subject = repo | user

interface reason {
    __typename?: 'PermissionsSyncJobReasonWithGroup'
    group: PermissionsSyncJobReasonGroup
    reason: PermissionsSyncJobReason
}

function getSyncJobs(): PermissionsSyncJob[] {
    const jobs: PermissionsSyncJob[] = []

    for (let index = 0; index < 20; index++) {
        let state: PermissionsSyncJobState
        let reason: reason
        switch (index % 6) {
            case 0:
                state = PermissionsSyncJobState.CANCELED
                reason = {
                    group: PermissionsSyncJobReasonGroup.WEBHOOK,
                    reason: PermissionsSyncJobReason.REASON_GITHUB_REPO_EVENT,
                }
                break
            case 1:
                state = PermissionsSyncJobState.COMPLETED
                reason = {
                    group: PermissionsSyncJobReasonGroup.WEBHOOK,
                    reason: PermissionsSyncJobReason.REASON_GITHUB_REPO_EVENT,
                }
                break
            case 2:
                state = PermissionsSyncJobState.ERRORED
                reason = {
                    group: PermissionsSyncJobReasonGroup.MANUAL,
                    reason: PermissionsSyncJobReason.REASON_MANUAL_REPO_SYNC,
                }
                break
            case 3:
                state = PermissionsSyncJobState.FAILED
                reason = {
                    group: PermissionsSyncJobReasonGroup.MANUAL,
                    reason: PermissionsSyncJobReason.REASON_MANUAL_USER_SYNC,
                }
                break
            case 4:
                state = PermissionsSyncJobState.PROCESSING
                reason = {
                    group: PermissionsSyncJobReasonGroup.SCHEDULE,
                    reason: PermissionsSyncJobReason.REASON_REPO_OUTDATED_PERMS,
                }
                break
            case 5:
            default:
                state = PermissionsSyncJobState.QUEUED
                reason = {
                    group: PermissionsSyncJobReasonGroup.SOURCEGRAPH,
                    reason: PermissionsSyncJobReason.REASON_USER_EMAIL_VERIFIED,
                }
                break
        }

        const subject: subject =
            index % 2 === 0
                ? {
                      __typename: 'Repository',
                      id: index.toString(),
                      name: `sourcegraph/repo-${index}`,
                      url: `/ghe.sgdev.org/milton/repo-${index}/`,
                      externalRepository: {
                          serviceType: index % 3 === 0 ? ExternalServiceKind.GITHUB : ExternalServiceKind.GITLAB,
                          serviceID: index % 3 === 0 ? 'github.com' : 'gitlab.com',
                      },
                  }
                : {
                      __typename: 'User',
                      id: index.toString(),
                      username: `username-${index}`,
                      displayName: 'Test User',
                      email: 'example@sourcegraph.com',
                      avatarURL: null,
                  }

        jobs.push(
            createSyncJobMock(
                index.toString(),
                state,
                subject,
                reason,
                state === PermissionsSyncJobState.COMPLETED && index > 10,
                index % 4 === 0 ? 0 : index + 10,
                index % 4 === 0 ? 0 : index + 5
            )
        )
    }
    return jobs
}

function generateResponse(
    state: PermissionsSyncJobState | null,
    reasonGroup: PermissionsSyncJobReasonGroup | null,
    jobs: PermissionsSyncJob[],
    count: number,
    partial: boolean = false
) {
    return {
        request: {
            query: getDocumentNode(PERMISSIONS_SYNC_JOBS_QUERY),
            variables: {
                first: 20,
                last: null,
                after: null,
                before: null,
                reasonGroup: reasonGroup ?? null,
                state: state ?? null,
                searchType: null,
                query: '',
                partial,
            },
        },
        result: {
            data: {
                permissionsSyncJobs: {
                    nodes: jobs,
                    totalCount: count,
                    pageInfo: {
                        hasNextPage: true,
                        hasPreviousPage: false,
                        startCursor: null,
                        endCursor: null,
                    },
                },
            },
        },
        nMatches: Number.POSITIVE_INFINITY,
    }
}

function createSyncJobMock(
    id: string,
    state: PermissionsSyncJobState,
    subject: subject,
    reason: reason,
    partial: boolean = false,
    permissionsAdded: number = 1337,
    permissionsRemoved: number = 42
): PermissionsSyncJob {
    return {
        __typename: 'PermissionsSyncJob',
        id,
        state,
        subject,
        reason,
        triggeredByUser: {
            username: 'super-site-admin',
        },
        queuedAt: formatRFC3339(TIMESTAMP_MOCK),
        startedAt: state !== PermissionsSyncJobState.QUEUED ? formatRFC3339(addMinutes(TIMESTAMP_MOCK, 1)) : null,
        finishedAt:
            state !== PermissionsSyncJobState.QUEUED && state !== PermissionsSyncJobState.PROCESSING
                ? formatRFC3339(addMinutes(TIMESTAMP_MOCK, 2))
                : null,
        processAfter: null,
        permissionsAdded,
        permissionsRemoved,
        permissionsFound: permissionsAdded + permissionsRemoved,
        failureMessage: null,
        cancellationReason: null,
        ranForMs: null,
        numResets: null,
        numFailures: null,
        lastHeartbeatAt: null,
        workerHostname: 'worker-hostname',
        cancel: false,
        priority: PermissionsSyncJobPriority.LOW,
        noPerms: false,
        invalidateCaches: false,
        codeHostStates: partial
            ? [
                  {
                      __typename: 'CodeHostState',
                      providerID: '1',
                      providerType: 'github',
                      status: CodeHostStatus.SUCCESS,
                      message: 'success!',
                  },
                  {
                      __typename: 'CodeHostState',
                      providerID: '1',
                      providerType: 'gitlab',
                      status: CodeHostStatus.ERROR,
                      message: 'error!',
                  },
              ]
            : [],
        partialSuccess: partial,
        placeInQueue: 1,
    }
}
