import { noop } from 'lodash'
import { type Observable, Subscription } from 'rxjs'
import * as uuid from 'uuid'

import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import type { TelemetryServiceV2 } from '@sourcegraph/shared/src/telemetry/telemetryServiceV2'

import { storage } from '../../browser-extension/web-extension-api/storage'
import { recordEvent } from '../backend/userTelemetry'
import { isInPage } from '../context'

const uidKey = 'sourcegraphAnonymousUid'

/**
 * Telemetry Service which only logs when the enable flag is set. Accepts an
 * observable that emits the enabled value.
 *
 */
export class ConditionalTelemetryService implements TelemetryServiceV2 {
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
    public record(feature: string, action?: any, source?: any, parameters?: any, marketingTracking?: any): void {
        // Wait for this.isEnabled to get a new value
        setTimeout(() => {
            if (this.isEnabled) {
                this.innerTelemetryService.record(feature, action, source, parameters, marketingTracking)
            }
        })
    }
    public recordPageView(feature: string, source: any, parameters?: any, marketingTracking?: any): void {
        // Wait for this.isEnabled to get a new value
        setTimeout(() => {
            if (this.isEnabled) {
                this.innerTelemetryService.recordPageView(feature, source, parameters, marketingTracking)
            }
        })
    }
    /**
     * Records page view events, adding a suffix
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

export class EventRecorder implements TelemetryServiceV2 {
    private uid: string | null = null

    /**
     * Buffered Observable for the latest Sourcegraph URL
     */

    constructor(private requestGraphQL: PlatformContext['requestGraphQL']) {
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
     * Record a user action on the associated Sourcegraph instance
     */
    private async recordEvent(
        feature: string,
        action?: any,
        source?: any,
        parameters?: any,
        marketingTracking?: any
    ): Promise<void> {
        recordEvent(
            {
                feature: feature,
                action: action,
                source: source,
                parameters: parameters,
                marketingTracking: marketingTracking,
            },
            this.requestGraphQL
        )
    }

    /**
     * Implements {@link TelemetryService}.
     *
     * @todo Handle arbitrary action IDs.
     *
     * @param feature The ID of the feature executed.
     */
    public async record(
        feature: string,
        action?: string,
        source?: string,
        parameters?: any,
        marketingTracking?: any
    ): Promise<void> {
        await this.recordEvent(feature, action, source, parameters, marketingTracking)
    }

    /**
     * Implements {@link TelemetryService}.
     *
     * @param eventName The name of the entity being viewed.
     */
    public async recordPageView(
        feature: string,
        source: any,
        parameters?: any,
        marketingTracking?: any
    ): Promise<void> {
        await this.recordEvent(feature, 'VIEWED', source, parameters, marketingTracking)
    }
}
