import * as uuid from 'uuid'
import { TelemetryService } from '../../../shared/src/telemetry/telemetryService'
import { browserExtensionMessageReceived, handleQueryEvents, pageViewQueryParameters } from './analyticsUtils'
import { serverAdmin } from './services/serverAdminWrapper'

const uidKey = 'sourcegraphAnonymousUid'

/**
 * Props interface that can be extended by React components that need access to the full webapp EventLogger.
 * The EventLogger provides more functionality than TelemetryService, but can only be used in the webapp.
 */
export interface EventLoggerProps {
    /**
     * The full webapp EventLogger to log telemetry events.
     */
    telemetryService: EventLogger
}

export class EventLogger implements TelemetryService {
    private hasStrippedQueryParameters = false

    private anonUid?: string

    constructor() {
        // EventLogger is never teared down
        // eslint-disable-next-line rxjs/no-ignored-subscription
        browserExtensionMessageReceived.subscribe(() => {
            this.log('BrowserExtensionConnectedToServer')

            if (localStorage && localStorage.getItem('eventLogDebug') === 'true') {
                console.debug('%cBrowser extension detected, sync completed', 'color: #aaa')
            }
        })
    }

    /**
     * Log a pageview.
     * Page titles should be specific and human-readable in pascal case, e.g. "SearchResults" or "Blob" or "NewOrg"
     */
    public logViewEvent(pageTitle: string, logAsActiveUser = true): void {
        if (window.context?.userAgentIsBot || !pageTitle) {
            return
        }
        pageTitle = `View${pageTitle}`

        const props = pageViewQueryParameters(window.location.href)
        serverAdmin.trackPageView(pageTitle, logAsActiveUser)
        this.logToConsole(pageTitle, props)

        // Use flag to ensure URL query params are only stripped once
        if (!this.hasStrippedQueryParameters) {
            handleQueryEvents(window.location.href)
            this.hasStrippedQueryParameters = true
        }
    }

    /**
     * Log a user action or event.
     * Event labels should be specific and follow a ${noun}${verb} structure in pascal case, e.g. "ButtonClicked" or "SignInInitiated"
     */
    public log(eventLabel: string, eventProperties?: any): void {
        if (window.context?.userAgentIsBot || !eventLabel) {
            return
        }
        serverAdmin.trackAction(eventLabel, eventProperties)
        this.logToConsole(eventLabel, eventProperties)
    }

    private logToConsole(eventLabel: string, object?: any): void {
        if (localStorage && localStorage.getItem('eventLogDebug') === 'true') {
            console.debug('%cEVENT %s', 'color: #aaa', eventLabel, object)
        }
    }

    /**
     * Get the anonymous identifier for this user (used to allow site admins
     * on a Sourcegraph instance to see a count of unique users on a daily,
     * weekly, and monthly basis).
     */
    public getAnonUserID(): string {
        if (this.anonUid) {
            return this.anonUid
        }

        let id = localStorage.getItem(uidKey)
        if (id === null || id === '') {
            id = uuid.v4()
            localStorage.setItem(uidKey, id)
        }
        this.anonUid = id
        return this.anonUid
    }
}

export const eventLogger = new EventLogger()
