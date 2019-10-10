declare module 'node-resque'

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
    redis: Redis
    constructor(options: ConnectionOptions)
    connect(): Promise<void>
    end(): Promise<void>
    key(...args: string[]): string
}

export interface Job<T> {
    plugins?: (() => any)[]
    pluginOptions?: { [K: string]: any }
    perform: (...args: any[]) => Promise<T>
}

export interface JobMeta {
    class: string
    args: any[]
}

export interface FailedJobMeta {
    payload: JobMeta
    failed_at: string
    error: string
}

export interface WorkerMeta {
    payload: JobMeta
    run_at: string
}

export interface QueueOptions {
    connection?: ConnectionOptions
}

export class Queue extends NodeJS.EventEmitter {
    connection: Connection
    constructor(options: QueueOptions, jobs?: JobsHash)
    connect(): Promise<void>
    end(): Promise<void>
    enqueue(queue: string, jobName: string, args: any[]): Promise<void>
    enqueueIn(milliseconds: number, queue: string, jobName: string, args: any[]): Promise<void>
    allWorkingOn(): Promise<{ [K: string]: WorkerMeta | 'started' }>
    queued(q: string, start: number, stop: number): Promise<JobMeta[]>
    failed(start: number, stop: number): Promise<FailedJobMeta[]>
    stats(): Promise<{ processed: number | undefined; failed: number | undefined }>
    length(q: string): Promise<number>
    failedCount(): Promise<number>
    on(event: 'error', cb: (error: Error, queue: string) => void): this
}

export interface SchedulerOptions {
    connection?: ConnectionOptions
    name?: string
    timeout?: number
    stuckWorkerTimeout?: number
    masterLockTimeout?: number
}

export class Scheduler extends NodeJS.EventEmitter {
    constructor(options: SchedulerOptions, jobs?: JobsHash)

    connect(): Promise<void>
    start(): Promise<void>
    end(): Promise<void>

    // This rule wants to incorrectly combine events with distinct callback types.
    /* eslint-disable @typescript-eslint/unified-signatures */
    on(event: 'start' | 'end' | 'poll' | 'master', cb: () => void): this
    on(event: 'cleanStuckWorker', cb: (workerName: string, errorPayload: ErrorPayload, delta: number) => void): this
    on(event: 'error', cb: (error: Error, queue: string) => void): this
    on(event: 'workingTimestamp', cb: (timestamp: number) => void): this
    on(event: 'transferredJob', cb: (timestamp: number, job: Job<any>) => void): this
}

export interface ErrorPayload {
    worker: string
    queue: string
    payload: any
    exception: string
    error: string
    backtrace: string[]
    failed_at: string
}

export interface WorkerOptions {
    connection?: ConnectionOptions
    queues: string[]
    name?: string
    timeout?: number
    looping?: boolean
}

export class Worker extends NodeJS.EventEmitter {
    connection: Connection
    constructor(options: WorkerOptions, jobs?: JobsHash)
    connect(): Promise<void>
    start(): Promise<void>
    end(): Promise<void>

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

export interface JobsHash {
    [jobName: string]: Job<any>
}
