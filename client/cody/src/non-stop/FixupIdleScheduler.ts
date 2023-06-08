class FixupIdleScheduler {
    private callbacks: (() => void)[] = []
    private lastActivityTime = Date.now()
    private timeoutId: NodeJS.Timeout | undefined

    constructor(private readonly idleTime: number) {
        this.processCallback = this.processCallback.bind(this)
    }

    private scheduleTimeout(): void {
        if (this.timeoutId !== undefined) {
            clearTimeout(this.timeoutId)
        }
        const timeSinceLastActivity = Date.now() - this.lastActivityTime
        const delay = Math.max(this.idleTime - timeSinceLastActivity, 0)
        this.timeoutId = setTimeout(() => {
            this.processCallback()
        }, delay)
    }

    private processCallback(): void {
        this.lastActivityTime = Date.now()
        const pendingCallbacks = this.callbacks.slice()
        this.callbacks = []
        for (const callback of pendingCallbacks) {
            callback()
        }
    }

    public registerCallback(callback: () => void): void {
        this.callbacks.push(callback)
        if (this.timeoutId === undefined) {
            const timeSinceLastActivity = Date.now() - this.lastActivityTime
            if (timeSinceLastActivity >= this.idleTime) {
                this.processCallback()
                return
            }
            this.scheduleTimeout()
        }
    }

    public registerActivity(): void {
        this.lastActivityTime = Date.now()
        if (this.timeoutId === undefined) {
            this.scheduleTimeout()
        }
    }

    public dispose(): void {
        if (this.timeoutId !== undefined) {
            clearTimeout(this.timeoutId)
        }
    }
}
