/**
 * Props interface that can be extended by React components depending on the TelemetryServiceV2.
 */
export interface TelemetryV2Props {
    /**
     * A telemetry service v2 implementation to log events.
     */
    telemetryServiceV2: TelemetryServiceV2
}

/**
 * The telemetry service v2 logs events.
 */
export interface TelemetryServiceV2 {
    /**
     * Record an event (by sending it to the server).
     */
    record(eventName: string, eventProperties?: any, publicArgument?: any): void
    /**
     * Log a pageview event (by sending it to the server).
     * Adheres to the new event naming policy
     */
    recordPageView(eventName: string, eventProperties?: any, publicArgument?: any): void
    /**
     * Listen for event logs
     *
     * @returns a cleanup/removeTelemetryEventListener function
     */
    addTelemetryEventListener?(callback: (eventName: string) => void): () => void
}

/**
 * A noop telemetry service.
 */
export const NOOP_TELEMETRY_SERVICE: TelemetryServiceV2 = {
    record: () => {
        /* noop */
    },
    recordPageView: () => {
        /* noop */
    },
}
