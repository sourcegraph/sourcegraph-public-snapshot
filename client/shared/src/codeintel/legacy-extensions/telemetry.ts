import * as sourcegraph from '@sourcegraph/extension-api-types'

/**
 * A wrapper around telemetry events. A new instance of this class
 * should be instantiated at the start of each action as it handles
 * latency tracking.
 */
export class TelemetryEmitter {
    private languageID: string
    private repoID: number
    private started: number
    private enabled: boolean
    private emitted = new Set<string>()

    /**
     * Creates a new telemetry emitter object for a given
     * language ID and repository ID.
     * Emitting is enabled by default
     *
     * @param languageID The language identifier e.g. 'java'.
     * @param repoID numeric repository identifier.
     * @param enabled Whether telemetry is enabled.
     */
    constructor(languageID: string, repoID: number, enabled = true) {
        this.languageID = languageID
        this.started = Date.now()
        this.repoID = repoID
        this.enabled = enabled
    }

    /**
     * Emit a telemetry event with a durationMs attribute only if the
     * same action has not yet emitted for this instance. This method
     * returns true if an event was emitted and false otherwise.
     */
    public emitOnce(action: string, args: object = {}): boolean {
        if (this.emitted.has(action)) {
            return false
        }

        this.emitted.add(action)
        this.emit(action, args).catch(error => console.error(error))
        return true
    }

    /**
     * Emit a telemetry event with durationMs and languageId attributes.
     */
    public async emit(action: string, args: object = {}): Promise<void> {
        if (!this.enabled) {
            return
        }

        try {
            await sourcegraph.commands.executeCommand('logTelemetryEvent', `codeintel.${action}`, {
                ...args,
                durationMs: this.elapsed(),
                languageId: this.languageID,
                repositoryId: this.repoID,
            })
        } catch {
            // Older version of Sourcegraph may have not registered this
            // command, causing the promise to reject. We can safely ignore
            // this condition.
        }
    }

    private elapsed(): number {
        return Date.now() - this.started
    }
}
