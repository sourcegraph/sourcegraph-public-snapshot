import * as metrics from './metrics'
import * as sharedMetrics from '../database/metrics'
import * as xrepoModels from '../models/xrepo'
import { addrFor, getCommitsNear, gitserverExecLines } from './commits'
import { MAX_TRAVERSAL_LIMIT } from '../constants'
import { Brackets, Connection, EntityManager } from 'typeorm'
import { createFilter, testFilter } from './bloom-filter'
import { dbFilename, tryDeleteFile } from '../paths'
import { logAndTraceCall, logSpan, TracingContext } from '../tracing'
import { instrumentQuery, withInstrumentedTransaction } from '../database/postgres'
import { TableInserter } from '../database/inserter'

/**
 * The insertion metrics for the cross-repo database.
 */
const insertionMetrics = {
    durationHistogram: metrics.xrepoInsertionDurationHistogram,
    errorsCounter: sharedMetrics.xrepoQueryErrorsCounter,
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
 * A wrapper around the cross-repository database that stitches together the references
 * between projects at a specific commit. This is used for cross-repository jump to
 * definition and find references features.
 */
export class XrepoDatabase {
    /**
     * Create a new `XrepoDatabase` backed by the given database connection.
     *
     * @param connection The Postgres connection.
     * @param storageRoot The path where SQLite databases are stored.
     */
    constructor(private connection: Connection, private storageRoot: string) {}

    /**
     * Get the dumps for a repository.
     *
     * @param repository The repository.
     * @param query A search query.
     * @param visibleAtTip If true, only return dumps visible at tip.
     * @param limit The maximum number of dumps to return.
     * @param offset The number of dumps to skip.
     */
    public async getDumps(
        repository: string,
        query: string,
        visibleAtTip: boolean,
        limit: number,
        offset: number
    ): Promise<{ dumps: xrepoModels.LsifDump[]; totalCount: number }> {
        const [dumps, totalCount] = await instrumentQuery(() => {
            let queryBuilder = this.connection
                .getRepository(xrepoModels.LsifDump)
                .createQueryBuilder()
                .where({ repository })
                .orderBy('uploaded_at', 'DESC')
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

            if (visibleAtTip) {
                queryBuilder = queryBuilder.andWhere('visible_at_tip = true')
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
    public getDumpById(id: xrepoModels.DumpId): Promise<xrepoModels.LsifDump | undefined> {
        return instrumentQuery(() => this.connection.getRepository(xrepoModels.LsifDump).findOne({ id }))
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
     * @param repository The repository name.
     * @param commit The head of the default branch.
     * @param ctx The tracing context.
     */
    public updateDumpsVisibleFromTip(repository: string, commit: string, ctx: TracingContext = {}): Promise<void> {
        const query = `
            -- Get all ancestors of the tip
            WITH RECURSIVE lineage(id, repository, "commit", parent) AS (
                SELECT c.* FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
                UNION
                SELECT c.* FROM lineage a JOIN lsif_commits c ON a.repository = c.repository AND a.parent = c."commit"
            ),
            ${visibleDumps()}

            -- Update dump records by:
            --   (1) unsetting the visibility flag of all previously visible dumps, and
            --   (2) setting the visibility flag of all currently visible dumps
            UPDATE lsif_dumps d
            SET visible_at_tip = id IN (SELECT * from visible_ids)
            WHERE d.repository = $1 AND (d.id IN (SELECT * from visible_ids) OR d.visible_at_tip)
        `

        return logAndTraceCall(ctx, 'Updating dumps visible from tip', () =>
            instrumentQuery(() => this.connection.query(query, [repository, commit]))
        )
    }

    /**
     * Return the dump 'closest' to the given target commit (a direct descendant or ancestor of
     * the target commit). If no closest commit can be determined, this method returns undefined.
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
    ): Promise<xrepoModels.LsifDump | undefined> {
        // Request updated commit data from gitserver if this commit isn't
        // already tracked. This will pull back ancestors for this commit
        // up to a certain (configurable) depth and insert them into the
        // cross-repository database. This populates the necessary data for
        // the following query.
        if (gitserverUrls) {
            await this.updateCommits(
                repository,
                await this.discoverCommits({ repository, commit, gitserverUrls, ctx }),
                ctx
            )
        }

        return logAndTraceCall(ctx, 'Finding closest dump', async () => {
            const results: xrepoModels.LsifDump[] = await instrumentQuery(() =>
                this.connection.query('select * from closest_dump($1, $2, $3, $4)', [
                    repository,
                    commit,
                    file,
                    MAX_TRAVERSAL_LIMIT,
                ])
            )

            if (results.length === 0) {
                return undefined
            }

            return results[0]
        })
    }

    /**
     * Update the known commits for a repository. The input commits must be a map from commits to
     * a set of parent commits. Commits without a parent should have an empty set of parents, but
     * should still be present in the map.
     *
     * @param repository The repository name.
     * @param commits The commit parentage data.
     * @param ctx The tracing context.
     */
    public updateCommits(
        repository: string,
        commits: Map<string, Set<string>>,
        ctx: TracingContext = {}
    ): Promise<void> {
        return logAndTraceCall(ctx, 'Updating commits', () =>
            withInstrumentedTransaction(this.connection, async entityManager => {
                const commitInserter = new TableInserter(
                    entityManager,
                    xrepoModels.Commit,
                    xrepoModels.Commit.BatchSize,
                    insertionMetrics,
                    true // Do nothing on conflict
                )

                for (const [commit, parentCommits] of commits) {
                    if (parentCommits.size === 0) {
                        await commitInserter.insert({ repository, commit, parentCommit: null })
                    }

                    for (const parentCommit of parentCommits) {
                        await commitInserter.insert({ repository, commit, parentCommit })
                    }
                }

                await commitInserter.flush()
            })
        )
    }

    /**
     * Find the package that defines the given `scheme`, `name`, and `version`.
     *
     * @param scheme The package manager scheme (e.g. npm, pip).
     * @param name The package name.
     * @param version The package version.
     */
    public getPackage(
        scheme: string,
        name: string,
        version: string | null
    ): Promise<xrepoModels.PackageModel | undefined> {
        return instrumentQuery(() =>
            this.connection.getRepository(xrepoModels.PackageModel).findOne({
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
     * @param uploadedAt The time the dump was uploaded.
     * @param packages The list of packages that this repository defines (scheme, name, and version).
     * @param references The list of packages that this repository depends on (scheme, name, and version) and the symbols that the package references.
     * @param ctx The tracing context.
     */
    public addPackagesAndReferences(
        repository: string,
        commit: string,
        root: string,
        uploadedAt: Date,
        packages: Package[],
        references: SymbolReferences[],
        ctx: TracingContext = {}
    ): Promise<xrepoModels.LsifDump> {
        return withInstrumentedTransaction(this.connection, async entityManager => {
            const dump = await logAndTraceCall(ctx, 'Inserting dump', () =>
                this.insertDump(repository, commit, root, uploadedAt, entityManager)
            )

            await logAndTraceCall(ctx, 'Inserting packages', async () => {
                const packageInserter = new TableInserter<xrepoModels.PackageModel, new () => xrepoModels.PackageModel>(
                    entityManager,
                    xrepoModels.PackageModel,
                    xrepoModels.PackageModel.BatchSize,
                    insertionMetrics,
                    true // Do nothing on conflict
                )

                for (const pkg of packages) {
                    await packageInserter.insert({ dump_id: dump.id, ...pkg })
                }

                await packageInserter.flush()
            })

            await logAndTraceCall(ctx, 'Inserting references', async () => {
                const referenceInserter = new TableInserter<
                    xrepoModels.ReferenceModel,
                    new () => xrepoModels.ReferenceModel
                >(entityManager, xrepoModels.ReferenceModel, xrepoModels.ReferenceModel.BatchSize, insertionMetrics)

                for (const reference of references) {
                    await referenceInserter.insert({
                        dump_id: dump.id,
                        filter: await createFilter(reference.identifiers),
                        ...reference.package,
                    })
                }

                await referenceInserter.flush()
            })

            return dump
        })
    }

    /**
     * Inserts the given repository and commit into the `lsif_dumps` table.
     *
     * @param repository The repository.
     * @param commit The commit.
     * @param root The root of all files that are in this dump.
     * @param uploadedAt The time the dump was uploaded.
     * @param entityManager The EntityManager for the connection to the xrepo database.
     */
    public async insertDump(
        repository: string,
        commit: string,
        root: string,
        uploadedAt: Date = new Date(),
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<xrepoModels.LsifDump> {
        // Get existing dumps from the same repo@commit that overlap with the current
        // root (where the existing root is a prefix of the current root, or vice versa).

        const dumps = await entityManager
            .getRepository(xrepoModels.LsifDump)
            .createQueryBuilder()
            .select()
            .where({ repository, commit })
            .andWhere(
                new Brackets(qb =>
                    qb.where(":root LIKE (root || '%')", { root }).orWhere("root LIKE (:root || '%')", { root })
                )
            )
            .getMany()

        for (const dump of dumps) {
            await this.deleteDump(dump, entityManager)
        }

        const dump = new xrepoModels.LsifDump()
        dump.repository = repository
        dump.commit = commit
        dump.root = root
        dump.uploadedAt = uploadedAt
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
    public getDump(repository: string, commit: string, file: string): Promise<xrepoModels.LsifDump | undefined> {
        return instrumentQuery(() =>
            this.connection
                .getRepository(xrepoModels.LsifDump)
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
    public getOldestPrunableDump(): Promise<xrepoModels.LsifDump | undefined> {
        return instrumentQuery(() =>
            this.connection
                .getRepository(xrepoModels.LsifDump)
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
        dump: xrepoModels.LsifDump,
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

        await entityManager.getRepository(xrepoModels.LsifDump).delete(dump.id)
    }

    /**
     * Find the visible dumps. This method is used for testing.
     *
     * @param repository The repository.
     */
    public getVisibleDumps(repository: string): Promise<xrepoModels.LsifDump[]> {
        return instrumentQuery(() =>
            this.connection
                .getRepository(xrepoModels.LsifDump)
                .createQueryBuilder()
                .select()
                .where({ repository, visibleAtTip: true })
                .getMany()
        )
    }

    /**
     * Find all references pointing to the given identifier in the given package. The returned
     * results may (but is not likely to) include a repository/commit pair that does not reference
     * the given identifier. See cache.ts for configuration values that tune the bloom filter false
     * positive rates. The total count of matching dumps, that ignores limit and offset, is also
     * returned.
     *
     * This method does NOT include any dumps for the given repository.
     *
     * @param args Parameter bag.
     */
    public getReferences({
        repository,
        scheme,
        name,
        version,
        identifier,
        limit,
        offset,
        ctx = {},
    }: {
        /** The source repository of the search. */
        repository: string
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
        /** The tracing context. */
        ctx?: TracingContext
    }): Promise<{ references: xrepoModels.ReferenceModel[]; totalCount: number; newOffset: number }> {
        // We do this inside of a transaction so that we get consistent results from multiple
        // distinct queries: one count query and one or more select queries, depending on the
        // sparsity of the use of the given identifier.
        return withInstrumentedTransaction(this.connection, async entityManager => {
            // Create a base query that selects all active uses of the target package. This
            // is used as the common prefix for both the count and the getPage queries.
            const baseQuery = entityManager
                .getRepository(xrepoModels.ReferenceModel)
                .createQueryBuilder('reference')
                .leftJoinAndSelect('reference.dump', 'dump')
                .where({ scheme, name, version })
                .andWhere('dump.repository != :repository', { repository })
                .andWhere('dump.visible_at_tip = true')

            // Get total number of items in this set of results
            const totalCount = await baseQuery.getCount()

            // Construct method to select a page of possible references
            const getPage = (pageOffset: number): Promise<xrepoModels.ReferenceModel[]> =>
                baseQuery
                    .orderBy('dump.repository')
                    .addOrderBy('dump.root')
                    .limit(limit)
                    .offset(pageOffset)
                    .getMany()

            // Invoke getPage with increasing offsets until we get a page size's worth of
            // references that actually use the given identifier as indicated by result of
            // the bloom filter query.
            const { references, newOffset } = await this.gatherReferences({
                getPage,
                identifier,
                offset,
                limit,
                totalCount,
                ctx,
            })

            return { references, totalCount, newOffset }
        })
    }

    /**
     * Find all references pointing to the given identifier in the given package within dumps of the
     * given repository. The returned results may (but is not likely to) include a repository/commit
     * pair that does not reference  the given identifier. See cache.ts for configuration values that
     * tune the bloom filter false positive rates. The total count of matching dumps, that ignores limit
     * and offset. is also returned.
     *
     * @param args Parameter bag.
     */
    public getSameRepoRemoteReferences({
        repository,
        commit,
        scheme,
        name,
        version,
        identifier,
        limit,
        offset,
        ctx = {},
    }: {
        /** The source repository of the search. */
        repository: string
        /** The commit of the references query. */
        commit: string
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
        /** The tracing context. */
        ctx?: TracingContext
    }): Promise<{ references: xrepoModels.ReferenceModel[]; totalCount: number; newOffset: number }> {
        const visibleIdsQuery = `
            -- lineage is a recursively defined CTE that returns all ancestor an descendants
            -- of the given commit for the given repository. This happens to evaluate in
            -- Postgres as a lazy generator, which allows us to pull the "next" closest commit
            -- in either direction from the source commit as needed.
            WITH RECURSIVE lineage(id, repository, "commit", parent_commit, direction) AS (
                SELECT l.* FROM (
                    -- seed recursive set with commit looking in ancestor direction
                    SELECT c.*, 'A' FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
                    UNION
                    -- seed recursive set with commit looking in descendant direction
                    SELECT c.*, 'D' FROM lsif_commits c WHERE c.repository = $1 AND c."commit" = $2
                ) l

                UNION

                SELECT * FROM (
                    WITH l_inner AS (SELECT * FROM lineage)
                    -- get next ancestor
                    SELECT c.*, l.direction FROM l_inner l JOIN lsif_commits c ON l.direction = 'A' AND c.repository = l.repository AND c."commit" = l.parent_commit
                    UNION
                    -- get next descendant
                    SELECT c.*, l.direction FROM l_inner l JOIN lsif_commits c ON l.direction = 'D' and c.repository = l.repository AND c.parent_commit = l."commit"
                ) subquery
            ),
            ${visibleDumps()}
            SELECT * FROM visible_ids
        `

        const countQuery = `
            SELECT count(*) FROM lsif_references r
            WHERE r.scheme = $1 AND r.name = $2 AND r.version = $3 AND r.dump_id = ANY($4)
        `

        const referenceIdsQuery = `
            SELECT r.id FROM lsif_references r
            LEFT JOIN lsif_dumps d on r.dump_id = d.id
            WHERE r.scheme = $1 AND r.name = $2 AND r.version = $3 AND r.dump_id = ANY($4)
            ORDER BY d.root OFFSET $5 LIMIT $6
        `

        // We do this inside of a transaction so that we get consistent results from multiple
        // distinct queries: one count query and one or more select queries, depending on the
        // sparsity of the use of the given identifier.
        return withInstrumentedTransaction(this.connection, async entityManager => {
            // Extract numeric ids from query results that return objects
            const extractIds = (results: { id: number }[]): number[] => results.map(r => r.id)

            // We need the set of identifiers for visible lsif dumps for both the count
            // and the getPage queries. The results of this query do not change based on
            // the page size or offset, so we query it separately here and pass the result
            // as a parameter.
            const visible_ids = extractIds(await entityManager.query(visibleIdsQuery, [repository, commit]))

            // Get total number of items in this set of results
            const rawCount: { count: string }[] = await entityManager.query(countQuery, [
                scheme,
                name,
                version,
                visible_ids,
            ])

            // Oddly, this comes back as a string value in the result set
            const totalCount = parseInt(rawCount[0].count, 10)

            // Construct method to select a page of possible references. We first perform
            // the query defined above that returns reference identifiers, then perform a
            // second query to select the models by id so that we load the relationships.
            const getPage = async (pageOffset: number): Promise<xrepoModels.ReferenceModel[]> => {
                const args = [scheme, name, version, visible_ids, pageOffset, limit]
                const results = await entityManager.query(referenceIdsQuery, args)
                const referenceIds = extractIds(results)
                const references = await entityManager.getRepository(xrepoModels.ReferenceModel).findByIds(referenceIds)

                // findByIds doesn't return models in the same order as they were requested,
                // so we need to sort them here before returning.

                const modelsById = new Map(references.map(r => [r.id, r]))
                return referenceIds
                    .map(id => modelsById.get(id))
                    .filter(<T>(x: T | undefined): x is T => x !== undefined)
            }

            // Invoke getPage with increasing offsets until we get a page size's worth of
            // references that actually use the given identifier as indicated by result of
            // the bloom filter query.
            const { references, newOffset } = await this.gatherReferences({
                getPage,
                identifier,
                offset,
                limit,
                totalCount,
                ctx,
            })

            return { references, totalCount, newOffset }
        })
    }

    /**
     * Select a page of possible results via the `getPage` function and collect the references that include
     * a use of the given identifier. As the given results may depend on the target package but not import
     * the given identifier, the remaining set of references may small (or empty). In order to get a full
     * page of results, we repeat the process until we have the proper number of results.
     *
     * @param args Parameter bag.
     */
    private async gatherReferences({
        getPage,
        identifier,
        offset,
        limit,
        totalCount,
        ctx = {},
    }: {
        /** The function to invoke to query the next set of references. */
        getPage: (offset: number) => Promise<xrepoModels.ReferenceModel[]>
        /** The identifier to test. */
        identifier: string
        /** The maximum number of repository records to return. */
        offset: number
        /** The number of repository records to skip. */
        limit: number
        /**
         * The total number of items in this set of results. Alternatively, the maximum
         * number of items that can be returned by `getPage` starting from offset zero.
         */
        totalCount: number
        /** The tracing context. */
        ctx?: TracingContext
    }): Promise<{ references: xrepoModels.ReferenceModel[]; newOffset: number }> {
        return logAndTraceCall(ctx, 'Gathering references', async ctx => {
            let numScanned = 0
            let numFetched = 0
            let numFiltered = 0
            let newOffset = offset
            const references: xrepoModels.ReferenceModel[] = []

            while (references.length < limit && newOffset < totalCount) {
                // Copy for use in the following anonymous function, otherwise the
                // re-assignment of newOffset triggers a non-atomic update warning.
                const localOffset = newOffset

                const page = await logAndTraceCall(ctx, 'Fetching page of references', () => getPage(localOffset))
                if (page.length === 0) {
                    // Shouldn't happen, but just in case of a bug we
                    // don't want this to throw up into an infinite loop.
                    break
                }

                const { references: filtered, scanned } = await this.applyBloomFilter(
                    page,
                    identifier,
                    limit - references.length
                )

                for (const reference of filtered) {
                    references.push(reference)
                }

                newOffset += scanned
                numScanned += scanned
                numFetched += page.length
                numFiltered += scanned - filtered.length
            }

            logSpan(ctx, 'reference_results', { numScanned, numFetched, numFiltered })
            return { references, newOffset }
        })
    }

    /**
     * Filter out the references which do not contain the given identifier in their bloom filter. Returns at most
     * `limit` values in the return array and also the number of references that were checked (left to right).
     *
     * @param references The set of references to filter.
     * @param identifier The identifier to test.
     * @param limit The maximum number of references to return.
     */
    private async applyBloomFilter(
        references: xrepoModels.ReferenceModel[],
        identifier: string,
        limit: number
    ): Promise<{ references: xrepoModels.ReferenceModel[]; scanned: number }> {
        // Test the bloom filter of each reference model concurrently
        const keepFlags = await Promise.all(references.map(result => testFilter(result.filter, identifier)))

        const filtered = []
        for (const [index, flag] of keepFlags.entries()) {
            // Record hit and miss counts
            metrics.bloomFilterEventsCounter.labels(flag ? 'hit' : 'miss').inc()

            if (flag) {
                filtered.push(references[index])

                if (filtered.length >= limit) {
                    // We got enough - stop scanning here and return the number of
                    // results we actually went through so we can compute an offset
                    // for the next page of results that don't skip the remainder
                    // of this set of results.
                    return { references: filtered, scanned: index + 1 }
                }
            }
        }

        // We scanned the entire set of references
        return { references: filtered, scanned: references.length }
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
        repository,
        commit,
        gitserverUrls,
        ctx = {},
    }: {
        /** The repository name. */
        repository: string
        /** The commit from which the gitserver queries should start. */
        commit: string
        /** The set of ordered gitserver urls. */
        gitserverUrls: string[]
        /** The tracing context. */
        ctx?: TracingContext
    }): Promise<Map<string, Set<string>>> {
        const matchingRepos = await instrumentQuery(() =>
            this.connection.getRepository(xrepoModels.LsifDump).count({ where: { repository } })
        )
        if (matchingRepos === 0) {
            return new Map()
        }

        const matchingCommits = await instrumentQuery(() =>
            this.connection.getRepository(xrepoModels.Commit).count({ where: { repository, commit } })
        )
        if (matchingCommits > 0) {
            return new Map()
        }

        return getCommitsNear(addrFor(repository, gitserverUrls), repository, commit, ctx)
    }

    /**
     * Query gitserver for the head of the default branch for the given repository.
     *
     * @param args Parameter bag.
     */
    public async discoverTip({
        repository,
        gitserverUrls,
        ctx = {},
    }: {
        /** The repository name. */
        repository: string
        /** The set of ordered gitserver urls. */
        gitserverUrls: string[]
        /** The tracing context. */
        ctx?: TracingContext
    }): Promise<string | undefined> {
        const lines = await logAndTraceCall(ctx, 'Getting repository metadata', () =>
            gitserverExecLines(addrFor(repository, gitserverUrls), repository, ['rev-parse', 'HEAD'], ctx)
        )

        return lines ? lines[0] : undefined
    }
}

/**
 * Return a set of CTE definitions assuming the definition of a previous CTE named `lineage`.
 * This creates the CTE `visible_ids`, which gathers the set of LSIF dump identifiers whose
 * commit occurs in `lineage` (within the given traversal limit) and whose root does not
 * overlap another visible dump.
 *
 * @param limit The maximum number of dumps that can be extracted from `lineage`.
 */
function visibleDumps(limit: number = MAX_TRAVERSAL_LIMIT): string {
    return `
        -- Limit the visibility to the maximum traversal depth and approximate
        -- each commit's depth by its row number.
        limited_lineage AS (
            SELECT a.*, row_number() OVER() as n from lineage a LIMIT ${limit}
        ),
        -- Correlate commits to dumps and filter out commits without LSIF data
        lineage_with_dumps AS (
            SELECT a.*, d.root, d.id as dump_id FROM limited_lineage a
            JOIN lsif_dumps d ON d.repository = a.repository AND d."commit" = a."commit"
        ),
        visible_ids AS (
            -- Remove dumps where there exists another visible dump of smaller depth with an overlapping root.
            -- Such dumps would not be returned with a closest commit query so we don't want to return results
            -- for them in global find-reference queries either.
            SELECT DISTINCT t1.dump_id as id FROM lineage_with_dumps t1 WHERE NOT EXISTS (
                SELECT 1 FROM lineage_with_dumps t2
                WHERE t2.n < t1.n AND (
                    t2.root LIKE (t1.root || '%') OR
                    t1.root LIKE (t2.root || '%')
                )
            )
        )
    `
}
