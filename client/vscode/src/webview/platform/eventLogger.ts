import * as Comlink from 'comlink'
import * as uuid from 'uuid'

import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { EventSource, UserEventVariables } from '../../graphql-operations'
import { SourcegraphVSCodeExtensionAPI } from '../contract'

export const ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'
export const COHORT_ID_KEY = 'sourcegraphCohortId'
export const FIRST_SOURCE_URL_KEY = 'sourcegraphSourceUrl'
export const LAST_SOURCE_URL_KEY = 'sourcegraphRecentSourceUrl'
export const DEVICE_ID_KEY = 'sourcegraphDeviceId'

export class EventLogger implements TelemetryService {
    private anonymousUserID = ''
    private cohortID?: string
    private deviceID = ''
    private eventID = 0
    private listeners: Set<(eventName: string) => void> = new Set()
    private sourcegraphVSCodeExtensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI>

    constructor(extensionAPI: Comlink.Remote<SourcegraphVSCodeExtensionAPI>) {
        this.sourcegraphVSCodeExtensionAPI = extensionAPI
        this.initializeLogParameters()
            .then(() => {})
            .catch(() => {})
    }

    /**
     * Log a pageview.
     * Page titles should be specific and human-readable in pascal case, e.g. "SearchResults" or "Blob" or "NewOrg"
     */
    public logViewEvent(pageTitle: string, eventProperties?: any, logAsActiveUser = true, url?: string): void {
        if (!pageTitle) {
            return
        }
        pageTitle = `View${pageTitle}`

        this.trackPageView(pageTitle, logAsActiveUser, eventProperties, url)
    }

    /**
     * Log a user action or event.
     * Event labels should be specific and follow a ${noun}${verb} structure in pascal case, e.g. "ButtonClicked" or "SignInInitiated"
     *
     * @param eventLabel: the event name.
     * @param eventProperties: event properties. These get logged to our database, but do not get
     * sent to our analytics systems. This may contain private info such as repository names or search queries.
     * @param publicArgument: event properties that include only public information. Do NOT
     * include any private information, such as full URLs that may contain private repo names or
     * search queries. The contents of this parameter are sent to our analytics systems.
     */
    public log(eventLabel: string, eventProperties?: any, publicArgument?: any, uri?: string): void {
        if (eventLabel === 'DynamicFilterClicked') {
            eventLabel = 'VSCE_Sidebar_DynamicFiltersClick'
        }
        if (eventLabel === 'SearchSnippetClicked') {
            eventLabel = 'VSCE_Sidebar_RepositoriesClick'
        }
        if (eventLabel === 'SearchReferenceOpened') {
            eventLabel = 'VSCE_Sidebar_SearchReferenceClick'
        }
        for (const listener of this.listeners) {
            listener(eventLabel)
        }
        if (!eventLabel) {
            return
        }
        this.trackAction(eventLabel, eventProperties, publicArgument, uri)
    }

    /**
     * Get the anonymous identifier for this user (used to allow site admins
     * on a Sourcegraph instance to see a count of unique users on a daily,
     * weekly, and monthly basis).
     */
    public getAnonymousUserID(): string {
        return this.anonymousUserID
    }

    /**
     * The cohort ID is generated when the anonymous user ID is generated.
     * Users that have visited before the introduction of cohort IDs will not have one.
     */
    public getCohortID(): string | undefined {
        return this.cohortID
    }

    public getFirstSourceURL(): string {
        return ''
    }

    public getLastSourceURL(): string {
        return ''
    }

    // Device ID is a require field for Amplitude events.
    // https://developers.amplitude.com/docs/http-api-v2
    public getDeviceID(): string {
        return this.deviceID
    }

    // Insert ID is used to deduplicate events in Amplitude.
    // https://developers.amplitude.com/docs/http-api-v2#optional-keys
    public getInsertID(): string {
        return uuid.v4()
    }

    // Event ID is used to deduplicate events in Amplitude.
    // This is used in the case that multiple events with the same userID and timestamp
    // are sent. https://developers.amplitude.com/docs/http-api-v2#optional-keys
    public getEventID(): number {
        this.eventID++
        return this.eventID
    }

    public getReferrer(): string {
        return 'VSCE'
    }

    /**
     * Gets the anonymous user ID and cohort ID of the user from cookies.
     * If user doesn't have an anonymous user ID yet, a new one is generated, along with
     * a cohort ID of the week the user first visited.
     *
     * If the user already has an anonymous user ID before the introduction of cohort IDs,
     * the user will not haved a cohort ID.
     *
     * If user had an anonymous user ID in localStorage, it will be migrated to cookies.
     */
    private async initializeLogParameters(): Promise<void> {
        let anonymousUserID = await this.sourcegraphVSCodeExtensionAPI.getLocalStorageItem(ANONYMOUS_USER_ID_KEY)
        let cohortID = await this.sourcegraphVSCodeExtensionAPI.getLocalStorageItem(COHORT_ID_KEY)
        if (!anonymousUserID) {
            anonymousUserID = uuid.v4()
            cohortID = new Date().toString()
        }

        // Use cookies instead of localStorage so that the ID can be shared with subdomains (about.sourcegraph.com).
        // Always set to renew expiry and migrate from localStorage
        await this.sourcegraphVSCodeExtensionAPI.setLocalStorageItem(ANONYMOUS_USER_ID_KEY, anonymousUserID)
        if (cohortID) {
            await this.sourcegraphVSCodeExtensionAPI.setLocalStorageItem(COHORT_ID_KEY, cohortID)
        }

        let deviceID = await this.sourcegraphVSCodeExtensionAPI.getLocalStorageItem(DEVICE_ID_KEY)
        if (!deviceID) {
            // If device ID does not exist, use the anonymous user ID value so these are consolidated.
            deviceID = anonymousUserID
            await this.sourcegraphVSCodeExtensionAPI.setLocalStorageItem(DEVICE_ID_KEY, deviceID)
        }
        this.anonymousUserID = anonymousUserID
        this.cohortID = cohortID
        this.deviceID = deviceID
    }

    public addEventLogListener(callback: (eventName: string) => void): () => void {
        this.listeners.add(callback)
        return () => this.listeners.delete(callback)
    }

    public trackPageView(eventAction: string, eventProperties?: any, publicArgument?: any, uri?: string): void {
        this.logUserEvent(eventAction, eventProperties, publicArgument, uri)
    }

    public trackAction(eventAction: string, eventProperties?: any, publicArgument?: any, uri?: string): void {
        this.logUserEvent(eventAction, eventProperties, publicArgument, uri)
    }

    private logUserEvent(event: string, eventProperties?: unknown, publicArgument?: unknown, uri?: string): void {
        const userEventVariables = {
            name: event,
            userCookieID: this.getAnonymousUserID(),
            cohortID: this.getCohortID() || null,
            referrer: this.getReferrer(),
            url: uri || '',
            source: EventSource.CODEHOSTINTEGRATION,
            argument: eventProperties ? JSON.stringify(eventProperties) : null,
            publicArgument: publicArgument ? JSON.stringify(publicArgument) : null,
            deviceID: this.getDeviceID(),
            eventID: this.getEventID(),
            insertID: this.getInsertID(),
        }
        this.logging(userEventVariables)
            .then(() => {})
            .catch(() => {})
    }

    private async logging(userEventVariables: UserEventVariables): Promise<void> {
        await this.sourcegraphVSCodeExtensionAPI.logVsceEvent(userEventVariables)
    }
}
