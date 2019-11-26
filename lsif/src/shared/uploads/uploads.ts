import pRetry from 'p-retry'
import * as xrepoModels from '../models/xrepo'
import { Connection, Brackets, EntityManager } from 'typeorm'
import { FORMAT_TEXT_MAP, Span, Tracer } from 'opentracing'
import { instrumentQuery, withInstrumentedTransaction } from '../database/postgres'

// TODO - rename

/**
 * A wrapper around the database tables that control uploads. This class has
 * behaviors to enqueue uploads and and dequeue them for the worker to convert
 * in a transasctional manner.
 */
export class Queue {
    /**
     * Create a new `Queue` backed by the given database connection.
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
                queryBuilder = queryBuilder.andWhere(
                    new Brackets(qb =>
                        qb
                            .where("repository LIKE '%' || :query || '%'", { query })
                            .orWhere("\"commit\" LIKE '%' || :query || '%'", { query })
                            .orWhere("root LIKE '%' || :query || '%'", { query })
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
        // TODO - check this condition
        return !!(await instrumentQuery(() =>
            this.connection.query('DELETE FROM lsif_uploads WHERE id = $1 RETURNING id', [id])
        ))
    }

    /**
     * Remove all uploads that are older than `maxAge` seconds. Returns the count of deleted uploads.
     *
     * @param maxAge The maximum age for an upload.
     */
    public async clean(maxAge: number): Promise<number> {
        return (
            (
                await instrumentQuery(() =>
                    this.connection
                        .getRepository(xrepoModels.LsifUpload)
                        .createQueryBuilder()
                        .delete()
                        .where("uploaded_at < now() - (:maxAge * interval '1 second')", { maxAge })
                        .execute()
                )
            ).affected || 0
        )
    }

    /**
     * Move all processing uploads started more than `maxAge` seconds ago that are not currently
     * locked back to the `queued` state.
     *
     * @param maxAge The maximum age for an unlocked upload in the `processing` state.
     */
    public async resetStalled(maxAge: number): Promise<number[]> {
        const results: [{ id: number }[]] = await instrumentQuery(() =>
            this.connection.query(
                `
                    UPDATE lsif_uploads u SET state = 'queued', started_at = null WHERE id = ANY(
                        SELECT id FROM lsif_uploads
                        WHERE state = 'processing' AND started_at < now() - ($1 * interval '1 second')
                        FOR UPDATE SKIP LOCKED
                    )
                    RETURNING u.id
                `,
                [maxAge]
            )
        )

        return results[0].map(r => r.id)
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
     * Lock and convert a queued upload. If the conversion function throws an error, then
     * the error summary and stack trace will be written to the upload record.
     *
     * @param convert The function to call with the locked upload.
     */
    public async dequeueAndConvert(convert: (upload: xrepoModels.LsifUpload) => Promise<void>): Promise<boolean> {
        // First, we select the next oldest upload with a state of `queued` and set
        // its state to `processing`. We do this outside of a transaction so that this
        // state transition is visible to the API. We skip any locked rows as they are
        // being handled by another worker process.
        const lockResult: [{ id: number }[]] = await this.connection.query(`
            UPDATE lsif_uploads u SET state = 'processing', started_at = now() WHERE id = (
                SELECT id FROM lsif_uploads
                WHERE state = 'queued'
                ORDER BY uploaded_at
                FOR UPDATE SKIP LOCKED LIMIT 1
            )
            RETURNING u.id
        `)
        if (lockResult[0].length === 0) {
            return false
        }
        const uploadId = lockResult[0][0].id

        return withInstrumentedTransaction(this.connection, async entityManager => {
            // TODO - do a few tests with this...
            const uploads: [
                xrepoModels.LsifUpload
            ] = await entityManager.query('SELECT * FROM lsif_uploads WHERE id = $1 FOR UPDATE LIMIT 1', [uploadId])

            let state = 'completed'
            let failureSummary: string | null = null
            let failureStacktrace: string | null = null

            try {
                await convert(uploads[0])
            } catch (error) {
                state = 'errored'
                failureSummary = error && error.message
                failureStacktrace = error && error.stack
            }

            await entityManager.query(
                `
                    UPDATE lsif_uploads
                    SET completed_or_errored_at = now(), state = $2, failure_summary = $3, failure_stacktrace = $4
                    WHERE id = $1
                `,
                [uploadId, state, failureSummary, failureStacktrace]
            )

            return true
        })
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
    public async waitForUploadToConvert(uploadId: number, maxWait: number): Promise<boolean> {
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
