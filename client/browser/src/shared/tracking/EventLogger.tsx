import uuid from 'uuid'
import { TelemetryService } from '../../../../../shared/src/telemetry/telemetryService'
import storage from '../../browser/storage'
import { isInPage } from '../../context'
import { logUserEvent } from '../backend/userEvents'

const uidKey = 'sourcegraphAnonymousUid'

export class EventLogger implements TelemetryService {
    private uid: string

    constructor() {
        // Fetch user ID on initial load.
        this.getAnonUserID().then(
            () => {
                /* noop */
            },
            () => {
                /* noop */
            }
        )
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
    private getAnonUserID = (): Promise<string> =>
        new Promise(resolve => {
            if (this.uid) {
                resolve(this.uid)
                return
            }

            if (isInPage) {
                let id = localStorage.getItem(uidKey)
                if (id === null) {
                    id = this.generateAnonUserID()
                    localStorage.setItem(uidKey, id)
                }
                this.uid = id
                resolve(this.uid)
            } else {
                storage.getSyncItem(uidKey, ({ sourcegraphAnonymousUid }) => {
                    if (sourcegraphAnonymousUid === '') {
                        sourcegraphAnonymousUid = this.generateAnonUserID()
                        storage.setSync({ sourcegraphAnonymousUid })
                    }
                    this.uid = sourcegraphAnonymousUid
                    resolve(sourcegraphAnonymousUid)
                })
            }
        })

    /**
     * Log a user action on the associated self-hosted Sourcegraph instance (allows site admins on a private
     * Sourcegraph instance to see a count of unique users on a daily, weekly, and monthly basis).
     *
     * This is never sent to Sourcegraph.com (i.e., when using the integration with open source code).
     */
    public logCodeIntelligenceEvent(): void {
        this.getAnonUserID().then(
            anonUserId => logUserEvent('CODEINTELINTEGRATION', anonUserId),
            () => {
                /* noop */
            }
        )
    }

    /**
     * Implements {@link TelemetryService}.
     *
     * @todo Use the eventName. It is currently ignored.
     *
     * @param _eventName This parameter is ignored; see the @todo.
     */
    public log(_eventName: string): void {
        this.logCodeIntelligenceEvent()
    }
}
