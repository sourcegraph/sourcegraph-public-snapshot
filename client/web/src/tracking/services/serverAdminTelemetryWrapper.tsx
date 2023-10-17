import { recordEvent } from '../../user/settings/backend'

class ServerAdminWrapper {
    public trackPageView(
        feature: string,
        action: string,
        source: any,
        parameters?: any,
        marketingTracking?: any
    ): void {
        recordEvent(feature, action, source, parameters, marketingTracking)
    }

    public trackAction(feature: string, action: string, source: any, parameters?: any, marketingTracking?: any): void {
        recordEvent(feature, action, source, parameters, marketingTracking)
    }
}

export const serverAdmin = new ServerAdminWrapper()
