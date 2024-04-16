/**
 * Props interface that can be extended by React components depending on the TelemetryService.
 * These properties are part of {@link PlatformContext}.
 *
 * @deprecated Use TelemetryV2Props for a '@sourcegraph/telemetry' implementation
 * instead.
 */
export interface TelemetryProps {
    /**
     * A telemetry service implementation to log events.
     *
     * @deprecated Use telemetryRecorder instead from TelemetryV2Props, if it is
     * non-null (i.e. if the new SDK is available for the platform).
     */
    telemetryService: TelemetryService
}

/**
 * The telemetry service logs events.
 *
 * @deprecated Use a '@sourcegraph/telemetry' implementation instead.
 */
export interface TelemetryService {
    /**
     * Log an event (by sending it to the server).
     *
     * @deprecated Use a '@sourcegraph/telemetry' implementation instead where
     * available.
     */
    log(eventName: string, eventProperties?: any, publicArgument?: any): void
    /**
     * @deprecated use logPageView instead
     *
     * Log a pageview event (by sending it to the server).
     */
    logViewEvent(eventName: string, eventProperties?: any, publicArgument?: any): void
    /**
     * Log a pageview event (by sending it to the server).
     * Adheres to the new event naming policy
     *
     * @deprecated Use a '@sourcegraph/telemetry' implementation instead.
     */
    logPageView(eventName: string, eventProperties?: any, publicArgument?: any): void
    /**
     * Listen for event logs
     *
     * @deprecated Use a '@sourcegraph/telemetry' implementation instead.
     * @returns a cleanup/removeEventListener function
     */
    addEventLogListener?(callback: (eventName: string) => void): () => void
}

/**
 * A noop telemetry service.
 */
export const NOOP_TELEMETRY_SERVICE: TelemetryService = {
    log: () => {
        /* noop */
    },
    logViewEvent: () => {
        /* noop */
    },
    logPageView: () => {
        /* noop */
    },
}
