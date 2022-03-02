import * as Comlink from 'comlink'
import * as uuid from 'uuid'

import { EventSource } from '@sourcegraph/shared/src/graphql-operations'

import { version } from '../../../package.json'
import { ExtensionCoreAPI } from '../../contract'
import { INSTANCE_VERSION_NUMBER_KEY, ANONYMOUS_USER_ID_KEY } from '../../settings/LocalStorageService'

import { VsceTelemetryService } from './telemetryService'

// Event Logger for VS Code Extension
export class EventLogger implements VsceTelemetryService {
    private anonymousUserID = ''
    private instanceVersion = ''
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
            // Adding eventProperties when none is provided would break the search history sync to cloud
            this.tracker(
                `View${pageTitle}`,
                eventProperties ? { platform: 'vscode', version, ...eventProperties } : eventProperties,
                logAsActiveUser,
                url
            )
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
                eventLabel = 'VSCESidebarDynamicFiltersClick'
                break
            case 'SearchSnippetClicked':
                eventLabel = 'VSCESidebarRepositoriesClick'
                break
            case 'SearchReferenceOpened':
                eventLabel = 'VSCESidebarSearchReferenceClick'
                break
        }
        for (const listener of this.listeners) {
            listener(eventLabel)
        }
<<<<<<< HEAD
        this.tracker(eventLabel, eventProperties, publicArgument, uri)
=======
        this.tracker(
            eventLabel,
            eventProperties ? { platform: 'vscode', version, ...eventProperties } : { platform: 'vscode', version },
            publicArgument,
            uri
        )
>>>>>>> 15b3d72fe5 (Collect pings for ide usage metrics)
    }

    /**
     * Gets the anonymous user ID and cohort ID of the user from VSCE storage utility.
     * If user doesn't have an anonymous user ID yet, a new one is generated
     */
    private async initializeLogParameters(): Promise<void> {
        let anonymousUserID = await this.vsceAPI.getLocalStorageItem(ANONYMOUS_USER_ID_KEY)
        // instance version is set during the initial authenticating step
        const instanceVersion = await this.vsceAPI.getLocalStorageItem(INSTANCE_VERSION_NUMBER_KEY)
        if (!anonymousUserID) {
            anonymousUserID = uuid.v4()
            await this.vsceAPI.setLocalStorageItem(ANONYMOUS_USER_ID_KEY, anonymousUserID)
        }
        this.anonymousUserID = anonymousUserID
        this.instanceVersion = instanceVersion
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
     * Regular instance version format: 3.38.2
     * Insider version format: 134683_2022-03-02_5188fes0101
     */
    public getEventSourceType(): EventSource {
        // assume instance version longer than 8 is using insider version
        const flattenVersion = this.instanceVersion.length > 8 ? '999999' : this.instanceVersion.split('.').join()
        // instances below 3.38.0 does not support EventSource.IDEEXTENSION
        return flattenVersion > '3380' ? EventSource.IDEEXTENSION : EventSource.BACKEND
    }

    /**
     * Event ID is used to deduplicate events in Amplitude.
     * This is used in the case that multiple events with the same userID and timestamp
     * are sent. https://developers.amplitude.com/docs/http-api-v2#optional-keys
     */
    public getEventID(): number {
        this.eventID++
        return this.eventID
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
            source: this.getEventSourceType(),
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
