export const vscodeTelemetryService: VsceTelemetryService = {
    log: () => {},
    logViewEvent: () => {},
}

/**
 * Props interface that can be extended by React components depending on the TelemetryService.
 */
export interface TelemetryProps {
    /**
     * A telemetry service implementation to log events.
     */
    telemetryService: VsceTelemetryService
}

/**
 * The telemetry service logs events.
 */
export interface VsceTelemetryService {
    /**
     * Log an event (by sending it to the server).
     */
    log(eventName: string, eventProperties?: any, publicArgument?: any, uri?: string): void
    /**
     * Log a pageview event (by sending it to the server).
     */
    logViewEvent(eventName: string, eventProperties?: any, publicArgument?: any, uri?: string): void
}

/**
 * A noop telemetry service.
 */
export const NOOP_TELEMETRY_SERVICE: VsceTelemetryService = {
    log: () => {
        /* noop */
    },
    logViewEvent: () => {
        /* noop */
    },
}
