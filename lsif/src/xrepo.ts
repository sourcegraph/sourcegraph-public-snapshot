import {
    bloomFilterEventsCounter,
    xrepoInsertionDurationHistogram,
    xrepoQueryDurationHistogram,
    xrepoQueryErrorsCounter,
} from './xrepo.metrics'
import * as crc32 from 'crc-32'
import { instrument } from './metrics'
import { Connection, EntityManager, Brackets } from 'typeorm'
import { createFilter, testFilter } from './encoding'
import { PackageModel, ReferenceModel, Commit, LsifDump, DumpId } from './xrepo.models'
import { TableInserter } from './inserter'
import { addrFor, getCommitsNear, gitserverExecLines } from './commits'
import { TracingContext, logAndTraceCall } from './tracing'
import { dbFilename, tryDeleteFile } from './util'
import { MAX_TRAVERSAL_LIMIT, ADVISORY_LOCK_ID_SALT, MAX_CONCURRENT_GITSERVER_REQUESTS } from './constants'
import { chunk } from 'lodash'

/**
 * The insertion metrics for the cross-repo database.
 */
const insertionMetrics = {
    durationHistogram: xrepoInsertionDurationHistogram,
    errorsCounter: xrepoQueryErrorsCounter,
}

/**
 * Represents a package provided by a project or a package that is a dependency
 * of a project, depending on its use.
 */
export interface Package {
    /**
     * The scheme of the package (e.g. npm, pip).
     */
    scheme: string

    /**
     * The name of the package.
     */
    name: string

    /**
     * The version of the package.
     */
    version: string | null
}

/**
 * Represents a use of a set of symbols from a particular dependent package of
 * a project.
 */
export interface SymbolReferences {
    /**
     * The package from which the symbols are imported.
     */
    package: Package

    /**
     * The unique identifiers of the symbols imported from the package.
     */
    identifiers: string[]
}

/**
 * A wrapper around a SQLite database that stitches together the references
 * between projects at a specific commit. This is used for cross-repository
 * jump to definition and find references features.
 */
export class XrepoDatabase {
    /**
     * Create a new `XrepoDatabase` backed by the given database connection.
     *
     * @param storageRoot The path where SQLite databases are stored.
     * @param connection The Postgres connection.
     */
    constructor(private storageRoot: string, private connection: Connection) {}

    /**
     * Get the dumps for a repository.
     *
     * @param repository The repository.
     * @param query A search query.
     * @param limit The maximum number of dumps to return.
     * @param offset The number of dumps to skip.
     */
    public async getDumps(
        repository: string,
        query: string,
        limit: number,
        offset: number
    ): Promise<{ dumps: LsifDump[]; totalCount: number }> {
        const [dumps, totalCount] = await this.withConnection(connection => {
            let queryBuilder = connection
                .getRepository(LsifDump)
                .createQueryBuilder()
                .where({ repository })
                .orderBy('uploaded_at')
                .limit(limit)
                .offset(offset)

            if (query) {
                queryBuilder = queryBuilder.andWhere(
                    new Brackets(qb =>
                        qb
                            .where("commit LIKE '%' || :query || '%'", { query })
                            .orWhere("root LIKE '%' || :query || '%'", { query })
                    )
                )
            }

            return queryBuilder.getManyAndCount()
        })

        return { dumps, totalCount }
    }

    /**
     * Get a dump by identifier.
     *
     * @param id The dump identifier.
     */
    public getDumpById(id: DumpId): Promise<LsifDump | undefined> {
        return this.withConnection(connection => connection.getRepository(LsifDump).findOne({ id }))
    }

    /**
     * Return the list of all repositories that have LSIF data.
     */
    public async getTrackedRepositories(): Promise<string[]> {
        const payload = await this.withConnection(connection =>
            connection
                .getRepository(LsifDump)
                .createQueryBuilder()
                .select('DISTINCT repository')
                .getRawMany()
        )

        return (payload as { repository: string }[]).map(e => e.repository)
    }

