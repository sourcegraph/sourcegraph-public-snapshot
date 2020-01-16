import * as pgModels from '../models/pg'
import pRetry, { AbortError } from 'p-retry'
import { Brackets, Connection, EntityManager } from 'typeorm'
import { FORMAT_TEXT_MAP, Span, Tracer } from 'opentracing'
import { instrumentQuery, withInstrumentedTransaction } from '../database/postgres'
import { PlainObjectToDatabaseEntityTransformer } from 'typeorm/query-builder/transformer/PlainObjectToDatabaseEntityTransformer'
import { Logger } from 'winston'

/**
 * A wrapper around the database tables that control uploads. This class has
 * behaviors to enqueue uploads and dequeue them for the worker to convert
 * in a transactional manner.
 */
export class UploadManager {
    /**
     * Create a new `UploadManager` backed by the given database connection.
     *
     * @param connection The Postgres connection.
     */
    constructor(private connection: Connection) {}

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
            .getRepository(pgModels.LsifUpload)
            .createQueryBuilder()
            .select()
            .where({ state })
            .getCount()
    }

    /**
     * Get the uploads in the given state.
     *
     * @param repository The repository.
     * @param state The state.
     * @param query A search query.
     * @param visibleAtTip If true, only return dumps visible at tip.
     * @param limit The maximum number of uploads to return.
     * @param offset The number of uploads to skip.
     */
    public async getUploads(
        repository: string,
        state: pgModels.LsifUploadState | undefined,
        query: string,
        visibleAtTip: boolean,
        limit: number,
        offset: number
    ): Promise<{ uploads: pgModels.LsifUpload[]; totalCount: number }> {
        const [uploads, totalCount] = await instrumentQuery(() => {
            let queryBuilder = this.connection
                .getRepository(pgModels.LsifUpload)
                .createQueryBuilder('upload')
                .where({ repository })
                .orderBy('uploaded_at', 'DESC')
                .limit(limit)
                .offset(offset)

            if (state) {
                queryBuilder = queryBuilder.andWhere('state = :state', { state })
            }

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

            if (visibleAtTip) {
                queryBuilder = queryBuilder.andWhere('visible_at_tip = true')
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
    public getUpload(id: number): Promise<pgModels.LsifUpload | undefined> {
        return instrumentQuery(() => this.connection.getRepository(pgModels.LsifUpload).findOne({ id }))
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
     * Remove all uploads that are older than `maxAge` seconds. Returns the count of deleted uploads.
     *
     * @param maxAge The maximum age for an upload.
     */
    public async clean(maxAge: number): Promise<number> {
        return (
            (
                await instrumentQuery(() =>
                    this.connection
                        .getRepository(pgModels.LsifUpload)
                        .createQueryBuilder()
                        .delete()
                        .where("state != 'completed'")
                        .andWhere("uploaded_at < now() - (:maxAge * interval '1 second')", { maxAge })
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
    ): Promise<pgModels.LsifUpload> {
        const tracing = {}
        if (tracer && span) {
            tracer.inject(span, FORMAT_TEXT_MAP, tracing)
        }

        const upload = new pgModels.LsifUpload()
        upload.repository = repository
        upload.commit = commit
        upload.root = root
        upload.filename = filename
        upload.tracingContext = JSON.stringify(tracing)
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
    public async waitForUploadToConvert(uploadId: number, maxWait: number | undefined): Promise<boolean> {
        const UPLOADINPROGRESS = 'UploadInProgressError'
        class UploadInProgressError extends Error {
            public readonly name = UPLOADINPROGRESS
            public readonly code = UPLOADINPROGRESS
            constructor() {
                super('upload in progress')
            }
        }

        const checkUploadState = async (): Promise<void> => {
            const upload = await instrumentQuery(() =>
                this.connection.getRepository(pgModels.LsifUpload).findOneOrFail({ id: uploadId })
            )

            if (upload.state === 'errored') {
                const error = new Error(upload.failureSummary || '')
                error.stack = upload.failureStacktrace || ''
                throw new AbortError(error)
            }

            if (upload.state !== 'completed') {
                throw new UploadInProgressError()
            }
        }

        const retryConfig =
            maxWait === undefined
                ? { forever: true }
                : {
                      factor: 1,
                      retries: maxWait,
                      minTimeout: 1000,
                      maxTimeout: 1000,
                  }

        try {
            await pRetry(checkUploadState, retryConfig)
        } catch (error) {
            if (error && error.code === UPLOADINPROGRESS) {
                return false
            }

            throw error
        }

        return true
    }

    /**
     * Lock and convert a queued upload. If the conversion function throws an error, then
     * the error summary and stack trace will be written to the upload record and the state
     * will be set to "errored".
     *
     * The convert callback is invoked with the locked upload record and the entity manager
     * that locked the record. The callback should use it to operate in the same transaction.
     *
     * This method does NOT mark the upload as complete and the callback MUST be sure to call
     * the `markComplete` method on successful processing. Otherwise the record will block the
     * head of the queue by being re-processed ad nauseam.
     *
     * @param convert The function to call with the locked upload.
     * @param logger The logger instance.
     */
    public async dequeueAndConvert(
        convert: (upload: pgModels.LsifUpload, entityManager: EntityManager) => Promise<void>,
        logger: Logger
    ): Promise<boolean> {
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
            const results: object[] = await entityManager.query(
                'SELECT * FROM lsif_uploads WHERE id = $1 FOR UPDATE LIMIT 1',
                [uploadId]
            )
            if (results.length === 0) {
                // Record was deleted in race, retry
                return this.dequeueAndConvert(convert, logger)
            }

            // Transform locked result into upload entity
            const repo = entityManager.getRepository(pgModels.LsifUpload)
            const meta = repo.manager.connection.getMetadata(pgModels.LsifUpload)
            const transformer = new PlainObjectToDatabaseEntityTransformer(repo.manager)
            const upload = (await transformer.transform(results[0], meta)) as pgModels.LsifUpload

            try {
                await convert(upload, entityManager)
            } catch (error) {
                logger.error('Failed to convert upload', { error })

                await entityManager.query(
                    `
                        UPDATE lsif_uploads
                        SET state = 'errored', finished_at = now(), failure_summary = $2, failure_stacktrace = $3
                        WHERE id = $1
                    `,
                    [uploadId, error?.message, error?.stack]
                )
            }

            return true
        })
    }

    /**
     * Mark an upload as complete and set its finished timestamp.
     *
     * @param upload The upload.
     * @param entityManager The EntityManager to use as part of a transaction.
     */
    public markComplete(
        upload: pgModels.LsifUpload,
        entityManager: EntityManager = this.connection.createEntityManager()
    ): Promise<void> {
        return entityManager.query("UPDATE lsif_uploads SET state = 'completed', finished_at = now() WHERE id = $1", [
            upload.id,
        ])
    }
}
