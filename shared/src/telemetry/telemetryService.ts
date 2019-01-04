/**
 * The telemetry service logs events.
 */
export interface TelemetryService {
    /**
     * Log an event (by sending it to the server).
     */
    log(eventName: string): void
}

/**
 * A noop telemetry service.
 */
export const NOOP_TELEMETRY_SERVICE: TelemetryService = {
    log: () => {
        /* noop */
    },
}
