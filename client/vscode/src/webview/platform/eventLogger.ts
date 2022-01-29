import * as Comlink from 'comlink'
import * as uuid from 'uuid'

import { EventSource } from '@sourcegraph/shared/src/graphql-operations'

import { ExtensionCoreAPI } from '../../contract'

import { VsceTelemetryService } from './telemetryService'

export const ANONYMOUS_USER_ID_KEY = 'sourcegraphAnonymousUid'

export class EventLogger implements VsceTelemetryService {
    private anonymousUserID = ''
    private eventID = 0
    private listeners: Set<(eventName: string) => void> = new Set()
    private vsceAPI: Comlink.Remote<ExtensionCoreAPI>

    constructor(extensionAPI: Comlink.Remote<ExtensionCoreAPI>) {
        this.vsceAPI = extensionAPI
        this.initializeLogParameters()
            .then(() => {})
            .catch(() => {})
    }

    /**
     * Log a pageview.
     * Page titles should be specific and human-readable in pascal case, e.g. "SearchResults" or "Blob" or "NewOrg"
     */
    public logViewEvent(pageTitle: string, eventProperties?: any, logAsActiveUser = true, url?: string): void {
        if (pageTitle) {
            this.tracker(`View${pageTitle}`, logAsActiveUser, eventProperties, url)
        }
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
        if (!eventLabel) {
            return
        }
        switch (eventLabel) {
            case 'DynamicFilterClicked':
                eventLabel = 'VSCE_Sidebar_DynamicFiltersClick'
                break
            case 'SearchSnippetClicked':
                eventLabel = 'VSCE_Sidebar_RepositoriesClick'
                break
            case 'SearchReferenceOpened':
                eventLabel = 'VSCE_Sidebar_SearchReferenceClick'
                break
        }
        for (const listener of this.listeners) {
            listener(eventLabel)
        }
        this.tracker(eventLabel, eventProperties, publicArgument, uri)
    }

    /**
     * Get the anonymous identifier for this user (used to allow site admins
     * on a Sourcegraph instance to see a count of unique users on a daily,
     * weekly, and monthly basis).
     */
    public getAnonymousUserID(): string {
        return this.anonymousUserID
    }

    // Event ID is used to deduplicate events in Amplitude.
    // This is used in the case that multiple events with the same userID and timestamp
    // are sent. https://developers.amplitude.com/docs/http-api-v2#optional-keys
    public getEventID(): number {
        this.eventID++
        return this.eventID
    }

    /**
     * Gets the anonymous user ID and cohort ID of the user from VSCE storage utility.
     * If user doesn't have an anonymous user ID yet, a new one is generated
     */
    private async initializeLogParameters(): Promise<void> {
        let anonymousUserID = await this.vsceAPI.getLocalStorageItem(ANONYMOUS_USER_ID_KEY)
        if (!anonymousUserID) {
            anonymousUserID = uuid.v4()
            await this.vsceAPI.setLocalStorageItem(ANONYMOUS_USER_ID_KEY, anonymousUserID)
        }
        this.anonymousUserID = anonymousUserID
    }

    public addEventLogListener(callback: (eventName: string) => void): () => void {
        this.listeners.add(callback)
        return () => this.listeners.delete(callback)
    }

    public tracker(eventName: string, eventProperties?: unknown, publicArgument?: unknown, uri?: string): void {
        const userEventVariables = {
            event: eventName,
            userCookieID: this.getAnonymousUserID(),
            referrer: 'VSCE',
            url: uri || '',
            source: EventSource.CODEHOSTINTEGRATION,
            argument: eventProperties ? JSON.stringify(eventProperties) : null,
            publicArgument: publicArgument ? JSON.stringify(publicArgument) : null,
            deviceID: this.getAnonymousUserID(),
            eventID: this.getEventID(),
        }
        this.vsceAPI
            .logEvents(userEventVariables)
            .then(() => {})
            .catch(error => console.log(error))
    }
}
