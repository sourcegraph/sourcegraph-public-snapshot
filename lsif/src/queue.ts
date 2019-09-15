import { Queue, Worker, Job } from 'node-resque'
import { JobClasses } from './jobs'

/**
 * This type provides additional methods defined in node-resque but not
 * defined in @types/node-resque. These methods are used to control the
 * queue via the HTTP API and emit metrics.
 */
export type RealQueue = Queue & {
    /**
     * Return basic stats about the queue.
     */
    stats(): Promise<{ processed: number | undefined; failed: number | undefined }>

    /**
     * Return all queued jobs.
     *
     * @param q The queue name.
     * @param start The first index.
     * @param stop The lasts index (-1 for end of list).
     */
    queued(q: string, start: number, stop: number): Promise<JobMeta[]>

    /**
     * Return all failed jobs.
     *
     * @param start The first index.
     * @param stop The lasts index (-1 for end of list).
     */
    failed(start: number, stop: number): Promise<FailedJobMeta[]>

    /**
     * Return the current status of each worker. This returns a map from worker
     * name to the details of their current job. Defunct workers may have the
     * status 'started', which needs to be filtered out.
     */
    allWorkingOn(): Promise<{ [K: string]: WorkerMeta | 'started' }>
}

/**
 * This type updates the type of job, success, and failure callbacks of the
 * node-resque Worker class. These types are ill-defined in @types/node-resque.
 */
export type RealWorker = Worker & {
    on(event: 'job', cb: (queue: string, job: Job<any> & JobMeta) => void): Worker
    on(event: 'success', cb: (queue: string, job: Job<any> & JobMeta, result: any) => void): Worker
    on(event: 'failure', cb: (queue: string, job: Job<any> & JobMeta, failure: any) => void): Worker
}

/**
 * Metadata about a job (queued, failed, or active).
 */
export interface JobMeta {
    /**
     * The type of the job.
     */
    class: JobClasses

    /**
     * The arguments of the job.
     */
    args: any[]
}

/**
 * Metadata about a job in the failed queue.
 */
export interface FailedJobMeta {
    /**
     * The job parameters.
     */
    payload: JobMeta

    /**
     * The time at which the job failed.
     */
    failed_at: string

    /**
     * The exception message that occurred.
     */
    error: string
}

/**
 * Metadata about a worker's current state.
 */
export interface WorkerMeta {
    /**
     * The job parameters.
     */
    payload: JobMeta

    /**
     * The time at which the job was accepted by the worker.
     */
    run_at: string
}

/**
 * Rewrite a job payload to return to the uer. This rewrites the arguments
 * array from a positional list into an object with meaningful names.
 *
 * @param job The job to rewrite.
 */
export function rewriteJobMeta(job: JobMeta): any {
    const argumentTransformers = {
        convert: (args: any[]) => ({ repository: args[0], commit: args[1] }),
    }

    return {
        class: job.class,
        args: argumentTransformers[job.class](job.args),
    }
}
