import AsyncPolling from 'async-polling'
import { logAndTraceCall, TracingContext } from './tracing'
import { tryWithLock } from './store/locks'
import { Connection } from 'typeorm'
import { Logger } from 'winston'

interface Task {
    intervalMs: number
    task: () => Promise<void>
}

export class TaskRunner {
    private tasks: Task[] = []

    constructor(private connection: Connection, private logger: Logger) {}

    public register(
        name: string,
        intervalMs: number,
        task: (ctx: TracingContext) => Promise<void>,
        silent = false
    ): void {
        this.tasks.push({ intervalMs, task: this.wrapTask(name, task, silent) })
    }

    public run(): void {
        for (const { intervalMs, task } of this.tasks) {
            this.runTask(task, intervalMs)
        }
    }

    /**
     * Each task is performed with an exclusive advisory lock in Postgres. If another
     * server is already running this task, then this server instance will skip the
     * attempt.
     *
     * @param name The task name. Used for logging the span and generating the lock id.
     * @param task The task function.
     */
    private wrapTask(name: string, task: (ctx: TracingContext) => Promise<void>, silent = false): () => Promise<void> {
        return () =>
            tryWithLock(this.connection, name, () =>
                silent ? task({}) : logAndTraceCall({ logger: this.logger }, name, task)
            )
    }

    /**
     * Start invoking the given task on an interval.
     *
     * @param task The task to invoke.
     * @param intervalMs The interval between invocations.
     */
    private runTask(task: () => Promise<void>, intervalMs: number): void {
        const f = async (end: () => void): Promise<void> => {
            await task()
            end()
        }

        AsyncPolling(f, intervalMs * 1000).run()
    }
}