    /**
     * Determine if we have any LSIf data for this repository.
     *
     * @param repository The repository name.
     */
    public isRepositoryTracked(repository: string): Promise<boolean> {
        return this.withConnection(
            async connection =>
                (await connection.getRepository(LsifDump).findOne({ where: { repository } })) !== undefined
        )
    }

    /**
     * Determine if we have commit parent data for this commit.
     *
     * @param repository The repository name.
     * @param commit The target commit.
     */
    public isCommitTracked(repository: string, commit: string): Promise<boolean> {
        return this.withConnection(
            async connection =>
                (await connection.getRepository(Commit).findOne({ where: { repository, commit } })) !== undefined
        )
    }

    /**
     * Determine the set of dumps which are 'visible' from the given commit and set the
     * `visible_at_tip` flags. Unset the flag for each invisible dump for this repository.
     * This will traverse all ancestor commits but not descendants, as the given commit
     * is assumed to be the tip of the default branch. For each dump that is filtered out
     * of the result set, there must be a dump with a smaller depth from the given commit
     * that has a root that overlaps with the filtered dump. The other such dump is
     * necessarily a dump associated with a closer commit for the same root.
     *
     * @param repository The repository name.
     * @param commit The head of the default branch.
     */
    public async updateDumpsVisibleFromTip(repository: string, commit: string): Promise<void> {
        const query = `
            -- Get all ancestors of the tip
            WITH RECURSIVE ancestors(id, repository, "commit", parent) AS (
                SELECT c.* FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
                UNION
                SELECT c.* FROM ancestors a JOIN lsif_commits c ON a.repository = c.repository AND a.parent = c."commit"
            ),
            -- Limit the visibility to the maximum traversal depth and approximate
            -- each commit's depth by its row number.
            limited_ancestors AS (
                SELECT a.*, row_number() OVER() as n from ancestors a LIMIT $3
            ),
            -- Correlate commits to dumps and filter out commits without LSIF data
            ancestors_with_dumps AS (
                SELECT a.*, d.root, d.id as dump_id FROM limited_ancestors a
                JOIN lsif_dumps d ON d.repository = a.repository AND d."commit" = a."commit"
            ),
            visible_ids AS (
                -- Remove dumps where there exists another visible dump of smaller depth with an overlapping root.
                -- Such dumps would not be returned with a closest commit query so we don't want to return results
                -- for them in global find-reference queries either.
                SELECT DISTINCT t1.dump_id as id FROM ancestors_with_dumps t1 WHERE NOT EXISTS (
                    SELECT 1 FROM ancestors_with_dumps t2
                    WHERE t2.n < t1.n AND (
                        t2.root LIKE (t1.root || '%') OR
                        t1.root LIKE (t2.root || '%')
                    )
                )
            )

            -- Update dump records by:
            --   (1) unsetting the visibility flag of all previously visible dumps, and
            --   (2) setting the visibility flag of all currently visible dumps

            UPDATE lsif_dumps d
            SET visible_at_tip = id IN (SELECT * from visible_ids)
            WHERE d.repository = $1 AND (d.id IN (SELECT * from visible_ids) OR d.visible_at_tip)
        `

        await this.withConnection(connection => connection.query(query, [repository, commit, MAX_TRAVERSAL_LIMIT]))
    }

    /**
     * Return the commit that has LSIF data 'closest' to the given target commit (a direct descendant
     * or ancestor of the target commit). If no closest commit can be determined, this method returns
     * undefined.s
     *
     * @param repository The repository name.
     * @param commit The target commit.
     * @param file One of the files in the dump.
     * @param ctx The tracing context.
     * @param gitserverUrls The set of ordered gitserver urls.
     */
    public async findClosestDump(
        repository: string,
        commit: string,
        file: string,
        ctx: TracingContext = {},
        gitserverUrls?: string[]
    ): Promise<LsifDump | undefined> {
        // Request updated commit data from gitserver if this commit isn't
        // already tracked. This will pull back ancestors for this commit
        // up to a certain (configurable) depth and insert them into the
        // cross-repository database. This populates the necessary data for
        // the following query.
        if (gitserverUrls) {
            await this.discoverAndUpdateCommit({ repository, commit, gitserverUrls, ctx })
        }

        return this.withConnection(async connection => {
            const results = (await connection.query('select * from closest_dump($1, $2, $3, $4)', [
                repository,
                commit,
                file,
                MAX_TRAVERSAL_LIMIT,
            ])) as LsifDump[]

            if (results.length === 0) {
                return undefined
            }

            return results[0]
        })
    }

