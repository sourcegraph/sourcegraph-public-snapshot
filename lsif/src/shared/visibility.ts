import { EntityManager } from 'typeorm'
import { DumpManager } from './store/dumps'
import { TracingContext } from './tracing'

/**
 * Update the commits for this repo, and update the visible_at_tip flag on the dumps
 * of this repository. This will query for commits starting from both the current tip
 * of the repo and from given commit.
 *
 * @param args Parameter bag.
 */
export async function updateCommitsAndDumpsVisibleFromTip({
    entityManager,
    dumpManager,
    frontendUrl,
    repositoryId,
    commit,
    ctx = {},
}: {
    /** The EntityManager to use as part of a transaction. */
    entityManager: EntityManager
    /** The dumps manager instance. */
    dumpManager: DumpManager
    /** The url of the frontend internal API. */
    frontendUrl: string
    /** The repository id. */
    repositoryId: number
    /**
     * An optional commit. This should be supplied if an upload was just
     * processed. If no commit is supplied, then the commits will be queried
     * only from the tip commit of the default branch.
     */
    commit?: string
    /** The tracing context. */
    ctx?: TracingContext
}): Promise<void> {
    const tipCommit = await dumpManager.discoverTip({
        repositoryId,
        frontendUrl,
        ctx,
    })
    if (tipCommit === undefined) {
        throw new Error('No tip commit available for repository')
    }

    const commits = commit
        ? await dumpManager.discoverCommits({
              repositoryId,
              commit,
              frontendUrl,
              ctx,
          })
        : new Map()

    if (tipCommit !== commit) {
        // If the tip is ahead of this commit, we also want to discover all of
        // the commits between this commit and the tip so that we can accurately
        // determine what is visible from the tip. If we do not do this before the
        // updateDumpsVisibleFromTip call below, no dumps will be reachable from
        // the tip and all dumps will be invisible.

        const tipCommits = await dumpManager.discoverCommits({
            repositoryId,
            commit: tipCommit,
            frontendUrl,
            ctx,
        })

        for (const [k, v] of tipCommits.entries()) {
            commits.set(
                k,
                new Set<string>([...(commits.get(k) || []), ...v])
            )
        }
    }

    await dumpManager.updateCommits(repositoryId, commits, ctx, entityManager)
    await dumpManager.updateDumpsVisibleFromTip(repositoryId, tipCommit, ctx, entityManager)
}
