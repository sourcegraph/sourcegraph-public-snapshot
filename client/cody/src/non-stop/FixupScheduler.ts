import { FixupIdleTaskRunner } from './roles'

/**
 * Runs callbacks "later".
 */
export class FixupScheduler implements FixupIdleTaskRunner {
    private work_: (() => void)[] = []
    private timeout_: NodeJS.Timeout
    private scheduled_ = false

    constructor(delayMsec: number) {
        this.timeout_ = setTimeout(this.doWorkNow.bind(this), delayMsec).unref()
    }

    // TODO: Consider making this disposable and not running tasks after
    // being disposed

    // TODO: Add a callback so the scheduler knows when the user is typing
    // and add a cooldown period

    /**
     * Schedules a callback which will run when the event loop is idle.
     * @param callback the callback to run.
     */
    public scheduleIdle<T>(worker: () => T): Promise<T> {
        if (!this.work_.length) {
            // First work item, so schedule the window callback
            this.scheduleCallback()
        }
        return new Promise((resolve, reject) => {
            this.work_.push(() => {
                try {
                    resolve(worker())
                } catch (error: any) {
                    reject(error)
                }
            })
        })
    }

    private scheduleCallback(): void {
        if (!this.scheduled_) {
            this.scheduled_ = true
            this.timeout_.refresh()
        }
    }

    public doWorkNow(): void {
        this.scheduled_ = false
        const item = this.work_.shift()
        if (!item) {
            return
        }
        if (this.work_.length) {
            this.scheduleCallback()
        }
        item()
    }
}
