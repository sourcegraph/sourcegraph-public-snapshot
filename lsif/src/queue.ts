import { Queue as ResqueQueue, Worker as ResqueWorker, Job } from 'node-resque'

/**
 * The names of jobs performed by the LSIF worker.
 */
export type JobClass = 'convert'

/**
 * This type provides additional methods defined in node-resque but not
 * defined in @types/node-resque. These methods are used to control the
 * queue via the HTTP API and emit metrics. Additionally, we ensure that
 * the enqueue method supplies only a job class that is defined above.
 */
export interface Queue extends Omit<ResqueQueue, 'enqueue'> {
    /**
     * Enqueue a job for a worker.
     *
     * @param queue The name of the queue.
     * @param jobName The name of the job class.
     * @param args The positional arguments supplied to the worker.
     */
    enqueue(queue: string, jobName: JobClass, args: any[]): Promise<void>

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
export interface Worker extends Omit<ResqueWorker, 'on'> {
    // This rule wants to incorrectly combine events with distinct callback types.
    /* eslint-disable @typescript-eslint/unified-signatures */
    on(event: 'start' | 'end' | 'pause', cb: () => void): this
    on(event: 'cleaning_worker', cb: (worker: string, pid: string) => void): this
    on(event: 'poll', cb: (queue: string) => void): this
    on(event: 'ping', cb: (time: number) => void): this
    on(event: 'job', cb: (queue: string, job: Job<any> & JobMeta) => void): this
    on(event: 'reEnqueue', cb: (queue: string, job: Job<any> & JobMeta, plugin: string) => void): this
    on(event: 'success', cb: (queue: string, job: Job<any> & JobMeta, result: any) => void): this
    on(event: 'failure', cb: (queue: string, job: Job<any> & JobMeta, failure: any) => void): this
    on(event: 'error', cb: (error: Error, queue: string, job: Job<any> & JobMeta) => void): this
}

/**
 * Metadata about a job (queued, failed, or active).
 */
export interface JobMeta {
    /**
     * The type of the job.
     */
    class: JobClass

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
 * Transformers that convert positional arguments for a job into an object
 * with named properties.
 */
const argumentTransformers: { [K in JobClass]: (args: any[]) => { [K: string]: any } } = {
    convert: (args: any[]) => ({ repository: args[0], commit: args[1] }),
}

/**
 * Rewrite a job payload to return to the uer. This rewrites the arguments
 * array from a positional list into an object with meaningful names.
 *
 * @param job The job to rewrite.
 */
export function rewriteJobMeta(job: JobMeta): any {
    return {
        class: job.class,
        args: argumentTransformers[job.class](job.args),
    }
}
