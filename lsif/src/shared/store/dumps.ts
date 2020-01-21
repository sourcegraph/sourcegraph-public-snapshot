import * as sharedMetrics from '../database/metrics'
import * as pgModels from '../models/pg'
import { getCommitsNear, getHead } from '../gitserver/gitserver'
import { Brackets, Connection, EntityManager } from 'typeorm'
import { dbFilename, tryDeleteFile } from '../paths'
import { logAndTraceCall, TracingContext } from '../tracing'
import { instrumentQuery, instrumentQueryOrTransaction, withInstrumentedTransaction } from '../database/postgres'
import { TableInserter } from '../database/inserter'
import { visibleDumps, lineageWithDumps, ancestorLineage, bidirectionalLineage } from '../models/queries'

/**
 * The insertion metrics for Postgres.
 */
const insertionMetrics = {
    durationHistogram: sharedMetrics.postgresInsertionDurationHistogram,
    errorsCounter: sharedMetrics.postgresQueryErrorsCounter,
}

/**
 * A wrapper around the database tables that control dumps and commits.
 */
export class DumpManager {
    /**
     * Create a new `DumpManager` backed by the given database connection.
     *
     * @param connection The Postgres connection.
     * @param storageRoot The path where SQLite databases are stored.
     */
    constructor(private connection: Connection, private storageRoot: string) {}

    /**
     * Find the dump for the given repository and commit.
     *
     * @param repositoryId The repository identifier.
     * @param commit The commit.
     * @param file A filename that should be included in the dump.
     */
    public getDump(repositoryId: number, commit: string, file: string): Promise<pgModels.LsifDump | undefined> {
        return instrumentQuery(() =>
            this.connection
                .getRepository(pgModels.LsifDump)
                .createQueryBuilder()
                .select()
                .where({ repositoryId, commit })
                .andWhere(":file LIKE (root || '%')", { file })
                .getOne()
        )
    }

    /**
     * Get a dump by identifier.
     *
     * @param id The dump identifier.
     */
    public getDumpById(id: pgModels.DumpId): Promise<pgModels.LsifDump | undefined> {
        return instrumentQuery(() => this.connection.getRepository(pgModels.LsifDump).findOne({ id }))
    }

    /**
     * Return a map from upload ids to their state.
     *
     * @param ids The upload ids to fetch.
     */
    public async getUploadStates(ids: pgModels.DumpId[]): Promise<Map<pgModels.DumpId, pgModels.LsifUploadState>> {
        if (ids.length === 0) {
            return new Map()
        }

        const result: { id: pgModels.DumpId; state: pgModels.LsifUploadState }[] = await instrumentQuery(() =>
            this.connection
                .getRepository(pgModels.LsifUpload)
                .createQueryBuilder()
                .select(['id', 'state'])
                .where('id IN (:...ids)', { ids })
                .getRawMany()
        )

        return new Map<pgModels.DumpId, pgModels.LsifUploadState>(result.map(u => [u.id, u.state]))
    }

    /**
     * Find the visible dumps. This method is used for testing.
     *
     * @param repositoryId The repository identifier.
     */
    public getVisibleDumps(repositoryId: number): Promise<pgModels.LsifDump[]> {
        return instrumentQuery(() =>
            this.connection
                .getRepository(pgModels.LsifDump)
                .createQueryBuilder()
                .select()
                .where({ repositoryId, visibleAtTip: true })
                .getMany()
        )
    }

