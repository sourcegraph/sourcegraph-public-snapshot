import AsyncPolling from 'async-polling'
import { Connection } from 'typeorm'
import { logAndTraceCall, TracingContext } from './tracing'
import { Logger } from 'winston'
import { tryWithLock } from './store/locks'

interface Task {
    intervalMs: number
    handler: () => Promise<void>
}

/**
 * A collection of tasks that are invoked periodically, each holding an
 * exclusive advisory lock on a Postgres database connection.
 */
export class ExclusivePeriodicTaskRunner {
    private tasks: Task[] = []

    /**
     * Create a new task runner.
     *
     * @param connection The Postgres connection.
     * @param logger The logger instance.
     */
    constructor(private connection: Connection, private logger: Logger) {}

    /**
     * Register a task to be performed while holding an exclusive advisory lock in Postgres.
     *
     * @param args Parameter bag
     */
    public register({
        /** The task name. */
        name,
        /** The interval between task invocations. */
        intervalMs,
        /** The function to invoke. */
        task,
        /** Whether or not to silence logging. */
        silent = false,
    }: {
        name: string
        intervalMs: number
        task: ({ connection, ctx }: { connection: Connection; ctx: TracingContext }) => Promise<void>
        silent?: boolean
    }): void {
        const taskArgs = { connection: this.connection, ctx: {} }

        this.tasks.push({
            intervalMs,
            handler: () =>
                tryWithLock(this.connection, name, () =>
                    silent
                        ? task(taskArgs)
                        : logAndTraceCall({ logger: this.logger }, name, ctx => task({ ...taskArgs, ctx }))
                ),
        })
    }

    /** Start running all registered tasks on the specified interval. */
    public run(): void {
        for (const { intervalMs, handler } of this.tasks) {
            const fn = async (end: () => void): Promise<void> => {
                await handler()
                end()
            }

            AsyncPolling(fn, intervalMs * 1000).run()
        }
    }
}
