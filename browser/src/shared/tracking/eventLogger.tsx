import { noop } from 'lodash'
import { switchMap, take } from 'rxjs/operators'
import uuid from 'uuid'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { TelemetryService } from '../../../../shared/src/telemetry/telemetryService'
import { storage } from '../../browser/storage'
import { isInPage } from '../../context'
import { logUserEvent } from '../backend/userEvents'
import { observeSourcegraphURL } from '../util/context'

const uidKey = 'sourcegraphAnonymousUid'

export class EventLogger implements TelemetryService {
    private uid: string | null = null
    private sourcegraphURL: string | undefined
    private requestGraphQL: PlatformContext['requestGraphQL']

    constructor(isExtension: boolean, requestGraphQL: PlatformContext['requestGraphQL']) {
        observeSourcegraphURL(isExtension)
            .pipe(take(1))
            .subscribe(sourcegraphURL => (this.sourcegraphURL = sourcegraphURL))
        this.requestGraphQL = requestGraphQL
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
     * Log a user action on the associated self-hosted Sourcegraph instance (allows site admins on a private
     * Sourcegraph instance to see a count of unique users on a daily, weekly, and monthly basis).
     *
     * This is never sent to Sourcegraph.com (i.e., when using the integration with open source code).
     */
    public async logCodeIntelligenceEvent(event: GQL.UserEvent): Promise<void> {
        if (this.sourcegraphURL) {
            const anonUserId = await this.getAnonUserID()
            logUserEvent(event, anonUserId, this.sourcegraphURL, this.requestGraphQL)
        }
    }

    /**
     * Implements {@link TelemetryService}.
     *
     * @todo Handle arbitrary action IDs.
     *
     * @param eventName The ID of the action executed.
     */
    public log(eventName: string): void {
        if (eventName === 'goToDefinition' || eventName === 'goToDefinition.preloaded' || eventName === 'hover') {
            this.logCodeIntelligenceEvent(GQL.UserEvent.CODEINTELINTEGRATION)
        } else if (eventName === 'findReferences') {
            this.logCodeIntelligenceEvent(GQL.UserEvent.CODEINTELINTEGRATIONREFS)
        }
    }
}
