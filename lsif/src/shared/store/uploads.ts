import * as pgModels from '../models/pg'
import { Brackets, Connection, EntityManager } from 'typeorm'
import { FORMAT_TEXT_MAP, Span, Tracer } from 'opentracing'
import { instrumentQuery, withInstrumentedTransaction } from '../database/postgres'
import { PlainObjectToDatabaseEntityTransformer } from 'typeorm/query-builder/transformer/PlainObjectToDatabaseEntityTransformer'
import { Logger } from 'winston'

export interface LsifUploadWithPlaceInQueue extends pgModels.LsifUpload {
    placeInQueue: number | null
}

/**
 * A wrapper around the database tables that control uploads. This class has
 * behaviors to enqueue uploads and dequeue them for the processor to convert
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
     * @param repositoryId The repository identifier.
     * @param state The state.
     * @param query A search query.
     * @param visibleAtTip If true, only return dumps visible at tip.
     * @param limit The maximum number of uploads to return.
     * @param offset The number of uploads to skip.
     */
    public async getUploads(
        repositoryId: number,
        state: pgModels.LsifUploadState | undefined,
        query: string,
        visibleAtTip: boolean,
        limit: number,
        offset: number
    ): Promise<{ uploads: LsifUploadWithPlaceInQueue[]; totalCount: number }> {
        const { uploads, raw, totalCount } = await instrumentQuery<{
            uploads: pgModels.LsifUpload[]
            raw: { upload_id: number; rank: string | undefined }[]
            totalCount: number
        }>(async () => {
            let queryBuilder = this.connection
                .getRepository(pgModels.LsifUpload)
                .createQueryBuilder('upload')
                .addSelect('ranked.rank', 'rank')
                .leftJoin(
                    qb =>
                        qb
                            .subQuery()
                            .select('ranked.id, RANK() OVER (ORDER BY ranked.uploaded_at) as rank')
                            .from(pgModels.LsifUpload, 'ranked')
                            .where("ranked.state = 'queued'"),
                    'ranked',
                    'ranked.id = upload.id'
                )
                .where({ repositoryId })
                .orderBy('uploaded_at', 'DESC')
                .limit(limit)
                .offset(offset)

            if (state) {
                queryBuilder = queryBuilder.andWhere('state = :state', { state })
            }

            if (query) {
                const clauses = ['commit', 'root', 'indexerName', 'failure_summary', 'failure_stacktrace'].map(
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

            const [{ entities, raw: rawEntities }, count] = await Promise.all([
                queryBuilder.getRawAndEntities(),
                queryBuilder.getCount(),
            ])

            return { uploads: entities, raw: rawEntities, totalCount: count }
        })

        const ranks = new Map(raw.map(r => [r.upload_id, parseInt(r.rank || '', 10)]))
        return { uploads: uploads.map(u => ({ ...u, placeInQueue: ranks.get(u.id) || null })), totalCount }
    }

    /**
     * Get an upload by identifier.
     *
     * @param id The upload identifier.
     */
    public getUpload(id: number): Promise<LsifUploadWithPlaceInQueue | undefined> {
        return withInstrumentedTransaction(this.connection, async () => {
            const {
                entities,
                raw,
            }: { entities: pgModels.LsifUpload[]; raw: { rank: string | null }[] } = await this.connection
                .getRepository(pgModels.LsifUpload)
                .createQueryBuilder('upload')
                .addSelect('ranked.rank', 'rank')
                .leftJoin(
                    qb =>
                        qb
                            .subQuery()
                            .select('ranked.id, RANK() OVER (ORDER BY ranked.uploaded_at) as rank')
                            .from(pgModels.LsifUpload, 'ranked')
                            .where("ranked.state = 'queued'"),
                    'ranked',
                    'ranked.id = upload.id'
                )
                .where({ id })
                .limit(1)
                .getRawAndEntities()

            if (entities.length === 0) {
                return undefined
            }

            return { ...entities[0], placeInQueue: parseInt(raw[0].rank || '', 10) || null }
        })
    }

    /**
     * Delete an upload. This returns true if the upload existed. Also remove referenced
     * package and reference rows if the upload was successfully processed.
     *
     * Does not delete the file on disk directly. This will be cleaned up later as part
     * of the `removeDeadDumps` function performed at the start of the `purgeOldDumps`
     * task that is run on a schedule in the server context.
     *
     * @param id The upload identifier.
     * @param updateVisibility A function that updates the dumps visible at the tip for
     *     the given repository. This is called if the deleted dump was visible at tip,
     *     as a previously non-visible dump may become visible after deletion.
     */
    public async deleteUpload(
        id: number,
        updateVisibility: (entityManager: EntityManager, repositoryId: number) => Promise<void>
    ): Promise<boolean> {
        return withInstrumentedTransaction(this.connection, async entityManager => {
            const [affected, numAffected]: [
                { repository_id: number; visible_at_tip: boolean }[],
                number
            ] = await instrumentQuery(() =>
                entityManager.query('DELETE FROM lsif_uploads WHERE id = $1 RETURNING repository_id, visible_at_tip', [
                    id,
                ])
            )

            if (numAffected === 0) {
                return false
            }

            if (affected[0].visible_at_tip) {
                await updateVisibility(entityManager, affected[0].repository_id)
            }

            return true
        })
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
            repositoryId,
            commit,
            root,
            filename,
            indexer,
        }: {
            /** The repository identifier. */
            repositoryId: number
            /** The commit. */
            commit: string
            /** The root. */
            root: string
            /** The filename. */
            filename: string
            /** The indexer binary name that produced this dump as specified by the metadata. */
            indexer: string
        },
        tracer?: Tracer,
        span?: Span
    ): Promise<pgModels.LsifUpload> {
        const tracing = {}
        if (tracer && span) {
            tracer.inject(span, FORMAT_TEXT_MAP, tracing)
        }

        const upload = new pgModels.LsifUpload()
        upload.repositoryId = repositoryId
        upload.commit = commit
        upload.root = root
        upload.indexer = indexer
        upload.filename = filename
        upload.tracingContext = JSON.stringify(tracing)
        await instrumentQuery(() => this.connection.createEntityManager().save(upload))

        return upload
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
        // being handled by another dump processor.
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
