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
     * @param name The task name.
     * @param intervalMs The interval between task invocations.
     * @param task The function to invoke.
     */
    public register(
        name: string,
        intervalMs: number,
        task: ({ connection, ctx }: { connection: Connection; ctx: TracingContext }) => Promise<void>
    ): void {
        this.tasks.push({
            intervalMs,
            handler: () =>
                tryWithLock(this.connection, name, () =>
                    logAndTraceCall({ logger: this.logger }, name, ctx => task({ connection: this.connection, ctx }))
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
