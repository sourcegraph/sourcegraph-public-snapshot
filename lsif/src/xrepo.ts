import { Connection, EntityManager } from 'typeorm'
import { ConnectionCache } from './cache'
import { createFilter, testFilter } from './encoding'
import { PackageModel, ReferenceModel } from './models'
import { readEnvInt } from './util'
import { TableInserter } from './inserter'

/**
 * The maximum number of databases to open when finding global references.
 */
const GLOBAL_REFERENCE_CHUNK_SIZE = readEnvInt('LSIF_GLOBAL_REFERENCE_CHUNK_SIZE', 50)

/**
 * Represents a package provided by a project or a package that is a dependency
 * of a project, depending on its use.
 */
export interface Package {
    // The scheme of the package (e.g. npm, pip).
    scheme: string

    // The name of the package.
    name: string

    // The version of the package.
    version: string
}

/**
 * Represents a use of a set of symbols from a particular dependent package of
 * a project.
 */
export interface SymbolReferences {
    // The package from which the symbols are imported.
    package: Package

    // The unique identifiers of the symbols imported from the package.
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
    public async getPackage(scheme: string, name: string, version: string): Promise<PackageModel | undefined> {
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
     * Correlate a `repository` and `commit` with a set of unqiue packages.
     *
     * @param repository The repository that defines the given package.
     * @param commit The commit of the that defines the given package.
     * @param packages The package list (scheme, name, and version).
     */
    public async addPackages(repository: string, commit: string, packages: Package[]): Promise<void> {
        return await this.withTransactionalEntityManager(async entityManager => {
            const inserter = new TableInserter(entityManager, PackageModel, 6)
            for (const pkg of packages) {
                // TODO - upsert
                await inserter.insert({ repository, commit, ...pkg })
            }

            await inserter.finalize()
        })
    }

    /**
     * Find all repository/commit pairs that reference `value` in the given package. The
     * returned results will include only repositories that have a dependency on the given
     * package. The returned results may (but is not likely to) include a repository/commit
     * pair that does not reference `uri`. See cache.ts for configuration values that tune
     * the bloom filter false positive rates.
     *
     * @param scheme The package manager scheme (e.g. npm, pip).
     * @param name The package name.
     * @param version The package version.
     * @param value The value to test.
     * @param page The page index being requested.
     */
    public async getReferences(
        scheme: string,
        name: string,
        version: string,
        value: string,
        page: number | undefined
    ): Promise<{ references: ReferenceModel[]; nextPage: number | null }> {
        const skip = ((page || 1) - 1) * GLOBAL_REFERENCE_CHUNK_SIZE

        return await this.withConnection(connection =>
            connection.getRepository(ReferenceModel).find({
                where: {
                    scheme,
                    name,
                    version,
                },
                skip,
                take: GLOBAL_REFERENCE_CHUNK_SIZE + 1,
            })
        ).then((results: ReferenceModel[]) => {
            const hasMore = results.length > GLOBAL_REFERENCE_CHUNK_SIZE
            if (hasMore) {
                results.pop()
            }

            const filteredResults = results.filter(async result => await testFilter(result.filter, value))

            return {
                references: filteredResults,
                nextPage: hasMore ? (page || 1) + 1 : null,
            }
        })
    }

    /**
     * Correlate the given `repository` and `commit` with the the names referenced from a
     * particular remote package.
     *
     * @param repository The repository that depends on the given pacakge.
     * @param commit The commit that depends on the given pacakge.
     * @param references The package data (scheme, name, and version) and the symbosl that the package references.
     */
    public async addReferences(repository: string, commit: string, references: SymbolReferences[]): Promise<void> {
        return await this.withTransactionalEntityManager(async entityManager => {
            const inserter = new TableInserter(entityManager, ReferenceModel, 7)
            for (const reference of references) {
                // TODO - upsert
                await inserter.insert({
                    repository,
                    commit,
                    filter: await createFilter(reference.identifiers),
                    ...reference.package,
                })
            }

            await inserter.finalize()
        })
    }

    /**
     * Invoke `callback` with a SQLite connection object obtained from the
     * cache or created on cache miss.
     *
     * @param callback The function invoke with the SQLite connection.
     */
    private async withConnection<T>(callback: (connection: Connection) => Promise<T>): Promise<T> {
        return await this.connectionCache.withConnection(this.database, [PackageModel, ReferenceModel], callback)
    }

    /**
     * Invoke `callback` with a transactional SQLite manager manager object
     * obtained from the cache or created on cache miss.
     *
     * @param callback The function invoke with the entity manager.
     */
    private async withTransactionalEntityManager<T>(callback: (conenection: EntityManager) => Promise<T>): Promise<T> {
        return await this.connectionCache.withTransactionalEntityManager(
            this.database,
            [PackageModel, ReferenceModel],
            callback
        )
    }
}
