import * as GQL from '../graphql/schema'

/**
 * Possible states for an LSIF job.
 */
export type LsifJobStatus = 'active' | 'queued' | 'completed' | 'failed'

/**
 * Counts of LSIF jobs in each state.
 */
export interface LsifJobStats {
    queued: number
    active: number
    completed: number
    failed: number
}

/**
 * Metadata about an LSIF job.
 */
export interface ILsifJob {
    jobId: string
    name: string
    args: object
    status: LsifJobStatus
    progress: string
    failedReason: string
    stacktrace: string[] | null
    timestamp: string
    finishedOn: string | null
    processedOn: string | null
}

/**
 * A wrapper to make LSIF job list results look like a GraphQL response.
 */
export interface ILsifJobConnection {
    __typename: 'LsifJobConnenction'
    nodes: ILsifJob[]
    totalCount: number | null
    pageInfo: GQL.IPageInfo
}

/**
 * Construct a meaningful description from the job name and args.
 *
 * @param job The job instance.
 */
export function lsifJobDescription(job: ILsifJob): string {
    if (job.name === 'convert') {
        const { repository, commit, root } = job.args as {
            repository: string
            commit: string
            root: string
        }

        return `Convert ${repository}@${commit.substring(0, 8)}${root === '' ? '' : `:${root}`}`
    }

    if (job.name === 'clean-old-jobs') {
        return 'Scheduled work queue prune job'
    }

    if (job.name === 'update-tips') {
        return 'Scheduled update-tips job'
    }

    return `${job.name} Job #${job.jobId}`
}