    /**
     * Update the known commits for a repository. The given commits are pairs of revhashes, where
     * [`a`, `b`] indicates that one parent of `b` is `a`. If a commit has no parents, then there
     * should be a pair of the form [`a`, ``] (where the parent is an empty string).
     *
     * @param repository The repository name.
     * @param commits The commit parentage data.
     */
    public updateCommits(repository: string, commits: [string, string][]): Promise<void> {
        return this.withTransactionalEntityManager(async entityManager => {
            const commitInserter = new TableInserter(
                entityManager,
                Commit,
                Commit.BatchSize,
                insertionMetrics,
                true // Do nothing on conflict
            )

            for (const [commit, parentCommit] of commits) {
                await commitInserter.insert({ repository, commit, parentCommit })
            }

            await commitInserter.flush()
        })
    }

    /**
     * Find the package that defines the given `scheme`, `name`, and `version`.
     *
     * @param scheme The package manager scheme (e.g. npm, pip).
     * @param name The package name.
     * @param version The package version.
     */
    public getPackage(scheme: string, name: string, version: string | null): Promise<PackageModel | undefined> {
        return this.withConnection(connection =>
            connection.getRepository(PackageModel).findOne({
                where: {
                    scheme,
                    name,
                    version,
                },
            })
        )
    }

    /**
     * Correlate a `repository` and `commit` with a set of unique packages it defines and
     * with  the the names referenced from a particular dependent package.
     *
     * @param repository The repository being updated.
     * @param commit The commit being updated.
     * @param root The root of all files in this dump.
     * @param packages The list of packages that this repository defines (scheme, name, and version).
     * @param references The list of packages that this repository depends on (scheme, name, and version) and the symbols that the package references.
     */
    public addPackagesAndReferences(
        repository: string,
        commit: string,
        root: string,
        packages: Package[],
        references: SymbolReferences[]
    ): Promise<LsifDump> {
        return this.withTransactionalEntityManager(async entityManager => {
            const dump = await this.insertDump(repository, commit, root, entityManager)

            const packageInserter = new TableInserter<PackageModel, new () => PackageModel>(
                entityManager,
                PackageModel,
                PackageModel.BatchSize,
                insertionMetrics,
                true // Do nothing on conflict
            )

            const referenceInserter = new TableInserter<ReferenceModel, new () => ReferenceModel>(
                entityManager,
                ReferenceModel,
                ReferenceModel.BatchSize,
                insertionMetrics
            )

            for (const pkg of packages) {
                await packageInserter.insert({ dump_id: dump.id, ...pkg })
            }

            for (const reference of references) {
                await referenceInserter.insert({
                    dump_id: dump.id,
                    filter: await createFilter(reference.identifiers),
                    ...reference.package,
                })
            }

            await packageInserter.flush()
            await referenceInserter.flush()

            return dump
        })
    }

