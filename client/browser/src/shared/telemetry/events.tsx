import { noop } from 'lodash'
import { type Observable, Subscription } from 'rxjs'
import * as uuid from 'uuid'

import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { TelemetryServiceV2, EventName, EventParameters } from '@sourcegraph/shared/src/telemetry/telemetryServiceV2'

import { storage } from '../../browser-extension/web-extension-api/storage'
import { isInPage } from '../context'
import { getExtensionVersion, getPlatformName } from '../util/context'

const uidKey = 'sourcegraphAnonymousUid'

/**
 * ConditionalTelemetryV2Service is a TelemetryV2Service which only logs when
 * the enable flag is set. Accepts an observable that emits the enabled value.
 *
 * EXPERIMENTAL
 */
export class ConditionalTelemetryV2Service implements TelemetryServiceV2 {
    /** Log events are passed on to the inner TelemetryService */
    private subscription = new Subscription()

    /** The enabled state set by an observable, provided upon instantiation */
    private isEnabled = false

    constructor(private innerTelemetryService: TelemetryServiceV2, isEnabled: Observable<boolean>) {
        this.subscription.add(
            isEnabled.subscribe(value => {
                this.isEnabled = value
            })
        )
    }

    public record(eventName: EventName, parameters?: EventParameters): void {
        // Wait for this.isEnabled to get a new value
        setTimeout(() => {
            if (this.isEnabled) {
                this.innerTelemetryService.record(eventName, parameters)
            }
        })
    }

    //going to delete this, just for testing
    public recordString(eventName: string, parameters?: EventParameters): void {
        // Wait for this.isEnabled to get a new value
        setTimeout(() => {
            if (this.isEnabled) {
                this.innerTelemetryService.recordString(eventName, parameters)
            }
        })
    }

    public unsubscribe(): void {
        // Reset initial state
        this.isEnabled = false
        return this.subscription.unsubscribe()
    }
}

/**
 * EventRecorder is the base implementation of TelemetryV2Service.
 *
 * EXPERIMENTAL
 */
export class EventRecorder implements TelemetryServiceV2 {
    private uid: string | null = null

    private platform = getPlatformName()
    private version = getExtensionVersion()

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

    constructor(private requestGraphQL: PlatformContext['requestGraphQL'], private sourcegraphURL: string) {
        // Fetch user ID on initial load.
        this.getAnonUserID().catch(noop)
    }

    /**
     * EXPERIMENTAL
     *
     * This implementation currently no-ops.
     */
    public record(eventName: EventName, parameters?: EventParameters): void {
        console.log(eventName)
    }

    /**
     * EXPERIMENTAL
     *
     * This implementation currently no-ops.
     */
    public recordString(eventName: string, parameters?: EventParameters): void {
        console.log(eventName, this.requestGraphQL)
    }
}
