/**
 *
 * This class listens for LLM activities and schedules callbacks to run when there is no other activity in progress.
 * Runs fixup tasks when Cody is in idle state, ex. when user is not sending chat requests to Cody.
 *
 * @param idleTime The amount of idle time in milliseconds before running callbacks.
 */
export class FixupIdleScheduler {
    private callbacks: (() => void)[] = []
    private lastActivityTime = Date.now()
    private timeoutId: NodeJS.Timeout | undefined

    constructor(private readonly idleTime: number) {
        this.processCallback = this.processCallback.bind(this)
    }

    /**
     * Registers a callback to run when idle
     *
     * @param callback The callback function to run after idle time.
     */
    public async registerCallback(callback: () => void): Promise<void> {
        this.callbacks.push(callback)
        //  currently no timeout scheduled
        if (this.timeoutId === undefined) {
            const timeSinceLastActivity = Date.now() - this.lastActivityTime
            if (timeSinceLastActivity >= this.idleTime) {
                await Promise.resolve().then(() => this.processCallback())
            }
            this.scheduleTimeout()
        }
    }

    /**
     * Registers user's non-fixup interactions with Cody to reschedule the timeout.
     *
     * This should be called whenever the user interacts with Cody through execute recipes.
     * When there is no activity, we should run the fixup tasks.
     */
    public registerActivity(): void {
        this.lastActivityTime = Date.now()
        // when there is no timeout scheduled, run the callback immediately
        if (this.timeoutId === undefined) {
            this.processCallback()
        }
        // Schedule the next timeout and don't rely on the user's next interaction
        this.scheduleTimeout()
    }

    private scheduleTimeout(): void {
        // currently a timeout is scheduled so clear it
        if (this.timeoutId !== undefined) {
            clearTimeout(this.timeoutId)
        }
        const delay = Math.max(this.idleTime, 0)
        this.timeoutId = setTimeout(() => {
            this.processCallback()
        }, delay)
    }

    // if there is a timeout scheduled, it means that the user has interacted with Cody and we should not run the callback
    // when there is no timeout scheduled and have callbacks to process, process the callbacks immediately until a new timeout is scheduled
    private processCallback(): void {
        while (this.timeoutId === undefined && this.callbacks.length > 0) {
            this.runCallback()
        }
    }

    private runCallback(): void {
        const callback = this.callbacks.shift() ?? (() => {})
        return callback()
    }

    public dispose(): void {
        if (this.timeoutId !== undefined) {
            clearTimeout(this.timeoutId)
        }
    }
}
