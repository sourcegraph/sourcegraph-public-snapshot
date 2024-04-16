import { logEvent } from '../../user/settings/backend'

class ServerAdminWrapper {
    /**
     * @deprecated Use a TelemetryRecorder or TelemetryRecorderProvider from
     * src/telemetry instead.
     */
    public trackPageView(eventAction: string, eventProperties?: any, publicArgument?: any): void {
        logEvent(eventAction, eventProperties, publicArgument)
    }

    /**
     * @deprecated Use a TelemetryRecorder or TelemetryRecorderProvider from
     * src/telemetry instead.
     */
    public trackAction(eventAction: string, eventProperties?: any, publicArgument?: any): void {
        logEvent(eventAction, eventProperties, publicArgument)
    }
}

/**
 * @deprecated Use a TelemetryRecorder or TelemetryRecorderProvider from
 * src/telemetry instead.
 */
export const serverAdmin = new ServerAdminWrapper()
