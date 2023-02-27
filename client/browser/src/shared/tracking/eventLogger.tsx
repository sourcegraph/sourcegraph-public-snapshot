import { noop } from 'lodash'
import { Observable, Subscription } from 'rxjs'
import * as uuid from 'uuid'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { storage } from '../../browser-extension/web-extension-api/storage'
import { logEvent } from '../backend/userEvents'
import { isInPage } from '../context'
import { getExtensionVersion, getPlatformName } from '../util/context'

const uidKey = 'sourcegraphAnonymousUid'

/**
 * Telemetry Service which only logs when the enable flag is set. Accepts an
 * observable that emits the enabled value.
 *
 * This was implemented as a wrapper around TelemetryService in order to avoid
 * modifying EventLogger, but the enabled flag could be rolled into EventLogger.
 *
 * TODO: Potential to be improved by buffering log events until the first emit
 * of the enabled value.
 */
export class ConditionalTelemetryService implements TelemetryService {
    /** Log events are passed on to the inner TelemetryService */
    private subscription = new Subscription()

    /** The enabled state set by an observable, provided upon instantiation */
    private isEnabled = false

    constructor(private innerTelemetryService: TelemetryService, isEnabled: Observable<boolean>) {
        this.subscription.add(
            isEnabled.subscribe(value => {
                this.isEnabled = value
            })
        )
    }
    public log(eventName: string, eventProperties?: any, publicArgument?: any): void {
        // Wait for this.isEnabled to get a new value
        setTimeout(() => {
            if (this.isEnabled) {
                this.innerTelemetryService.log(eventName, eventProperties, publicArgument)
            }
        })
    }
    /**
     * @deprecated Use logPageView instead
     */
    public logViewEvent(eventName: string, eventProperties?: any): void {
        // Wait for this.isEnabled to get a new value
        setTimeout(() => {
            if (this.isEnabled) {
                this.innerTelemetryService.logViewEvent(eventName, eventProperties)
            }
        })
    }
    public logPageView(eventName: string, eventProperties?: any, publicArgument?: any): void {
        // Wait for this.isEnabled to get a new value
        setTimeout(() => {
            if (this.isEnabled) {
                this.innerTelemetryService.logPageView(eventName, eventProperties, publicArgument)
            }
        })
    }
    /**
     * Logs page view events, adding a suffix
     *
     * @returns
     *
     */
    public unsubscribe(): void {
        // Reset initial state
        this.isEnabled = false
        return this.subscription.unsubscribe()
    }
}

export class EventLogger implements TelemetryService {
    private uid: string | null = null

    private platform = getPlatformName()
    private version = getExtensionVersion()

    /**
     * Buffered Observable for the latest Sourcegraph URL
     */

    constructor(private requestGraphQL: PlatformContext['requestGraphQL'], private sourcegraphURL: string) {
        // Fetch user ID on initial load.
        this.getAnonUserID().catch(noop)
    }

    /**
     * Generate a new anonymous user ID if one has not yet been set and stored.
     */
    private generateAnonUserID = (): string => uuid.v4()

    /**
     * Get the anonymous identifier for this user (allows site admins on a private Sourcegraph
     * instance to see a count of unique users on a daily, weekly, and monthly basis).
     *
     * Not used at all for public/Sourcegraph.com usage.
     */
    private async getAnonUserID(): Promise<string> {
        if (this.uid) {
            return this.uid
        }

        if (isInPage) {
            let id = localStorage.getItem(uidKey)
            if (id === null) {
                id = this.generateAnonUserID()
                localStorage.setItem(uidKey, id)
            }
            this.uid = id
            return this.uid
        }

        let { sourcegraphAnonymousUid } = await storage.sync.get()
        if (!sourcegraphAnonymousUid) {
            sourcegraphAnonymousUid = this.generateAnonUserID()
            await storage.sync.set({ sourcegraphAnonymousUid })
        }
        this.uid = sourcegraphAnonymousUid
        return sourcegraphAnonymousUid
    }

    /**
     * Log a user action on the associated Sourcegraph instance
     */
    private async logEvent(event: string, eventProperties?: any, publicArgument?: any): Promise<void> {
        const anonUserId = await this.getAnonUserID()
        logEvent(
            {
                name: event,
                userCookieID: anonUserId,
                url: this.sourcegraphURL,
                argument: { platform: this.platform, version: this.version, ...eventProperties },
                publicArgument: { platform: this.platform, version: this.version, ...publicArgument },
            },
            this.requestGraphQL
        )
    }

    /**
     * Implements {@link TelemetryService}.
     *
     * @todo Handle arbitrary action IDs.
     *
     * @param eventName The ID of the action executed.
     */
    public async log(eventName: string, eventProperties?: any, publicArgument?: any): Promise<void> {
        await this.logEvent(eventName, eventProperties, publicArgument)
    }

    /**
     * Implements {@link TelemetryService}.
     *
     * @deprecated Use logPageView instead
     *
     * @param pageTitle The title of the page being viewed.
     */
    public async logViewEvent(pageTitle: string, eventProperties?: any): Promise<void> {
        await this.logEvent(`View${pageTitle}`, eventProperties)
    }

    /**
     * Implements {@link TelemetryService}.
     *
     * @param eventName The name of the entity being viewed.
     */
    public async logPageView(eventName: string, eventProperties?: any, publicArgument?: any): Promise<void> {
        await this.logEvent(`${eventName}Viewed`, eventProperties, publicArgument)
    }
}
