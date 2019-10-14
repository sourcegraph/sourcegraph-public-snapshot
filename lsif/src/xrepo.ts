import {
    bloomFilterEventsCounter,
    xrepoInsertionDurationHistogram,
    xrepoQueryDurationHistogram,
    xrepoQueryErrorsCounter,
} from './xrepo.metrics'
import { instrument } from './metrics'
import { Connection, EntityManager } from 'typeorm'
import { createFilter, testFilter } from './encoding'
import { PackageModel, ReferenceModel, Commit, DumpID, LsifDump } from './xrepo.models'
import { TableInserter } from './inserter'
import { discoverAndUpdateCommit } from './commits'
import { TracingContext } from './tracing'

/**
 * The maximum number of commits to visit breadth-first style when when finding
 * the closest commit.
 */
export const MAX_TRAVERSAL_LIMIT = 100

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
     * @param connection The Postgres connection.
     */
    constructor(private connection: Connection) {}

    /**
     * Determine if we have commit parent data for this commit.
     *
     * @param repository The name of the repository.
     * @param commit The target commit.
     */
    public isCommitTracked(repository: string, commit: string): Promise<boolean> {
        return this.withConnection(
            async connection =>
                (await connection.getRepository(Commit).findOne({ where: { repository, commit } })) !== undefined
        )
    }

    /**
     * Return the commit that has LSIF data 'closest' to the given target commit (a direct descendant
     * or ancestor of the target commit). If no closest commit can be determined, this method returns
     * undefined.s
     *
     * @param repository The name of the repository.
     * @param commit The target commit.
     * @param ctx The tracing context.
     * @param gitserverUrls The set of ordered gitserver urls.
     */
    public async findClosestCommitWithData(
        repository: string,
        commit: string,
        ctx: TracingContext = {},
        gitserverUrls?: string[]
    ): Promise<string | undefined> {
        // Request updated commit data from gitserver if this commit isn't
        // already tracked. This will pull back ancestors for this commit
        // up to a certain (configurable) depth and insert them into the
        // cross-repository database. This populates the necessary data for
        // the following query.
        if (gitserverUrls) {
            await discoverAndUpdateCommit({ xrepoDatabase: this, repository, commit, gitserverUrls, ctx })
        }

        const query = `
            WITH RECURSIVE lineage(repository, "commit", parent_commit, has_lsif_data, direction) AS (
                -- seed result set with the target repository and commit marked by traversal direction
                SELECT l.* FROM (
                    SELECT c.*, 'A' FROM lsif_commits_with_lsif_data c WHERE c.repository = $1 AND c."commit" = $2
                    UNION
                    SELECT c.*, 'D' FROM lsif_commits_with_lsif_data c WHERE c.repository = $1 AND c."commit" = $2
                ) l

                UNION

                -- get the next commit in the ancestor or descendant direction
                SELECT * FROM (
                    WITH l_inner AS (SELECT * FROM lineage)

                    SELECT c.*, l.direction FROM l_inner l
                        JOIN lsif_commits_with_lsif_data c
                        ON l.direction = 'A' and c.repository = l.repository AND c.parent_commit = l."commit"
                    UNION
                    SELECT c.*, l.direction FROM l_inner l
                        JOIN lsif_commits_with_lsif_data c
                        ON l.direction = 'D' AND c.repository = l.repository AND c."commit" = l.parent_commit
                ) subquery
            )

            -- lineage is ordered by distance to the target commit by construction; get first commit with data
            SELECT l.commit FROM (SELECT * FROM lineage LIMIT $3) l WHERE l.has_lsif_data LIMIT 1;
        `

        return this.withConnection(async connection => {
            const results = (await connection.query(query, [repository, commit, MAX_TRAVERSAL_LIMIT])) as {
                commit: string
            }[]

            if (results.length === 0 || !results[0].commit) {
                return undefined
            }

            return results[0].commit
        })
    }

    /**
     * Update the known commits for a repository. The given commits are pairs of revhashes, where
     * [`a`, `b`] indicates that one parent of `b` is `a`. If a commit has no parents, then there
     * should be a pair of the form [`a`, ``] (where the parent is an empty string).
     *
     * @param repository The name of the repository.
     * @param commits The commit parentage data.
     */
    public async updateCommits(repository: string, commits: [string, string][]): Promise<void> {
        return await this.withTransactionalEntityManager(async entityManager => {
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
     * @param commit The commit of the being updated.
     * @param packages The list of packages that this repository defines (scheme, name, and version).
     * @param references The list of packages that this repository depends on (scheme, name, and version) and the symbols that the package references.
     */
    public addPackagesAndReferences(
        repository: string,
        commit: string,
        packages: Package[],
        references: SymbolReferences[]
    ): Promise<DumpID> {
        return this.withTransactionalEntityManager(async entityManager => {
            const dumpID = await this.insertDump(repository, commit, entityManager)

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
                await packageInserter.insert({ dump_id: dumpID, ...pkg })
            }

            for (const reference of references) {
                await referenceInserter.insert({
                    dump_id: dumpID,
                    filter: await createFilter(reference.identifiers),
                    ...reference.package,
                })
            }

            await packageInserter.flush()
            await referenceInserter.flush()

            return dumpID
        })
    }

    /**
     * Inserts the given repository and commit into the `lsif_dumps` table.
     *
     * @param repository The repository.
     * @param commit The commit.
     * @param entityManager The EntityManager for the connection to the xrepo database.
     */
    public async insertDump(
        repository: string,
        commit: string,
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<DumpID> {
        // Remove all previous package data for this repo/commit (this
        // cascades to packages and references)
        await entityManager
            .createQueryBuilder()
            .delete()
            .from(LsifDump)
            .where({ repository, commit })
            .execute()

        // Mark that we have data available for this commit
        const result = await entityManager
            .createQueryBuilder()
            .insert()
            .onConflict('DO NOTHING')
            .into(LsifDump)
            .values({ repository, commit })
            .execute()

        if (result.identifiers.length === 0) {
            throw new Error(`Unable to insert row into lsif_dumps table for repository ${repository} commit ${commit}.`)
        }

        return result.identifiers[0].id
    }

    /**
     * Find the dump for the given repository and commit.
     *
     * @param repository The repository.
     * @param commit The commit.
     * @param entityManager The EntityManager for the connection to the xrepo database.
     */
    public async getDump(repository: string, commit: string): Promise<LsifDump | undefined> {
        return await this.withConnection(connection =>
            connection.getRepository(LsifDump).findOne({ where: { repository, commit } })
        )
    }

    /**
     * Find all repository/commit pairs that reference `value` in the given package. The
     * returned results will include only repositories that have a dependency on the given
     * package. The returned results may (but is not likely to) include a repository/commit
     * pair that does not reference `value`. See cache.ts for configuration values that tune
     * the bloom filter false positive rates.
     *
     * @param args Parameter bag.
     */
    public async getReferences({
        scheme,
        name,
        version,
        value,
    }: {
        /** The package manager scheme (e.g. npm, pip).  */
        scheme: string
        /** The package name.  */
        name: string
        /** The package version.  */
        version: string | null
        /** The value to test.  */
        value: string
    }): Promise<ReferenceModel[]> {
        const results = await this.withConnection(connection =>
            connection.getRepository(ReferenceModel).find({ where: { scheme, name, version } })
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
}
