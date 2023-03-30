import { noop } from 'lodash'
import { Observable, Subscription } from 'rxjs'
import * as uuid from 'uuid'

import { gql } from '@sourcegraph/http-client'
import { EventSource } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { version as packageVersion } from '../../package.json'
import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

const uidKey = 'sourcegraphAnonymousUid'

interface DriverType {
    type: 'browser' | 'vscode'
}

// Event Logger for the Cody Extension
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
    private driverType: DriverType
    private version = packageVersion

    /**
     * Buffered Observable for the latest Sourcegraph URL
     */

    constructor(private client: SourcegraphGraphQLAPIClient, private sourcegraphURL: string, driverType: DriverType) {
        // Fetch user ID on initial load.
        this.getAnonUserID().catch(noop)
        this.driverType = driverType
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

        if (this.driverType.type === 'browser') {
            let id = localStorage.getItem(uidKey)
            if (id === null) {
                id = this.generateAnonUserID()
                localStorage.setItem(uidKey, id)
            }
            this.uid = id
            return this.uid
        }

        let { sourcegraphAnonymousUid } = await localStorage.sync.get()
        if (!sourcegraphAnonymousUid) {
            sourcegraphAnonymousUid = this.generateAnonUserID()
            await localStorage.sync.set({ sourcegraphAnonymousUid })
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
                argument: { platform: this.driverType, version: this.version, ...eventProperties },
                publicArgument: { platform: this.driverType, version: this.version, ...publicArgument },
            },
            this.client
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

/**
 * Log a raw user action on the associated Sourcegraph instance
 */
export const logEvent = async (
    event: { name: string; userCookieID: string; url: string; argument?: string | {}; publicArgument?: string | {} },
    client: SourcegraphGraphQLAPIClient
): Promise<void> => {
    await client.fetch({
        request: gql`
            mutation logEvent(
                $name: String!
                $userCookieID: String!
                $url: String!
                $source: EventSource!
                $argument: String
                $publicArgument: String
            ) {
                logEvent(
                    event: $name
                    userCookieID: $userCookieID
                    url: $url
                    source: $source
                    argument: $argument
                    publicArgument: $publicArgument
                ) {
                    alwaysNil
                }
            }
        `,
        variables: {
            ...event,
            source: EventSource.CODY,
            argument: event.argument && JSON.stringify(event.argument),
            publicArgument: event.publicArgument && JSON.stringify(event.publicArgument),
        },
        mightContainPrivateInfo: false,
        // eslint-disable-next-line rxjs/no-ignored-subscription
    })
}
