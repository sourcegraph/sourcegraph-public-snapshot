import React from 'react'

import { mdiAccount, mdiCloudQuestion } from '@mdi/js'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Badge, BADGE_VARIANTS, Icon, Text, Tooltip } from '@sourcegraph/wildcard'

import { PermissionsSyncJob, PermissionsSyncJobReasonGroup, PermissionsSyncJobState } from '../../graphql-operations'

import styles from './PermissionsSyncJobsTableItem.module.scss'

export interface ChangesetCloseNodeProps {
    node: PermissionsSyncJob
}

interface JobStateMetadata {
    badgeVariant: typeof BADGE_VARIANTS[number]
    temporalWording: string
    timeGetter: (job: PermissionsSyncJob) => string
}

const JOB_STATE_METADATA_MAPPING: Record<PermissionsSyncJobState, JobStateMetadata> = {
    QUEUED: {
        badgeVariant: 'secondary',
        temporalWording: 'Queued',
        timeGetter: job => job.queuedAt,
    },
    PROCESSING: {
        badgeVariant: 'primary',
        temporalWording: 'Began processing',
        timeGetter: job => job.startedAt ?? '',
    },
    COMPLETED: {
        badgeVariant: 'success',
        temporalWording: 'Completed',
        timeGetter: job => job.finishedAt ?? '',
    },
    ERRORED: {
        badgeVariant: 'danger',
        temporalWording: 'Errored',
        timeGetter: job => job.finishedAt ?? '',
    },
    FAILED: {
        badgeVariant: 'danger',
        temporalWording: 'Failed',
        timeGetter: job => job.finishedAt ?? '',
    },
    CANCELED: {
        badgeVariant: 'outlineSecondary',
        temporalWording: 'Canceled',
        timeGetter: job => job.finishedAt ?? '',
    },
}

export const ChangesetCloseNode: React.FunctionComponent<React.PropsWithChildren<ChangesetCloseNodeProps>> = ({
    node,
}) => (
    <li className={styles.job}>
        <span className={styles.jobSeparator} />
        <>
            <PermissionsSyncJobStatusBadge state={node.state} />
            <PermissionsSyncJobSubject job={node} />
            <PermissionsSyncJobReason job={node} />
            <PermissionsSyncJobNumbers job={node} added={true} />
            <PermissionsSyncJobNumbers job={node} added={false} />
            <div className="text-secondary">
                <b>{node.permissionsFound}</b>
            </div>
        </>
    </li>
)

const PermissionsSyncJobStatusBadge: React.FunctionComponent<{ state: PermissionsSyncJobState }> = ({ state }) => (
    <Badge variant={JOB_STATE_METADATA_MAPPING[state].badgeVariant}>{state}</Badge>
)

const PermissionsSyncJobSubject: React.FunctionComponent<{ job: PermissionsSyncJob }> = ({ job }) => (
    <div>
        <div>
            {job.subject.__typename === 'Repository' ? (
                <>
                    <Icon aria-hidden={true} svgPath={mdiCloudQuestion} /> {job.subject.name}
                    {/*    TODO(sashaostrikov) use code host related icons after GQL API is updated*/}
                </>
            ) : (
                <>
                    <Icon aria-hidden={true} svgPath={mdiAccount} /> {job.subject.username}
                </>
            )}
        </div>
        {JOB_STATE_METADATA_MAPPING[job.state].timeGetter(job) !== '' && (
            <Text className="mb-0 text-muted">
                <small>
                    {JOB_STATE_METADATA_MAPPING[job.state].temporalWording}{' '}
                    <Timestamp date={JOB_STATE_METADATA_MAPPING[job.state].timeGetter(job)} />
                </small>
            </Text>
        )}
    </div>
)

const PermissionsSyncJobReason: React.FunctionComponent<{ job: PermissionsSyncJob }> = ({ job }) => (
    <div>
        <div>{job.reason.group}</div>
        <Text className="mb-0 text-muted">
            <small>
                {job.reason.group === PermissionsSyncJobReasonGroup.MANUAL && job.triggeredByUser?.username
                    ? `by ${job.triggeredByUser.username}`
                    : job.reason.message}
            </small>
            {/*    TODO(sashaostrikov) use pretty-printed message*/}
        </Text>
    </div>
)

// added/removed access for X repositories/users
const PermissionsSyncJobNumbers: React.FunctionComponent<{ job: PermissionsSyncJob; added: boolean }> = ({
    job,
    added,
}) =>
    added ? (
        <Tooltip
            content={`Added access for ${job.permissionsAdded} ${
                job.subject.__typename === 'Repository' ? 'users' : 'repositories'
            }.`}
        >
            <div className="text-success">
                +<b>{job.permissionsAdded}</b>
            </div>
        </Tooltip>
    ) : (
        <Tooltip
            content={`Removed access for ${job.permissionsRemoved} ${
                job.subject.__typename === 'Repository' ? 'users' : 'repositories'
            }.`}
        >
            <div className="text-danger">
                -<b>{job.permissionsRemoved}</b>
            </div>
        </Tooltip>
    )
