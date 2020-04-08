import got from 'got'
import { MAX_COMMITS_PER_UPDATE } from '../constants'
import { TracingContext, logAndTraceCall } from '../tracing'
import { instrument } from '../metrics'
import * as metrics from './metrics'

/**
 * Get the children of a set of directories at a particular commit. This function
 * returns a list of sets `L` where `L[i]` is the set of children for `dirnames[i]`.
 *
 * Except for the root directory, denoted by the empty string, all supplied
 * directories should be disjoint (one should not contain another).
 *
 * @param args Parameter bag.
 */
export async function getDirectoryChildren({
    frontendUrl,
    repositoryId,
    commit,
    dirnames,
    ctx = {},
}: {
    /** The url of the frontend internal API. */
    frontendUrl: string
    /** The repository identifier. */
    repositoryId: number
    /** The commit from which the gitserver queries should start. */
    commit: string
    /** A list of repo-root-relative directories. */
    dirnames: string[]
    /** The tracing context. */
    ctx?: TracingContext
}): Promise<Map<string, Set<string>>> {
    const args = ['ls-tree', '--name-only', commit, '--']

    for (const dirname of dirnames) {
        if (dirname === '') {
            args.push('.')
        } else {
            args.push(dirname.endsWith('/') ? dirname : dirname + '/')
        }
    }

    // Retrieve a flat list of children of the dirnames constructed above. This returns a flat
    // list sorted alphabetically, which we then need to partition by parent directory.
    const uncategorizedChildren = await gitserverExecLines(frontendUrl, repositoryId, args, ctx)

    const childMap = new Map()
    for (const dirname of dirnames) {
        childMap.set(
            dirname,
            dirname === ''
                ? new Set(uncategorizedChildren.filter(line => !line.includes('/')))
                : new Set(uncategorizedChildren.filter(line => line.startsWith(dirname)))
        )
    }

    return childMap
}

/**
 * Get a list of commits for the given repository with their parent starting at the
 * given commit and returning at most `MAX_COMMITS_PER_UPDATE` commits. The output
 * is a map from commits to a set of parent commits. The set of parents may be empty.
 *
 * If the repository or commit is unknown by gitserver, then the the results will be
 * empty but no error will be thrown. Any other error type will be thrown without
 * modification.
 *
 * @param frontendUrl The url of the frontend internal API.
 * @param repositoryId The repository identifier.
 * @param commit The commit from which the gitserver queries should start.
 * @param ctx The tracing context.
 */
export async function getCommitsNear(
    frontendUrl: string,
    repositoryId: number,
    commit: string,
    ctx: TracingContext = {}
): Promise<Map<string, Set<string>>> {
    const args = ['log', '--pretty=%H %P', commit, `-${MAX_COMMITS_PER_UPDATE}`]

    try {
        return flattenCommitParents(await gitserverExecLines(frontendUrl, repositoryId, args, ctx))
    } catch (error) {
        if (error.response && error.response.statusCode === 404) {
            // Unknown repository
            return new Map()
        }

        throw error
    }
}

/**
 * Convert git log output into a parentage map. Each line of the input should have the
 * form `commit p1 p2 p3...`, where commits without a parent appear on a line of their
 * own. The output is a map from commits a set of parent commits. The set of parents may
 * be empty.
 *
 * @param lines The output lines of `git log`.
 */
export function flattenCommitParents(lines: string[]): Map<string, Set<string>> {
    const commits = new Map()
    for (const line of lines) {
        const trimmed = line.trim()
        if (trimmed === '') {
            continue
        }

        const [child, ...parentCommits] = trimmed.split(' ')
        commits.set(child, new Set<string>(parentCommits))
    }

    return commits
}

/**
 * Get the current tip of the default branch of the given repository.
 *
 * @param frontendUrl The url of the frontend internal API.
 * @param repositoryId The repository identifier.
 * @param ctx The tracing context.
 */
export async function getHead(
    frontendUrl: string,
    repositoryId: number,
    ctx: TracingContext = {}
): Promise<string | undefined> {
    const lines = await gitserverExecLines(frontendUrl, repositoryId, ['rev-parse', 'HEAD'], ctx)
    if (lines.length === 0) {
        return undefined
    }

    return lines[0]
}

/**
 * Execute a git command via gitserver and return its output split into non-empty lines.
 *
 * @param frontendUrl The url of the frontend internal API.
 * @param repositoryId The repository identifier.
 * @param args The command to run in the repository's git directory.
 * @param ctx The tracing context.
 */
export async function gitserverExecLines(
    frontendUrl: string,
    repositoryId: number,
    args: string[],
    ctx: TracingContext = {}
): Promise<string[]> {
    return (await gitserverExec(frontendUrl, repositoryId, args, ctx)).split('\n').filter(line => Boolean(line))
}

/**
 * Execute a git command via gitserver and return its raw output.
 *
 * @param frontendUrl The url of the frontend internal API.
 * @param repositoryId The repository identifier.
 * @param args The command to run in the repository's git directory.
 * @param ctx The tracing context.
 */
function gitserverExec(
    frontendUrl: string,
    repositoryId: number,
    args: string[],
    ctx: TracingContext = {}
): Promise<string> {
    if (args[0] === 'git') {
        // Prevent this from happening again:
        // https://github.com/sourcegraph/sourcegraph/pull/5941
        // https://github.com/sourcegraph/sourcegraph/pull/6548
        throw new Error('Gitserver commands should not be prefixed with `git`')
    }

    return logAndTraceCall(ctx, 'Executing git command', () =>
        instrument(metrics.gitserverDurationHistogram, metrics.gitserverErrorsCounter, async () => {
            // Perform request - this may fail with a 404 or 500
            const resp = await got.post(new URL(`http://${frontendUrl}/.internal/git/${repositoryId}/exec`).href, {
                body: JSON.stringify({ args }),
            })

            // Read trailers on a 200-level response
            const status = resp.trailers['x-exec-exit-status']
            const stderr = resp.trailers['x-exec-stderr']

            // Determine if underlying git command failed and throw an error
            // in that case. Status will be undefined in some of our tests and
            // will be the process exit code (given as a string) otherwise.
            if (status !== undefined && status !== '0') {
                throw new Error(`Failed to run git command ${['git', ...args].join(' ')}: ${String(stderr)}`)
            }

            return resp.body
        })
    )
}
