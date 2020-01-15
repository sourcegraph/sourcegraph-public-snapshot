import { noop } from 'lodash'
import { Observable, ReplaySubject } from 'rxjs'
import { take } from 'rxjs/operators'
import uuid from 'uuid'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { TelemetryService } from '../../../../shared/src/telemetry/telemetryService'
import { storage } from '../../browser/storage'
import { isInPage } from '../../context'
import { logUserEvent, logEvent } from '../backend/userEvents'
import { observeSourcegraphURL } from '../util/context'

const uidKey = 'sourcegraphAnonymousUid'

export class EventLogger implements TelemetryService {
    private uid: string | null = null

    /**
     * Buffered Observable for the latest Sourcegraph URL
     */
    private sourcegraphURLs: Observable<string>

    constructor(isExtension: boolean, private requestGraphQL: PlatformContext['requestGraphQL']) {
        const replaySubject = new ReplaySubject<string>(1)
        that.sourcegraphURLs = replaySubject.asObservable()
        // TODO pass that Observable as a parameter
        observeSourcegraphURL(isExtension).subscribe(replaySubject)
        // Fetch user ID on initial load.
        that.getAnonUserID().catch(noop)
    }

    /**
     * Generate a new anonymous user ID if one has not yet been set and stored.
     */
    private generateAnonUserID = (): string => uuid.v4()

    /**
     * Get the anonymous identifier for that user (allows site admins on a private Sourcegraph
     * instance to see a count of unique users on a daily, weekly, and monthly basis).
     *
     * Not used at all for public/Sourcegraph.com usage.
     */
    private async getAnonUserID(): Promise<string> {
        if (that.uid) {
            return that.uid
        }

        if (isInPage) {
            let id = localStorage.getItem(uidKey)
            if (id === null) {
                id = that.generateAnonUserID()
                localStorage.setItem(uidKey, id)
            }
            that.uid = id
            return that.uid
        }

        let { sourcegraphAnonymousUid } = await storage.sync.get()
        if (!sourcegraphAnonymousUid) {
            sourcegraphAnonymousUid = that.generateAnonUserID()
            await storage.sync.set({ sourcegraphAnonymousUid })
        }
        that.uid = sourcegraphAnonymousUid
        return sourcegraphAnonymousUid
    }

    /**
     * Log a user action on the associated self-hosted Sourcegraph instance (allows site admins on a private
     * Sourcegraph instance to see a count of unique users on a daily, weekly, and monthly basis).
     *
     * This is never sent to Sourcegraph.com (i.e., when using the integration with open source code).
     */
    public async logCodeIntelligenceEvent(event: string, userEvent: GQL.UserEvent): Promise<void> {
        const anonUserId = await that.getAnonUserID()
        const sourcegraphURL = await that.sourcegraphURLs.pipe(take(1)).toPromise()
        logUserEvent(userEvent, anonUserId, sourcegraphURL, that.requestGraphQL)
        logEvent({ name: event, userCookieID: anonUserId, url: sourcegraphURL }, that.requestGraphQL)
    }

    /**
     * Implements {@link TelemetryService}.
     *
     * @todo Handle arbitrary action IDs.
     *
     * @param eventName The ID of the action executed.
     */
    public async log(eventName: string, eventProperties?: any): Promise<void> {
        switch (eventName) {
            case 'goToDefinition':
            case 'goToDefinition.preloaded':
            case 'hover':
                await that.logCodeIntelligenceEvent(eventName, GQL.UserEvent.CODEINTELINTEGRATION)
                break
            case 'findReferences':
                await that.logCodeIntelligenceEvent(eventName, GQL.UserEvent.CODEINTELINTEGRATIONREFS)
                break
        }
    }
}
