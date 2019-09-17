import {
    bloomFilterEventsCounter,
    instrument,
    xrepoInsertionDurationHistogram,
    xrepoQueryDurationHistogram,
    xrepoQueryErrorsCounter,
} from './metrics'
import { Connection, EntityManager } from 'typeorm'
import { ConnectionCache } from './cache'
import { createFilter, testFilter } from './encoding'
import { PackageModel, ReferenceModel } from './models.xrepo'
import { TableInserter } from './inserter'

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
     * Create a new ` XrepoDatabase` backed by the given database filename.
     *
     * @param connectionCache The cache of SQLite connections.
     * @param database The filename of the database.
     */
    constructor(private connectionCache: ConnectionCache, private database: string) {}

    /**
     * Find the package that defines the given `scheme`, `name`, and `version`.
     *
     * @param scheme The package manager scheme (e.g. npm, pip).
     * @param name The package name.
     * @param version The package version.
     */
    public async getPackage(scheme: string, name: string, version: string | null): Promise<PackageModel | undefined> {
        return await this.withConnection(connection =>
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
     * @param commit The commit of the being updated.
     * @param packages The list of packages that this repository defines (scheme, name, and version).
     * @param references The list of packages that this repository depends on (scheme, name, and version) and the symbols that the package references.
     */
    public async addPackagesAndReferences(
        repository: string,
        commit: string,
        packages: Package[],
        references: SymbolReferences[]
    ): Promise<void> {
        return await this.withTransactionalEntityManager(async entityManager => {
            const packageInserter = new TableInserter(
                entityManager,
                PackageModel,
                PackageModel.BatchSize,
                insertionMetrics,
                true // Do nothing on conflict
            )

            const referenceInserter = new TableInserter(
                entityManager,
                ReferenceModel,
                ReferenceModel.BatchSize,
                insertionMetrics
            )

            // Remove all previous package data for this repo/commit
            await entityManager
                .createQueryBuilder()
                .delete()
                .from(PackageModel)
                .where({ repository, commit })
                .execute()

            // Remove all previous reference data for this repo/commit
            await entityManager
                .createQueryBuilder()
                .delete()
                .from(ReferenceModel)
                .where({ repository, commit })
                .execute()

            for (const pkg of packages) {
                await packageInserter.insert({ repository, commit, ...pkg })
            }

            for (const reference of references) {
                await referenceInserter.insert({
                    repository,
                    commit,
                    filter: await createFilter(reference.identifiers),
                    ...reference.package,
                })
            }

            await packageInserter.flush()
            await referenceInserter.flush()
        })
    }

    /**
     * Find all repository/commit pairs that reference `value` in the given package. The
     * returned results will include only repositories that have a dependency on the given
     * package. The returned results may (but is not likely to) include a repository/commit
     * pair that does not reference `value`. See cache.ts for configuration values that tune
     * the bloom filter false positive rates.
     *
     * @param scheme The package manager scheme (e.g. npm, pip).
     * @param name The package name.
     * @param version The package version.
     * @param value The value to test.
     */
    public async getReferences({
        scheme,
        name,
        version,
        value,
    }: {
        scheme: string
        name: string
        version: string | null
        value: string
    }): Promise<ReferenceModel[]> {
        const results = await this.withConnection(connection =>
            connection.getRepository(ReferenceModel).find({
                where: {
                    scheme,
                    name,
                    version,
                },
            })
        )

        // Test the bloom filter of each reference model concurrently
        const keepFlags = await Promise.all(results.map(result => testFilter(result.filter, value)))

        for (const flag of keepFlags) {
            // Record hit and miss counts
            bloomFilterEventsCounter.labels(flag ? 'hit' : 'miss').inc()
        }

        return results.filter((_, i) => keepFlags[i])
    }

    /**
     * Invoke `callback` with a SQLite connection object obtained from the
     * cache or created on cache miss.
     *
     * @param callback The function invoke with the SQLite connection.
     */
    private async withConnection<T>(callback: (connection: Connection) => Promise<T>): Promise<T> {
        return await this.connectionCache.withConnection(this.database, [PackageModel, ReferenceModel], connection =>
            instrument(xrepoQueryDurationHistogram, xrepoQueryErrorsCounter, () => callback(connection))
        )
    }

    /**
     * Invoke `callback` with a transactional SQLite manager manager object
     * obtained from the cache or created on cache miss.
     *
     * @param callback The function invoke with the entity manager.
     */
    private async withTransactionalEntityManager<T>(callback: (connection: EntityManager) => Promise<T>): Promise<T> {
        return await this.connectionCache.withTransactionalEntityManager(
            this.database,
            [PackageModel, ReferenceModel],
            callback
        )
    }
}
