import * as xrepoModels from '../models/xrepo'
import { Brackets, Connection, EntityManager } from 'typeorm'
import { instrumentQuery, withInstrumentedTransaction } from '../database/postgres'

/**
 * A wrapper around the database tables that control uploads.
 */
export class UploadsManager {
    /**
     * Create a new `UploadsManager` backed by the given database connection.
     *
     * @param connection The Postgres connection.
     */
    constructor(private connection: Connection) {}

    /**
     * Get the counts of uploads in each possible state.
     */
    public getCounts(): Promise<{
        queuedCount: number
        completedCount: number
        erroredCount: number
        processingCount: number
    }> {
        return withInstrumentedTransaction(this.connection, async entityManager => ({
            queuedCount: await this.getCount('queued', entityManager),
            completedCount: await this.getCount('completed', entityManager),
            erroredCount: await this.getCount('errored', entityManager),
            processingCount: await this.getCount('processing', entityManager),
        }))
    }

    /**
     * Get the count of uploads in the given state.
     *
     * @param state The state.
     * @param entityManager An entity manager to use if within a transaction.
     */
    public getCount(
        state: string,
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<number> {
        return entityManager
            .getRepository(xrepoModels.LsifUpload)
            .createQueryBuilder()
            .select()
            .where({ state })
            .getCount()
    }

    /**
     * Get the uploads in the given state.
     *
     * @param state The state.
     * @param query A search query.
     * @param limit The maximum number of uploads to return.
     * @param offset The number of uploads to skip.
     */
    public async getUploads(
        state: xrepoModels.LsifUploadState,
        query: string,
        limit: number,
        offset: number
    ): Promise<{ uploads: xrepoModels.LsifUpload[]; totalCount: number }> {
        const [uploads, totalCount] = await instrumentQuery(() => {
            let queryBuilder = this.connection
                .getRepository(xrepoModels.LsifUpload)
                .createQueryBuilder('upload')
                .where({ state })
                .orderBy('uploaded_at', 'DESC')
                .limit(limit)
                .offset(offset)

            if (query) {
                const clauses = ['repository', 'commit', 'root', 'failure_summary', 'failure_stacktrace'].map(
                    field => `"${field}" LIKE '%' || :query || '%'`
                )

                queryBuilder = queryBuilder.andWhere(
                    new Brackets(qb =>
                        clauses.slice(1).reduce((ob, c) => ob.orWhere(c, { query }), qb.where(clauses[0], { query }))
                    )
                )
            }

            return queryBuilder.getManyAndCount()
        })

        return { uploads, totalCount }
    }

    /**
     * Get a upload by identifier.
     *
     * @param id The upload identifier.
     */
    public getUpload(id: number): Promise<xrepoModels.LsifUpload | undefined> {
        return instrumentQuery(() => this.connection.getRepository(xrepoModels.LsifUpload).findOne({ id }))
    }

    /**
     * Delete an upload. This returns true if the upload existed.
     *
     * @param id The upload identifier.
     */
    public async deleteUpload(id: number): Promise<boolean> {
        const results: [{ id: number }[]] = await instrumentQuery(() =>
            this.connection.query('DELETE FROM lsif_uploads WHERE id = $1 RETURNING id', [id])
        )

        return results[0].length > 0
    }
}
