import got from 'got'
import { XrepoDatabase } from './xrepo'
import * as crypto from 'crypto'
import { TracingContext, logAndTraceCall } from './tracing'
import { chunk } from 'lodash'
import { MAX_CONCURRENT_GITSERVER_REQUESTS, MAX_COMMITS_PER_UPDATE } from './constants'

/**
 * Update the commits tables in the cross-repository database with the current
 * output of gitserver for the given repository around the given commit. If we
 * already have commit parentage information for this commit, this function
 * will do nothing.
 *
 * @param args Parameter bag.
 */
export async function discoverAndUpdateCommit({
    xrepoDatabase,
    repository,
    commit,
    gitserverUrls,
    ctx,
}: {
    /** The cross-repo database. */
    xrepoDatabase: XrepoDatabase
    /** The repository name. */
    repository: string
    /** The commit from which the gitserver queries should start. */
    commit: string
    /** The set of ordered gitserver urls. */
    gitserverUrls: string[]
    /** The tracing context. */
    ctx: TracingContext
}): Promise<void> {
    // No need to update if we already know about this commit
    if (await xrepoDatabase.isCommitTracked(repository, commit)) {
        return
    }

    // No need to pull commits for repos we don't have data for
    if (!(await xrepoDatabase.isRepositoryTracked(repository))) {
        return
    }

    const gitserverUrl = addrFor(repository, gitserverUrls)
    const commits = await logAndTraceCall(ctx, 'querying commits', () =>
        getCommitsNear(gitserverUrl, repository, commit)
    )
    await logAndTraceCall(ctx, 'updating commits', () => xrepoDatabase.updateCommits(repository, commits))
}

/**
 * Update the known tip of the default branch for every repository for which
 * we have LSIF data. This queries gitserver for the last known tip. From that,
 * we determine the closest commit with LSIF data and mark those as commits for
 * which we can return results in a global find-references query.
 *
 * @param args Parameter bag.
 */
export async function discoverAndUpdateTips({
    xrepoDatabase,
    gitserverUrls,
    ctx,
    batchSize = MAX_CONCURRENT_GITSERVER_REQUESTS,
}: {
    /** The cross-repo database. */
    xrepoDatabase: XrepoDatabase
    /** The set of ordered gitserver urls. */
    gitserverUrls: string[]
    /** The tracing context. */
    ctx: TracingContext
    /** The maximum number of requests to make at once. Set during testing.*/
    batchSize?: number
}): Promise<void> {
    for (const [repository, commit] of (await discoverTips({
        xrepoDatabase,
        gitserverUrls,
        ctx,
        batchSize,
    })).entries()) {
        await xrepoDatabase.updateDumpsVisibleFromTip(repository, commit)
    }
}

/**
 * Query gitserver for the head of the default branch for every repository that has
 * LSIF data.
 *
 * @param args Parameter bag.
 */
export async function discoverTips({
    xrepoDatabase,
    gitserverUrls,
    ctx,
    batchSize = MAX_CONCURRENT_GITSERVER_REQUESTS,
}: {
    /** The cross-repo database. */
    xrepoDatabase: XrepoDatabase
    /** The set of ordered gitserver urls. */
    gitserverUrls: string[]
    /** The tracing context. */
    ctx: TracingContext
    /** The maximum number of requests to make at once. Set during testing.*/
    batchSize?: number
}): Promise<Map<string, string>> {
    // Construct the calls we need to make to gitserver for each repository that
    // we know about. We're going to construct these as factories so they do not
    // start immediately and we can apply them in batches to not overload us or
    // gitserver with too many in-flight requests.

    const factories: (() => Promise<{ repository: string; commit: string }>)[] = []
    for (const repository of await xrepoDatabase.getTrackedRepositories()) {
        factories.push(async () => {
            const lines = await gitserverExecLines(addrFor(repository, gitserverUrls), repository, [
                'git',
                'rev-parse',
                'HEAD',
            ])

            return { repository, commit: lines ? lines[0] : '' }
        })
    }

    const tips = new Map<string, string>()
    for (const batch of chunk(factories, batchSize)) {
        // Perform this batch of calls to the appropriate gitserver instance
        const responses = await logAndTraceCall(ctx, 'getting repository metadata', () =>
            Promise.all(batch.map(factory => factory()))
        )

        // Combine the results
        for (const { repository, commit } of responses) {
            tips.set(repository, commit)
        }
    }

    return tips
}

/**
 * Determine the gitserver that holds data for the given repository. This matches the
 * same selection process as done by the frontend (see pkg/gitserver/client.go). The
 * set of gitserverUrls must be the same (and in the same order) as the frontend for
 * this to return consistent results.
 *
 * @param repository The repository name.
 * @param gitserverUrls The set of ordered gitserver urls.
 */
function addrFor(repository: string, gitserverUrls: string[]): string {
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
 * If the repository or commit is unknown by gitserver, then the the results will be
 * empty but no error will be thrown. Any other error type will b thrown without
 * modification.
 *
 * @param gitserverUrl The url of the gitserver for this repository.
 * @param repository The repository name.
 * @param commit The commit from which the gitserver queries should start.
 */
export async function getCommitsNear(
    gitserverUrl: string,
    repository: string,
    commit: string
): Promise<[string, string][]> {
    const args = ['log', '--pretty=%H %P', commit, `-${MAX_COMMITS_PER_UPDATE}`]

    try {
        return flattenCommitParents(await gitserverExecLines(gitserverUrl, repository, args))
    } catch (error) {
        if (error.statusCode === 404) {
            // repository unknown
            return []
        }

        throw error
    }
}

/**
 * Convert git log output into a map of (`child`, `parent`) commits. Each line should
 * have the form `commit parent1 parent2...`. Commits with no parents appear on a line
 * of their own.
 *
 * @param lines The output lines of `git log`.
 */
export function flattenCommitParents(lines: string[]): [string, string][] {
    return lines.flatMap(line => {
        const commits = line.split(' ')
        if (commits.length === 1) {
            return [[line, '']]
        }

        const child = commits.shift()!
        return commits.map<[string, string]>(commit => [child, commit])
    })
}

/**
 * Execute a git command via gitserver and return its output split into non-empty lines.
 *
 * @param gitserverUrl The url of the gitserver for this repository.
 * @param repository The repository name.
 * @param args The command to run in the repository's git directory.
 */
async function gitserverExecLines(gitserverUrl: string, repository: string, args: string[]): Promise<string[]> {
    return (await gitserverExec(gitserverUrl, repository, args)).split('\n').filter(line => Boolean(line))
}

/**
 * Execute a git command via gitserver and return its raw output.
 *
 * @param gitserverUrl The url of the gitserver for this repository.
 * @param repository The repository name.
 * @param args The command to run in the repository's git directory.
 */
async function gitserverExec(gitserverUrl: string, repository: string, args: string[]): Promise<string> {
    const body = JSON.stringify({ repo: repository, args })
    const resp = await got(new URL(`http://${gitserverUrl}/exec`).href, { body })
    return resp.body
}
