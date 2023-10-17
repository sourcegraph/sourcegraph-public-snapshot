import { eventRecorder } from '../../tracking/eventRecorder'
import { logEvent } from '../../user/settings/backend'

class ServerAdminWrapper {
    public trackPageView(eventAction: string, eventProperties?: any, publicArgument?: any): void {
        logEvent(eventAction, eventProperties, publicArgument)
    }

    public trackAction(eventAction: string, eventProperties?: any, publicArgument?: any): void {
        logEvent(eventAction, eventProperties, publicArgument)
    }

    public trackTelemetryPageView(feature: string, action: string, parameter?: any): void {
        eventRecorder.record(feature, action, parameter)
    }

    public trackTelemetryAction(feature: string, action: string, parameter?: any): void {
        eventRecorder.record(feature, action, parameter)
    }
}

export const serverAdmin = new ServerAdminWrapper()
