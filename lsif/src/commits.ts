import * as crypto from 'crypto'
import got from 'got'
import { XrepoDatabase } from './xrepo'
import { MonitoringContext, monitor } from './monitoring'

/**
 * THe number of commits to ask gitserver for when updating commit data for
 * a particular repository.
 */
const MAX_COMMITS_PER_UPDATE = 5000

/**
 * Update the commits tables in the cross-repository database with the current
 * output of gitserver for the given repository around the given commit. If we
 * already have commit parentage information for this commit, this function
 * will do nothing.
 *
 * @param gitserverUrls The set of ordered gitserver urls.
 * @param xrepoDatabase The cross-repo database.
 * @param repository The repository name.
 * @param commit The commit from which the gitserver queries should start.
 * @param ctx The monitoring context.
 */
export async function updateCommits(
    gitserverUrls: string[],
    xrepoDatabase: XrepoDatabase,
    repository: string,
    commit: string,
    ctx: MonitoringContext
): Promise<void> {
    if (await xrepoDatabase.isCommitTracked(repository, commit)) {
        return
    }

    const gitserverUrl = addrFor(gitserverUrls, repository)
    const commits = await monitor(ctx, 'querying commits', () => getCommitsNear(gitserverUrl, repository, commit))
    await monitor(ctx, 'updating commits', () => xrepoDatabase.updateCommits(repository, commits))
}

/**
 * Determine the gitserver that holds data for the given repository. This matches the
 * same selection process as done by the frontend (see pkg/gitserver/client.go). The
 * set of gitserverUrls must be the same (and in the same order) as the frontend for
 * this to return consistent results.
 *
 * @param gitserverUrls The set of ordered gitserver urls.
 * @param repository The repository name.
 */
function addrFor(gitserverUrls: string[], repository: string): string {
    if (gitserverUrls.length === 0) {
        throw new Error('No gitserverUrls provided.')
    }

    return gitserverUrls[hashmod(repository, gitserverUrls.length)]
}

/**
 * Determine the md5 hash of the given value, then determine the mod with respect to
 * the given max value. The md5 hash is treated as a uint64 (only the upper 16 hex
 * digits are considered).
 *
 * @param value The value to hash.
 * @param max The maximum bound.
 */
export function hashmod(value: string, max: number): number {
    const sum = crypto
        .createHash('md5')
        .update(value)
        .digest('hex')

    return mod(sum.substring(0, 16), max)
}

/**
 * Determine the mod of the hex string against the given max.
 *
 * @param sum The hex-string representation of the number to mod.
 * @param max The maximum bound.
 */
export function mod(sum: string, max: number): number {
    let index = 0
    for (const ch of sum) {
        index = (index * 16 + parseInt(ch, 16)) % max
    }

    return index
}

/**
 * Get a list of commits for the given repository with their parent starting at the
 * given commit and returning at most `MAX_COMMITS_PER_UPDATE` commits. Each value in
 * the return list of the form `[A, B]` indicates that `B` is a parent of `A`. The
 * second element is an empty string if the first element has no parent. Commits may
 * appear multiple times, but each pair is unique.
 *
 * @param gitserverUrl The url of the gitserver for this repository.
 * @param repository The repository name.
 * @param commit The commit from which the gitserver queries should start.
 */
async function getCommitsNear(gitserverUrl: string, repository: string, commit: string): Promise<[string, string][]> {
    const args = ['git', 'log', '--pretty=%H %P', commit, `-${MAX_COMMITS_PER_UPDATE}`]
    const lines = await gitserverExecLines(gitserverUrl, repository, args)
    return lines.map(line => line.split(' ', 2) as [string, string])
}

/**
 * Execute a git command via gitserver and return its output split into non-empty lines.
 *
 * @param gitserverUrl The url of the gitserver for this repository.
 * @param repository The repository name.
 * @param args The command to run in the repository's git directory.
 */
async function gitserverExecLines(gitserverUrl: string, repository: string, args: string[]): Promise<string[]> {
    return (await gitserverExec(gitserverUrl, repository, args)).split('\n').filter(line => !!line)
}

/**
 * Execute a git command via gitserver and return its raw output.
 *
 * @param gitserverUrl The url of the gitserver for this repository.
 * @param repository The repository name.
 * @param args The command to run in the repository's git directory.
 */
async function gitserverExec(gitserverUrl: string, repository: string, args: string[]): Promise<string> {
    return (await got(new URL(`http://${gitserverUrl}/exec`).href, {
        body: JSON.stringify({ repo: repository, args }),
    })).body
}
