import { logEvent } from '../../user/settings/backend'
import { eventRecorder } from '../tracking/eventRecorder'

class ServerAdminWrapper {
    public trackPageView(eventAction: string, eventProperties?: any, publicArgument?: any): void {
        logEvent(eventAction, eventProperties, publicArgument)
    }

    public trackAction(eventAction: string, eventProperties?: any, publicArgument?: any): void {
        logEvent(eventAction, eventProperties, publicArgument)
    }

    public trackTelemetryPageView(
        feature: string,
        action: string,
        source: any,
        parameter?: any,
        marketingTracking?: any
    ): void {
        eventRecorder.record(feature, action, source, parameter, marketingTracking)
    }

    public trackTelemetryAction(
        feature: string,
        action: string,
        source: any,
        parameter?: any,
        marketingTracking?: any
    ): void {
        eventRecorder.record(feature, action, source, parameter, marketingTracking)
    }
}

export const serverAdmin = new ServerAdminWrapper()
