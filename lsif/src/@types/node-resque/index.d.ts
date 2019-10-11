// Type definitions for node-resque 5.5.7
// Project: http://github.com/taskrabbit/node-resque
// Definitions by: Gordey Doronin <https://github.com/gordey4doronin>
// Modified by: Eric Fritz <eric@sourcegraph.com>

declare module 'node-resque' {
    // We use ioredis in this project, so we can import it
    // here to import the types of the connection. This is
    // necessary if we ever use plugins, as they will need
    // to access the underlying Redis instance to manipulate
    // the proper data.

    import { Redis } from 'ioredis'

    export interface ConnectionOptions {
        pkg?: string
        host?: string
        port?: number
        database?: number
        namespace?: string
        looping?: boolean
        options?: any
    }

    export class Connection extends NodeJS.EventEmitter {
        redis?: Redis
        constructor(options: ConnectionOptions)

        connect(): Promise<void>
        end(): Promise<void>
        key(...args: string[]): string

        on(event: 'error', cb: (error: Error) => void): this
        once(event: 'error', cb: (error: Error) => void): this
    }

    export interface JobDefinition<T> {
        plugins?: string[]
        pluginOptions?: { [pluginName: string]: any }
        perform: (...args: any[]) => Promise<T>
    }

    export interface JobsHash {
        [jobName: string]: JobDefinition<any>
    }

    export interface Job {
        class: string
        args: any[]
    }

    export interface WorkerPayload {
        worker: string
        queue: string
        payload: Job
        run_at: string
    }

    export interface ErrorPayload {
        worker: string
        queue: string
        payload: Job
        exception: string
        error: string
        backtrace: string[]
        failed_at: string
    }

    export interface Stats {
        processed: number | undefined
        failed: number | undefined
    }

    export interface QueueOptions {
        connection?: ConnectionOptions
    }

    export type QueueEvent = 'error'

    export class Queue extends NodeJS.EventEmitter {
        connection: Connection
        constructor(options: QueueOptions, jobs?: JobsHash)

        connect(): Promise<void>
        enqueue(queue: string, jobName: string, args: Readonlyany[]): Promise<void>
        enqueueAt(timestampMs: number, queue: string, jobName: string, args: Readonlyany[]): Promise<void>
        enqueueIn(milliseconds: number, queue: string, jobName: string, args: Readonlyany[]): Promise<void>
        queues(): Promise<string[]>
        delQueue(queue: string): Promise<void>
        length(queue: string): Promise<number>
        del(queue: string, jobName: string, args: Readonlyany[], count: number): Promise<void>
        delDelayed(queue: string, jobName: string, args: Readonlyany[]): Promise<number[]>
        scheduledAt(queue: string, jobName: string, args: Readonlyany[]): Promise<number[]>
        queued(queue: string, start: number, stop: number): Promise<Job[]>
        allDelayed(): Promise<{ [K: number]: Job[] }>
        delLock(key: string): Promise<void>
        workers(): Promise<{ [K: string]: string }>
        workingOn(workerName: string, queues: string[]): Promise<WorkerPayload>
        allWorkingOn(): Promise<{ [K: string]: WorkerPayload | 'started' }>
        forceCleanWorker(workerName: string): Promise<ErrorPayload>
        cleanOldWorkers(age): Promise<{ [K: string]: ErrorPayload }>
        failedCount(): Promise<number>
        failed(start: number, stop: number): Promise<ErrorPayload[]>
        removeFailed(failedJob: ErrorPayload): Promise<void>
        retryAndRemoveFailed(failedJob: ErrorPayload): Promise<void>
        stats(): Promise<Stats>
        end(): Promise<void>

        on(event: 'error', cb: (error: Error, queue: string) => void): this
        removeAllListeners(event: QueueEvent): this
    }

    export interface WorkerOptions {
        connection?: ConnectionOptions
        queues: string[]
        name?: string
        timeout?: number
        looping?: boolean
    }

    export type WorkerEvent =
        | 'start'
        | 'end'
        | 'cleaning_worker'
        | 'poll'
        | 'ping'
        | 'job'
        | 'reEnqueue'
        | 'success'
        | 'failure'
        | 'error'
        | 'pause'

    export class Worker extends NodeJS.EventEmitter {
        connection: Connection
        constructor(options: WorkerOptions, jobs?: JobsHash)

        connect(): Promise<void>
        start(): Promise<void>
        init(): Promise<void>
        end(): Promise<void>

        // This rule wants to incorrectly combine events with distinct callback types.
        /* eslint-disable @typescript-eslint/unified-signatures */
        on(event: 'start' | 'end' | 'pause', cb: () => void): this
        on(event: 'cleaning_worker', cb: (worker: string, pid: string) => void): this
        on(event: 'poll', cb: (queue: string) => void): this
        on(event: 'ping', cb: (time: number) => void): this
        on(event: 'job', cb: (queue: string, job: Job) => void): this
        on(event: 'reEnqueue', cb: (queue: string, job: Job, plugin: string) => void): this
        on(event: 'success', cb: (queue: string, job: Job, result: any) => void): this
        on(event: 'failure', cb: (queue: string, job: Job, failure: any) => void): this
        on(event: 'error', cb: (error: Error, queue: string, job: Job) => void): this
        removeAllListeners(event: WorkerEvent): this
    }

    export interface SchedulerOptions {
        connection?: ConnectionOptions
        name?: string
        timeout?: number
        stuckWorkerTimeout?: number
        masterLockTimeout?: number
    }

    export type SchedulerEvent =
        | 'start'
        | 'end'
        | 'poll'
        | 'master'
        | 'cleanStuckWorker'
        | 'error'
        | 'workingTimestamp'
        | 'transferredJob'

    export class Scheduler extends NodeJS.EventEmitter {
        connection: Connection
        constructor(options: SchedulerOptions, jobs?: JobsHash)

        connect(): Promise<void>
        start(): Promise<void>
        end(): Promise<void>

        // This rule wants to incorrectly combine events with distinct callback types.
        /* eslint-disable @typescript-eslint/unified-signatures */
        on(event: 'start' | 'end' | 'poll' | 'master', cb: () => void): this
        on(event: 'cleanStuckWorker', cb: (workerName: string, errorPayload: ErrorPayload, delta: number) => void): this
        on(event: 'error', cb: (error: Error, queue: string) => void): this
        on(event: 'workingTimestamp', cb: (timestampMs: number) => void): this
        on(event: 'transferredJob', cb: (timestampMs: number, job: Job) => void): this
        removeAllListeners(event: SchedulerEvent): this
    }
}
