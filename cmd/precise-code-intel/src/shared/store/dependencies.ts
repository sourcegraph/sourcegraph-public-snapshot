import * as metrics from './metrics'
import * as sharedMetrics from '../database/metrics'
import * as pgModels from '../models/pg'
import { Connection, EntityManager } from 'typeorm'
import { createFilter, testFilter } from '../datastructures/bloom-filter'
import { instrumentQuery, withInstrumentedTransaction } from '../database/postgres'
import { logAndTraceCall, logSpan, TracingContext } from '../tracing'
import { TableInserter } from '../database/inserter'
import { visibleDumps, bidirectionalLineage } from '../models/queries'

/** The insertion metrics for Postgres. */
const insertionMetrics = {
    durationHistogram: sharedMetrics.postgresInsertionDurationHistogram,
    errorsCounter: sharedMetrics.postgresQueryErrorsCounter,
}

/**
 * Represents a package provided by a project or a package that is a dependency
 * of a project, depending on its use.
 */
export interface Package {
    /** The scheme of the package (e.g. npm, pip). */
    scheme: string

    /** The name of the package. */
    name: string

    /** The version of the package. */
    version: string | null
}

/** Represents a use of a set of symbols from a particular dependent package of a project. */
export interface SymbolReferences {
    /** The package from which the symbols are imported. */
    package: Package

    /** The unique identifiers of the symbols imported from the package. */
    identifiers: string[]
}

/**
 * A wrapper around package and references tables that stitch together the references
 * between projects at a specific commit. This is used for cross-repository jump to
 * definition and find references features.
 */