    /**
     * Inserts the given repository and commit into the `lsif_dumps` table.
     *
     * @param repository The repository.
     * @param commit The commit.
     * @param root The root of all files that are in this dump.
     * @param entityManager The EntityManager for the connection to the xrepo database.
     */
    public async insertDump(
        repository: string,
        commit: string,
        root: string,
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<LsifDump> {
        // Get existing dumps from the same repo@commit that overlap with the current
        // root (where the existing root is a prefix of the current root, or vice versa).

        const dumps = await entityManager
            .getRepository(LsifDump)
            .createQueryBuilder()
            .select()
            .where({ repository, commit })
            .andWhere(
                new Brackets(qb =>
                    qb.where(":root LIKE (root || '%')", { root }).orWhere("root LIKE (:root || '%')", { root })
                )
            )
            .getMany()

        // Delete conflicting dumps
        for (const dump of dumps) {
            await this.deleteDump(dump, entityManager)
        }

        const dump = new LsifDump()
        dump.repository = repository
        dump.commit = commit
        dump.root = root
        await entityManager.save(dump)
        return dump
    }

    /**
     * Find the dump for the given repository and commit.
     *
     * @param repository The repository.
     * @param commit The commit.
     * @param file A filename that should be included in the dump.
     */
    public getDump(repository: string, commit: string, file: string): Promise<LsifDump | undefined> {
        return this.withConnection(connection =>
            connection
                .getRepository(LsifDump)
                .createQueryBuilder()
                .select()
                .where({ repository, commit })
                .andWhere(":file LIKE (root || '%')", { file })
                .getOne()
        )
    }

    /**
     * Get the oldest dump that is not visible at the tip of its repository.
     */
    public getOldestPrunableDump(): Promise<LsifDump | undefined> {
        return this.withConnection(connection =>
            connection
                .getRepository(LsifDump)
                .createQueryBuilder()
                .select()
                .where({ visibleAtTip: false })
                .orderBy('uploaded_at')
                .getOne()
        )
    }

    /**
     * Delete a dump. This removes data from the dumps, packages, and references table, and
     * deletes the SQLite file from the storage root.
     *
     * @param dump The dump to delete.
     * @param entityManager The EntityManager for the connection to the xrepo database.
     */
    public async deleteDump(
        dump: LsifDump,
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<void> {
        // Delete the SQLite file on disk (ignore errors if the file doesn't exist)
        const path = dbFilename(this.storageRoot, dump.id, dump.repository, dump.commit)
        await tryDeleteFile(path)

        // Delete the dump record. Do this AFTER the file is deleted because the retention
        // policy scans the database for deletion candidates, and we don't want to get into
        // the situation where the row is gone and the file is there. In this case, we don't
        // have any process to tell us that the file is okay to delete and will be orphaned
        // on disk forever.

        await entityManager.getRepository(LsifDump).delete(dump.id)
    }

    /**
     * Find the visible dumps. This method is used for testing.
     *
     * @param repository The repository.
     */
    public getVisibleDumps(repository: string): Promise<LsifDump[]> {
        return this.withConnection(connection =>
            connection
                .getRepository(LsifDump)
                .createQueryBuilder()
                .select()
                .where({ repository, visibleAtTip: true })
                .getMany()
        )
    }

    /**
     * Find all repository/commit pairs that reference `value` in the given package. The
     * returned results will include only dumps that have a dependency on the given package.
     * The returned results may (but is not likely to) include a repository/commit pair that
     * does not reference `value`. See cache.ts for configuration values that tune the bloom
     * filter false positive rates. The total count of matching dumps (ignoring limit/offset)
     * is also returned.
     *
     * @param args Parameter bag.
     */
    public async getReferences({
        scheme,
        name,
        version,
        identifier,
        limit,
        offset,
    }: {
        /** The package manager scheme (e.g. npm, pip). */
        scheme: string
        /** The package name. */
        name: string
        /** The package version. */
        version: string | null
        /** The identifier to test. */
        identifier: string
        /** The maximum number of repository records to return. */
        limit: number
        /** The number of repository records to skip. */
        offset: number
    }): Promise<{ references: ReferenceModel[]; count: number }> {
        // Return all active uses of the target package
        const [results, count] = await this.withConnection(connection =>
            connection
                .getRepository(ReferenceModel)
                .createQueryBuilder('reference')
                .leftJoinAndSelect('reference.dump', 'dump')
                .where({ scheme, name, version })
                .andWhere('dump.visible_at_tip = true')
                .orderBy('dump.repository')
                .addOrderBy('dump.root')
                .limit(limit)
                .offset(offset)
                .getManyAndCount()
        )

        // Test the bloom filter of each reference model concurrently
        const keepFlags = await Promise.all(results.map(result => testFilter(result.filter, identifier)))

        for (const flag of keepFlags) {
            // Record hit and miss counts
            bloomFilterEventsCounter.labels(flag ? 'hit' : 'miss').inc()
        }

        return { references: results.filter((_, i) => keepFlags[i]), count }
    }

    /**
     * Acquire an advisory lock with the given name. This will block until the lock can be
     * acquired.
     *
     * See https://www.postgresql.org/docs/9.6/static/explicit-locking.html#ADVISORY-LOCKS.
     *
     * @param name The lock name.
     */
    public lock(name: string): Promise<void> {
        return this.withConnection(connection =>
            connection.query('SELECT pg_advisory_lock($1)', [this.generateLockId(name)])
        )
    }

    /**
     * Release an advisory lock acquired by `lock`.
     *
     * @param name The lock name.
     */
    public unlock(name: string): Promise<void> {
        return this.withConnection(connection =>
            connection.query('SELECT pg_advisory_unlock($1)', [this.generateLockId(name)])
        )
    }

    /**
     * Generate an advisory lock identifier from the given name and application salt. This is
     * based on golang-migrate's advisory lock identifier generation technique, which is in turn
     * inspired by rails migrations.
     *
     * See https://github.com/golang-migrate/migrate/blob/6c96ef02dfbf9430f7286b58afc15718588f2e13/database/util.go#L12.
     *
     * @param name The lock name.
     */
    private generateLockId(name: string): number {
        return crc32.str(name) * ADVISORY_LOCK_ID_SALT
    }

    /**
     * Invoke `callback` with a SQLite connection object obtained from the
     * cache or created on cache miss.
     *
     * @param callback The function invoke with the SQLite connection.
     */
    private withConnection<T>(callback: (connection: Connection) => Promise<T>): Promise<T> {
        return instrument(xrepoQueryDurationHistogram, xrepoQueryErrorsCounter, () => callback(this.connection))
    }

    /**
     * Invoke `callback` with a transactional SQLite manager manager object
     * obtained from the cache or created on cache miss.
     *
     * @param callback The function invoke with the entity manager.
     */
    private withTransactionalEntityManager<T>(callback: (connection: EntityManager) => Promise<T>): Promise<T> {
        return this.withConnection(connection => connection.transaction(callback))
    }

    /**
     * Update the commits tables in the cross-repository database with the current
     * output of gitserver for the given repository around the given commit. If we
     * already have commit parentage information for this commit, this function
     * will do nothing.
     *
     * @param args Parameter bag.
     */
    public async discoverAndUpdateCommit({
        repository,
        commit,
        gitserverUrls,
        ctx,
    }: {
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
        if (await this.isCommitTracked(repository, commit)) {
            return
        }

        // No need to pull commits for repos we don't have data for
        if (!(await this.isRepositoryTracked(repository))) {
            return
        }

        const gitserverUrl = addrFor(repository, gitserverUrls)
        const commits = await logAndTraceCall(ctx, 'querying commits', () =>
            getCommitsNear(gitserverUrl, repository, commit)
        )
        await logAndTraceCall(ctx, 'updating commits', () => this.updateCommits(repository, commits))
    }

    /**
     * Update the known tip of the default branch for every repository for which
     * we have LSIF data. This queries gitserver for the last known tip. From that,
     * we determine the closest commit with LSIF data and mark those as commits for
     * which we can return results in a global find-references query.
     *
     * @param args Parameter bag.
     */
    public async discoverAndUpdateTips({
        gitserverUrls,
        ctx,
        batchSize = MAX_CONCURRENT_GITSERVER_REQUESTS,
    }: {
        /** The set of ordered gitserver urls. */
        gitserverUrls: string[]
        /** The tracing context. */
        ctx: TracingContext
        /** The maximum number of requests to make at once. Set during testing.*/
        batchSize?: number
    }): Promise<void> {
        for (const [repository, commit] of (await this.discoverTips({
            gitserverUrls,
            ctx,
            batchSize,
        })).entries()) {
            await this.updateDumpsVisibleFromTip(repository, commit)
        }
    }

    /**
     * Query gitserver for the head of the default branch for every repository that has
     * LSIF data.
     *
     * @param args Parameter bag.
     */
    public async discoverTips({
        gitserverUrls,
        ctx,
        batchSize = MAX_CONCURRENT_GITSERVER_REQUESTS,
    }: {
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
        for (const repository of await this.getTrackedRepositories()) {
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
}