    /**
     * Get the oldest dump that is not visible at the tip of its repository.
     *
     * @param entityManager The EntityManager to use as part of a transaction.
     */
    public getOldestPrunableDump(
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<pgModels.LsifDump | undefined> {
        return instrumentQuery(() =>
            entityManager
                .getRepository(pgModels.LsifDump)
                .createQueryBuilder()
                .select()
                .where({ visibleAtTip: false })
                .orderBy('uploaded_at')
                .getOne()
        )
    }

    /**
     * Return the dump 'closest' to the given target commit (a direct descendant or ancestor of
     * the target commit). If no closest commit can be determined, this method returns undefined.
     *
     * @param repositoryId The repository identifier.
     * @param commit The target commit.
     * @param file One of the files in the dump.
     * @param ctx The tracing context.
     * @param frontendUrl The url of the frontend internal API.
     */
    public async findClosestDump(
        repositoryId: number,
        commit: string,
        file: string,
        ctx: TracingContext = {},
        frontendUrl?: string
    ): Promise<pgModels.LsifDump | undefined> {
        // Request updated commit data from gitserver if this commit isn't already
        // tracked. This will pull back ancestors for this commit up to a certain
        // (configurable) depth and insert them into the database. This populates
        // the necessary data for the following query.
        if (frontendUrl) {
            await this.updateCommits(
                repositoryId,
                await this.discoverCommits({ repositoryId, commit, frontendUrl, ctx }),
                ctx
            )
        }

        return logAndTraceCall(ctx, 'Finding closest dump', async () => {
            const query = `
                WITH
                ${bidirectionalLineage()},
                ${lineageWithDumps()}

                SELECT d.dump_id FROM lineage_with_dumps d
                WHERE $3 LIKE (d.root || '%')
                ORDER BY d.n LIMIT 1
            `

            return withInstrumentedTransaction(this.connection, async entityManager => {
                const results: { dump_id: number }[] = await entityManager.query(query, [repositoryId, commit, file])
                if (results.length > 0) {
                    return entityManager.getRepository(pgModels.LsifDump).findOne(results[0].dump_id)
                }

                return undefined
            })
        })
    }

    /**
     * Determine the set of dumps which are 'visible' from the given commit and set the
     * `visible_at_tip` flags. Unset the flag for each invisible dump for this repository.
     * This will traverse all ancestor commits but not descendants, as the given commit
     * is assumed to be the tip of the default branch. For each dump that is filtered out
     * of the set of results, there must be a dump with a smaller depth from the given commit
     * that has a root that overlaps with the filtered dump. The other such dump is
     * necessarily a dump associated with a closer commit for the same root.
     *
     * @param repositoryId The repository identifier.
     * @param commit The head of the default branch.
     * @param ctx The tracing context.
     * @param entityManager The EntityManager to use as part of a transaction.
     */
    public updateDumpsVisibleFromTip(
        repositoryId: number,
        commit: string,
        ctx: TracingContext = {},
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<void> {
        const query = `
            WITH
            ${ancestorLineage()},
            ${visibleDumps()}

            -- Update dump records by:
            --   (1) unsetting the visibility flag of all previously visible dumps, and
            --   (2) setting the visibility flag of all currently visible dumps
            UPDATE lsif_dumps d
            SET visible_at_tip = id IN (SELECT * from visible_ids)
            WHERE d.repository_id = $1 AND (d.id IN (SELECT * from visible_ids) OR d.visible_at_tip)
        `

        return logAndTraceCall(ctx, 'Updating dumps visible from tip', () =>
            instrumentQuery(() => entityManager.query(query, [repositoryId, commit]))
        )
    }

    /**
     * Update the known commits for a repository. The input commits must be a map from commits to
     * a set of parent commits. Commits without a parent should have an empty set of parents, but
     * should still be present in the map.
     *
     * @param repositoryId The repository identifier.
     * @param commits The commit parentage data.
     * @param ctx The tracing context.
     * @param entityManager The EntityManager to use as part of a transaction.
     */
    public updateCommits(
        repositoryId: number,
        commits: Map<string, Set<string>>,
        ctx: TracingContext = {},
        entityManager?: EntityManager
    ): Promise<void> {
        return logAndTraceCall(ctx, 'Updating commits', () =>
            instrumentQueryOrTransaction(this.connection, entityManager, async definiteEntityManager => {
                const commitInserter = new TableInserter(
                    definiteEntityManager,
                    pgModels.Commit,
                    pgModels.Commit.BatchSize,
                    insertionMetrics,
                    true // Do nothing on conflict
                )

                for (const [commit, parentCommits] of commits) {
                    if (parentCommits.size === 0) {
                        await commitInserter.insert({ repositoryId, commit, parentCommit: null })
                    }

                    for (const parentCommit of parentCommits) {
                        await commitInserter.insert({ repositoryId, commit, parentCommit })
                    }
                }

                await commitInserter.flush()
            })
        )
    }

    /**
     * Get a list of commits for the given repository with their parent starting at the
     * given commit and returning at most `MAX_COMMITS_PER_UPDATE` commits. The output
     * is a map from commits to a set of parent commits. The set of parents may be empty.
     * If we already have commit parentage information for this commit, this function
     * will do nothing.
     *
     * @param args Parameter bag.
     */
    public async discoverCommits({
        repositoryId,
        commit,
        frontendUrl,
        ctx = {},
    }: {
        /** The repository identifier. */
        repositoryId: number
        /** The commit from which the gitserver queries should start. */
        commit: string
        /** The url of the frontend internal API. */
        frontendUrl: string
        /** The tracing context. */
        ctx?: TracingContext
    }): Promise<Map<string, Set<string>>> {
        const matchingRepos = await instrumentQuery(() =>
            this.connection.getRepository(pgModels.LsifUpload).count({ where: { repositoryId } })
        )
        if (matchingRepos === 0) {
            return new Map()
        }

        const matchingCommits = await instrumentQuery(() =>
            this.connection.getRepository(pgModels.Commit).count({ where: { repositoryId, commit } })
        )
        if (matchingCommits > 0) {
            return new Map()
        }

        return getCommitsNear(frontendUrl, repositoryId, commit, ctx)
    }

    /**
     * Query gitserver for the head of the default branch for the given repository.
     *
     * @param args Parameter bag.
     */
    public discoverTip({
        repositoryId,
        frontendUrl,
        ctx = {},
    }: {
        /** The repository identifier. */
        repositoryId: number
        /** The url of the frontend internal API. */
        frontendUrl: string
        /** The tracing context. */
        ctx?: TracingContext
    }): Promise<string | undefined> {
        return logAndTraceCall(ctx, 'Getting repository metadata', () => getHead(frontendUrl, repositoryId, ctx))
    }

    /**
     * Delete existing dumps from the same repo@commit that overlap with the current root
     * (where the existing root is a prefix of the current root, or vice versa).
     *
     * @param repositoryId The repository identifier.
     * @param commit The commit.
     * @param root The root of all files that are in this dump.
     * @param ctx The tracing context.
     * @param entityManager The EntityManager to use as part of a transaction.
     */
    public async deleteOverlappingDumps(
        repositoryId: number,
        commit: string,
        root: string,
        ctx: TracingContext = {},
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<void> {
        return logAndTraceCall(ctx, 'Clearing overlapping dumps', () =>
            instrumentQuery(async () => {
                await entityManager
                    .getRepository(pgModels.LsifUpload)
                    .createQueryBuilder()
                    .delete()
                    .where({ repositoryId, commit, state: 'completed' })
                    .andWhere(
                        new Brackets(qb =>
                            qb.where(":root LIKE (root || '%')", { root }).orWhere("root LIKE (:root || '%')", { root })
                        )
                    )
                    .execute()
            })
        )
    }

    /**
     * Delete a dump. This removes data from the dumps, packages, and references table, and
     * deletes the SQLite file from the storage root.
     *
     * @param dump The dump to delete.
     * @param entityManager The EntityManager to use as part of a transaction.
     */
    public async deleteDump(
        dump: pgModels.LsifDump,
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<void> {
        // Delete the SQLite file on disk (ignore errors if the file doesn't exist)
        const path = dbFilename(this.storageRoot, dump.id)
        await tryDeleteFile(path)

        // Delete the dump record. Do this AFTER the file is deleted because the retention
        // policy scans the database for deletion candidates, and we don't want to get into
        // the situation where the row is gone and the file is there. In this case, we don't
        // have any process to tell us that the file is okay to delete and will be orphaned
        // on disk forever.

        await entityManager.getRepository(pgModels.LsifDump).delete(dump.id)
    }
}