export class DependencyManager {
    /**
     * Create a new `DependencyManager` backed by the given database connection.
     *
     * @param connection The Postgres connection.
     */
    constructor(private connection: Connection) {}

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
    ): Promise<pgModels.PackageModel | undefined> {
        return instrumentQuery(() =>
            this.connection.getRepository(pgModels.PackageModel).findOne({
                where: {
                    scheme,
                    name,
                    version,
                },
            })
        )
    }

    /**
     * Find all package references pointing to the given identifier in the given package. The returned
     * results may (but is not likely to) include a repository/commit pair that does not reference
     * the given identifier. See cache.ts for configuration values that tune the bloom filter false
     * positive rates. The total count of matching dumps, that ignores limit and offset, is also
     * returned.
     *
     * This method does NOT include any dumps for the given repository.
     *
     * @param args Parameter bag.
     */
    public getPackageReferences({
        repositoryId,
        scheme,
        name,
        version,
        identifier,
        limit,
        offset,
        ctx = {},
    }: {
        /** The identifier of the source repository of the search. */
        repositoryId: number
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
    }): Promise<{ packageReferences: pgModels.ReferenceModel[]; totalCount: number; newOffset: number }> {
        // We do this inside of a transaction so that we get consistent results from multiple
        // distinct queries: one count query and one or more select queries, depending on the
        // sparsity of the use of the given identifier.
        return withInstrumentedTransaction(this.connection, async entityManager => {
            // Create a base query that selects all active uses of the target package. This
            // is used as the common prefix for both the count and the getPage queries.
            const baseQuery = entityManager
                .getRepository(pgModels.ReferenceModel)
                .createQueryBuilder('reference')
                .leftJoinAndSelect('reference.dump', 'dump')
                .where({ scheme, name, version })
                .andWhere('dump.repository_id != :repositoryId', { repositoryId })
                .andWhere('dump.visible_at_tip = true')

            // Get total number of items in this set of results
            const totalCount = await baseQuery.getCount()

            // Construct method to select a page of possible package references
            const getPage = (pageOffset: number): Promise<pgModels.ReferenceModel[]> =>
                baseQuery
                    .orderBy('dump.repository_id')
                    .addOrderBy('dump.root')
                    .limit(limit)
                    .offset(pageOffset)
                    .getMany()

            // Invoke getPage with increasing offsets until we get a page size's worth of
            // package references that actually use the given identifier as indicated by result
            // of the bloom filter query.
            const { packageReferences, newOffset } = await this.gatherPackageReferences({
                getPage,
                identifier,
                offset,
                limit,
                totalCount,
                ctx,
            })

            return { packageReferences, totalCount, newOffset }
        })
    }

    /**
     * Find all package references pointing to the given identifier in the given package within dumps of
     * the given repository. The returned results may (but is not likely to) include a repository/commit
     * pair that does not reference  the given identifier. See cache.ts for configuration values that
     * tune the bloom filter false positive rates. The total count of matching dumps, that ignores limit
     * and offset. is also returned.
     *
     * @param args Parameter bag.
     */
    public getSameRepoRemotePackageReferences({
        repositoryId,
        commit,
        scheme,
        name,
        version,
        identifier,
        limit,
        offset,
        ctx = {},
    }: {
        /** The identifier of the source repository of the search. */
        repositoryId: number
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
    }): Promise<{ packageReferences: pgModels.ReferenceModel[]; totalCount: number; newOffset: number }> {
        const visibleIdsQuery = `
            WITH
            ${bidirectionalLineage()},
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
            const visible_ids = extractIds(await entityManager.query(visibleIdsQuery, [repositoryId, commit]))

            // Get total number of items in this set of results
            const rawCount: { count: string }[] = await entityManager.query(countQuery, [
                scheme,
                name,
                version,
                visible_ids,
            ])

            // Oddly, this comes back as a string value in the result set
            const totalCount = parseInt(rawCount[0].count, 10)

            // Construct method to select a page of possible package references. We first
            // perform the query defined above that returns reference identifiers, then
            // perform a second query to select the models by id so that we load the
            // relationships.
            const getPage = async (pageOffset: number): Promise<pgModels.ReferenceModel[]> => {
                const args = [scheme, name, version, visible_ids, pageOffset, limit]
                const results = await entityManager.query(referenceIdsQuery, args)
                const referenceIds = extractIds(results)
                const packageReferences = await entityManager
                    .getRepository(pgModels.ReferenceModel)
                    .findByIds(referenceIds)

                // findByIds doesn't return models in the same order as they were requested,
                // so we need to sort them here before returning.

                const modelsById = new Map(packageReferences.map(r => [r.id, r]))
                return referenceIds
                    .map(id => modelsById.get(id))
                    .filter(<T>(x: T | undefined): x is T => x !== undefined)
            }

            // Invoke getPage with increasing offsets until we get a page size's worth of
            // package references that actually use the given identifier as indicated by
            // result of the bloom filter query.
            const { packageReferences, newOffset } = await this.gatherPackageReferences({
                getPage,
                identifier,
                offset,
                limit,
                totalCount,
                ctx,
            })

            return { packageReferences, totalCount, newOffset }
        })
    }

    /**
     * Correlate a `repository` and `commit` with a set of unique packages it defines and
     * with  the the names referenced from a particular dependent package.
     *
     * @param dumpId The identifier of the newly inserted dump.
     * @param packages The list of packages that this repository defines (scheme, name, and version).
     * @param symbolReferences The list of packages that this repository depends on (scheme, name, and version) and the symbols that the package references.
     * @param ctx The tracing context.
     * @param entityManager The EntityManager to use as part of a transaction.
     */
    public async addPackagesAndReferences(
        dumpId: number,
        packages: Package[],
        symbolReferences: SymbolReferences[],
        ctx: TracingContext = {},
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<void> {
        await logAndTraceCall(ctx, 'Inserting packages', async () => {
            const packageInserter = new TableInserter<pgModels.PackageModel, new () => pgModels.PackageModel>(
                entityManager,
                pgModels.PackageModel,
                pgModels.PackageModel.BatchSize,
                insertionMetrics,
                true // Do nothing on conflict
            )

            for (const pkg of packages) {
                await packageInserter.insert({ dump_id: dumpId, ...pkg })
            }

            await packageInserter.flush()
        })

        await logAndTraceCall(ctx, 'Inserting references', async () => {
            const referenceInserter = new TableInserter<pgModels.ReferenceModel, new () => pgModels.ReferenceModel>(
                entityManager,
                pgModels.ReferenceModel,
                pgModels.ReferenceModel.BatchSize,
                insertionMetrics
            )

            for (const { package: pkg, identifiers } of symbolReferences) {
                await referenceInserter.insert({
                    dump_id: dumpId,
                    filter: await createFilter(identifiers),
                    ...pkg,
                })
            }

            await referenceInserter.flush()
        })
    }

    /**
     * Select a page of possible results via the `getPage` function and collect the package references that
     * include a use of the given identifier. As the given results may depend on the target package but not
     * import the given identifier, the remaining set of package references may small (or empty). In order
     * to get a full page of results, we repeat the process until we have the proper number of results.
     *
     * @param args Parameter bag.
     */
    private async gatherPackageReferences({
        getPage,
        identifier,
        offset,
        limit,
        totalCount,
        ctx = {},
    }: {
        /** The function to invoke to query the next set of package references. */
        getPage: (offset: number) => Promise<pgModels.ReferenceModel[]>
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
    }): Promise<{ packageReferences: pgModels.ReferenceModel[]; newOffset: number }> {
        return logAndTraceCall(ctx, 'Gathering package references', async ctx => {
            let numScanned = 0
            let numFetched = 0
            let numFiltered = 0
            let newOffset = offset
            const packageReferences: pgModels.ReferenceModel[] = []

            while (packageReferences.length < limit && newOffset < totalCount) {
                // Copy for use in the following anonymous function, otherwise the
                // re-assignment of newOffset triggers a non-atomic update warning.
                const localOffset = newOffset

                const page = await logAndTraceCall(ctx, 'Fetching page of package references', () =>
                    getPage(localOffset)
                )
                if (page.length === 0) {
                    // Shouldn't happen, but just in case of a bug we
                    // don't want this to throw up into an infinite loop.
                    break
                }

                const { packageReferences: filteredPackageReferences, scanned } = await this.applyBloomFilter(
                    page,
                    identifier,
                    limit - packageReferences.length
                )

                for (const packageReference of filteredPackageReferences) {
                    packageReferences.push(packageReference)
                }

                newOffset += scanned
                numScanned += scanned
                numFetched += page.length
                numFiltered += scanned - filteredPackageReferences.length
            }

            logSpan(ctx, 'reference_results', { numScanned, numFetched, numFiltered })
            return { packageReferences, newOffset }
        })
    }

    /**
     * Filter out the package references which do not contain the given identifier in their bloom filter.
     * Returns at most `limit` values in the return array and also the number of package references that
     * were checked (left to right).
     *
     * @param packageReferences The set of package references to filter.
     * @param identifier The identifier to test.
     * @param limit The maximum number of package references to return.
     */
    private async applyBloomFilter(
        packageReferences: pgModels.ReferenceModel[],
        identifier: string,
        limit: number
    ): Promise<{ packageReferences: pgModels.ReferenceModel[]; scanned: number }> {
        // Test the bloom filter of each reference model concurrently
        const keepFlags = await Promise.all(packageReferences.map(result => testFilter(result.filter, identifier)))

        const filtered = []
        for (const [index, flag] of keepFlags.entries()) {
            // Record hit and miss counts
            metrics.bloomFilterEventsCounter.labels(flag ? 'hit' : 'miss').inc()

            if (flag) {
                filtered.push(packageReferences[index])

                if (filtered.length >= limit) {
                    // We got enough - stop scanning here and return the number of
                    // results we actually went through so we can compute an offset
                    // for the next page of results that don't skip the remainder
                    // of this set of results.
                    return { packageReferences: filtered, scanned: index + 1 }
                }
            }
        }

        // We scanned the entire set of package references
        return { packageReferences: filtered, scanned: packageReferences.length }
    }
}
