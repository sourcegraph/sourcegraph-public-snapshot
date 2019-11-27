import * as xrepoModels from '../models/xrepo'
import pRetry from 'p-retry'
import { Brackets, Connection, EntityManager } from 'typeorm'
import { FORMAT_TEXT_MAP, Span, Tracer } from 'opentracing'
import { instrumentQuery, withInstrumentedTransaction } from '../database/postgres'

/**
 * A wrapper around the database tables that control uploads. This class has
 * behaviors to enqueue uploads.
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

    /**
     * Create a new uploaded with a state of `queued`.
     *
     * @param args The upload payload.
     * @param tracer The tracer instance.
     * @param span The parent span.
     */
    public async enqueue(
        {
            repository,
            commit,
            root,
            filename,
        }: {
            /** The repository. */
            repository: string
            /** The commit. */
            commit: string
            /** The root. */
            root: string
            /** The filename. */
            filename: string
        },
        tracer?: Tracer,
        span?: Span
    ): Promise<xrepoModels.LsifUpload> {
        const tracing = {}
        if (tracer && span) {
            tracer.inject(span, FORMAT_TEXT_MAP, tracing)
        }

        // TODO -need fields for tracer
        const upload = new xrepoModels.LsifUpload()
        upload.repository = repository
        upload.commit = commit
        upload.root = root
        upload.filename = filename
        await instrumentQuery(() => this.connection.createEntityManager().save(upload))

        return upload
    }

    /**
     * Wait for the given upload to be converted. The function resolves to true if the
     * conversion completed within the given timeout and false otherwise. If the upload
     * conversion throws an error, that error is thrown in-band. A NaN-valued max wait
     * will block forever.
     *
     * @param uploadId The id of the upload to block on.
     * @param maxWait The maximum time (in seconds) to wait for the promise to resolve.
     */
    public waitForUploadToConvert(uploadId: number, maxWait: number): Promise<boolean> {
        const checkUploadState = async (): Promise<boolean> => {
            const upload = await instrumentQuery(() =>
                this.connection.getRepository(xrepoModels.LsifUpload).findOneOrFail({ id: uploadId })
            )
            if (upload.state === 'errored') {
                // TODO - test this
                const error = new Error(upload.failureSummary)
                error.stack = upload.failureStacktrace
                throw error
            }

            return upload.state === 'completed'
        }

        return pRetry(
            checkUploadState,
            isNaN(maxWait)
                ? { forever: true }
                : {
                      factor: 1,
                      retries: maxWait,
                      minTimeout: 1000,
                      maxTimeout: 1000,
                  }
        )
    }
}
