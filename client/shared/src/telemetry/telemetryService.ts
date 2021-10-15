import { EMPTY } from 'rxjs'

import { PlatformContext } from '../platform/context'

/**
 * Props interface that can be extended by React components depending on the TelemetryService.
 */
export interface TelemetryProps {
    /**
     * A telemetry service implementation to log events.
     */
    telemetryService: TelemetryService
}

/**
 * The telemetry service logs events.
 */
export interface TelemetryService {
    /**
     * Log an event (by sending it to the server).
     */
    log(eventName: string, eventProperties?: any, publicArgument?: any): void
    /**
     * Log a pageview event (by sending it to the server).
     */
    logViewEvent(eventName: string, eventProperties?: any, publicArgument?: any): void
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
}

export const NOOP_PLATFORM_CONTEXT: Pick<PlatformContext, 'requestGraphQL'> = {
    requestGraphQL: () => EMPTY,
}
