import * as vscode from 'vscode'

/**
 * Schedules callbacks to run after a period of idle time.
 *
 * This class listens for editor activity and schedules callbacks
 * to run after the editor has been idle for a certain amount of time.
 *
 * @param idleTime The amount of idle time in milliseconds before running callbacks.
 */
export class FixupIdleScheduler {
    private callbacks: (() => void)[] = []
    private lastActivityTime = Date.now()
    private timeoutId: NodeJS.Timeout | undefined
    private readonly timeoutLimit = 1000

    constructor(private readonly idleTime: number) {
        this.processCallback = this.processCallback.bind(this)
        // Call registerActivity when the user perform any action in the editor
        // ex. typing, scrolling, clicking, etc.
        vscode.window.onDidChangeTextEditorSelection(() => this.registerActivity())
        vscode.window.onDidChangeTextEditorVisibleRanges(() => this.registerActivity())
        vscode.window.onDidChangeTextEditorViewColumn(() => this.registerActivity())
        vscode.window.onDidChangeActiveTextEditor(() => this.registerActivity())
    }

    /**
     * Registers a callback to run after the idle time.
     *
     * @param callback The callback function to run after idle time.
     */
    public async registerCallback(callback: () => void): Promise<void> {
        this.callbacks.push(callback)
        if (this.timeoutId === undefined) {
            const timeSinceLastActivity = Date.now() - this.lastActivityTime
            if (timeSinceLastActivity >= this.idleTime) {
                await Promise.resolve().then(() => this.processCallback())
            }
            this.scheduleTimeout()
        }
    }

    /**
     * Registers editor activity and reschedules the timeout.
     *
     * This should be called whenever the user performs an action in the editor.
     * It reschedules the timeout to run after the idle time from the current time.
     */
    // Schedule the timeout onlyif there is work waiting
    private registerActivity(): void {
        this.lastActivityTime = Date.now()
        if (this.timeoutId === undefined && this.callbacks.length > 0) {
            this.scheduleTimeout()
        }
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
            setTimeout(() => {
                callback()
            }, this.timeoutLimit)
        }
    }

    public dispose(): void {
        if (this.timeoutId !== undefined) {
            clearTimeout(this.timeoutId)
        }
    }
}
