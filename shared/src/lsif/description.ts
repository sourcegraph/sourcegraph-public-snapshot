import * as GQL from '../graphql/schema'

//
// TODO - move this entire file?
//

/**
 * Construct a meaningful description from the job name and args.
 *
 * @param job The job instance.
 */
export function lsifJobDescription(job: GQL.ILsifJob): string {
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

    return `${job.name} Job #${job.id}`
}
