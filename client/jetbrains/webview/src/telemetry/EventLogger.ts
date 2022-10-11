import { EMPTY, Subject } from 'rxjs'
import { bufferTime, catchError, concatMap } from 'rxjs/operators'

import { gql } from '@sourcegraph/http-client'
import { EventSource } from '@sourcegraph/shared/src/graphql-operations'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Event, LogEventsResult, LogEventsVariables } from '../graphql-operations'
import { requestGraphQL } from '../search/lib/requestGraphQl'

// Log events in batches.
const events = new Subject<Event>()

events
    .pipe(
        bufferTime(1000),
        concatMap(events => {
            if (events.length > 0) {
                return requestGraphQL<LogEventsResult, LogEventsVariables>(logEventsMutation, {
                    events,
                })
            }
            return EMPTY
        }),
        catchError(error => {
            console.error('Error logging events:', error)
            return []
        })
    )
    // eslint-disable-next-line rxjs/no-ignored-subscription
    .subscribe()

const logEventsMutation = gql`
    mutation LogEvents($events: [Event!]) {
        logEvents(events: $events) {
            alwaysNil
        }
    }
`

function logEvent(eventVariable: Event): void {
    events.next(eventVariable)
}

let eventId = 1

// Event Logger for the JetBrains Extension
export class EventLogger implements TelemetryService {
    private readonly anonymousUserId: string
    private listeners: Set<(eventName: string) => void> = new Set()
    private readonly editorInfo: { editor: string; version: string }

    constructor(anonymousUserId: string, editorInfo: { editor: string; version: string }) {
        this.anonymousUserId = anonymousUserId
        this.editorInfo = editorInfo
    }

    /**
     * @deprecated use logPageView instead
     */
    public logViewEvent(): void {
        throw new Error('This method is not supported on JetBrains. This extension does not use the deprecated method.')
    }

    public logPageView(): void {
        throw new Error(
            'This method is not supported on JetBrains. This extension does not have the concept of a page.'
        )
    }

    /**
     * Log a user action or event.
     * Event names should be specific and follow a ${noun}${verb} structure in pascal case, e.g. "ButtonClicked" or "SignInInitiated"
     *
     * @param eventName -
     * @param eventProperties event properties. These get logged to our database, but do not get
     * sent to our analytics systems. This may contain private info such as repository names or search queries.
     * @param publicArgument event properties that include only public information. Do NOT
     * include any private information, such as full URLs that may contain private repo names or
     * search queries. The contents of this parameter are sent to our analytics systems.
     * @param uri -
     */
    public log(
        eventName: string,
        eventProperties?: Record<string, unknown>,
        publicArgument?: Record<string, unknown>,
        uri?: string
    ): void {
        this.tracker(
            eventName,
            { ...eventProperties, ...this.editorInfo },
            { ...publicArgument, ...this.editorInfo },
            uri
        )
    }

    /**
     * Event ID is used to deduplicate events in Amplitude.
     * This is used in the case that multiple events with the same userID and timestamp
     * are sent. https://developers.amplitude.com/docs/http-api-v2#optional-keys
     */
    public getEventId(): number {
        return eventId++
    }

    public addEventLogListener(callback: (eventName: string) => void): () => void {
        this.listeners.add(callback)
        return () => this.listeners.delete(callback)
    }

    private tracker(eventName: string, eventProperties?: unknown, publicArgument?: unknown, uri?: string): void {
        for (const listener of this.listeners) {
            listener(eventName)
        }

        const event: Event = {
            event: eventName,
            userCookieID: this.anonymousUserId,
            referrer: 'JETBRAINS',
            url: uri || '',
            source: EventSource.IDEEXTENSION,
            argument: eventProperties ? JSON.stringify(eventProperties) : null,
            publicArgument: JSON.stringify(publicArgument),
            deviceID: this.anonymousUserId,
            eventID: this.getEventId(),
        }
        logEvent(event)
    }
}
