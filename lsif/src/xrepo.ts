import { Connection, EntityManager } from 'typeorm'
import { testFilter, createFilter } from './encoding'
import { ConnectionCache } from './cache'
import { ReferenceModel, PackageModel } from './models'
import { Package, SymbolReferences } from './entities'
import { Inserter } from './inserter'

/**
 * `XrepoDatabase` is a SQLite database that stitches together the references
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
     * TODO - redocument
     * Correlate a `repository` and `commit` with a unique package manager `scheme`, `name`, and `version`.
     *
     * @param scheme The package manager scheme (e.g. npm, pip).
     * @param name The package name.
     * @param version The package version.
     * @param repository The repository that defines the given package.
     * @param commit The commit of the that defines the given package.
     */
    public async addPackages(repository: string, commit: string, packages: Package[]): Promise<void> {
        return await this.withTransactionalEntityManager(async entityManager => {
            const inserter = new Inserter(entityManager, PackageModel, 6)
            for (const pkg of packages) {
                // TODO - upsert
                await inserter.insert({ repository, commit, ...pkg })
            }

            await inserter.finalize()
        })
    }

    /**
     * Find all repository/commit pairs that reference `uri` in the given package. The
     * returned results will include only repositories that have a dependency on the given
     * package. The returned results may (but is not likely to) include a repository/commit
     * pair that does not reference `uri`. See cache.ts for configuration values that tune
     * the bloom filter false positive rates.
     *
     * @param scheme The package manager scheme (e.g. npm, pip).
     * @param name The package name.
     * @param version The package version.
     * @param uri The uri to test.
     */
    public async getReferences(scheme: string, name: string, version: string, uri: string): Promise<ReferenceModel[]> {
        return await this.withConnection(connection =>
            connection
                .getRepository(ReferenceModel)
                .find({
                    where: {
                        scheme,
                        name,
                        version,
                    },
                })
                .then((results: ReferenceModel[]) =>
                    results.filter(async result => await testFilter(result.filter, uri))
                )
        )
    }

    /**
     * TODO - redocument
     * Correlate the given `repository` and `commit` with the given package and mark the
     * set of `uris` that are used within that pacakge.
     *
     * @param scheme The package manager scheme (e.g. npm, pip).
     * @param name The package name.
     * @param version The package version.
     * @param repository The repository that depends on the given pacakge.
     * @param commit The commit that depends on the given pacakge.
     * @param uris The set of URIs that are used by this repository and commit in the given package.
     */
    public async addReferences(repository: string, commit: string, references: SymbolReferences[]): Promise<void> {
        return await this.withTransactionalEntityManager(async entityManager => {
            const inserter = new Inserter(entityManager, ReferenceModel, 7)
            for (const reference of references) {
                // TODO - upsert

                await inserter.insert({
                    repository,
                    commit,
                    filter: await createFilter(reference.identifiers), // TODO - rename from URIs
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
