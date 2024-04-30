import { EMPTY, type Observable, Subject, lastValueFrom } from 'rxjs'
import { bufferTime, catchError, concatMap } from 'rxjs/operators'

import { gql, dataOrThrowErrors, requestGraphQLCommon, type GraphQLResult } from '@sourcegraph/http-client'

import type { LogEventsResult, LogEventsVariables, Event } from '../../graphql-operations'

const getHeaders = (): { [header: string]: string } => {
    const headers: { [header: string]: string } = {
        ...window.context?.xhrHeaders,
        Accept: 'application/json',
        'Content-Type': 'application/json',
    }
    const parameters = new URLSearchParams(location.search)
    const trace = parameters.get('trace')
    if (trace) {
        headers['X-Sourcegraph-Should-Trace'] = trace
    }

    // Get values from URL and local overrides
    const feat = parameters.getAll('feat')
    if (feat.length) {
        headers['X-Sourcegraph-Override-Feature'] = feat.join(',')
    }
    return headers
}

const requestGraphQL = <TResult, TVariables = object>(
    request: string,
    variables?: TVariables
): Observable<GraphQLResult<TResult>> =>
    requestGraphQLCommon({
        request,
        variables,
        headers: getHeaders(),
    })

// Log events in batches.
const BATCHED_EVENTS = new Subject<Event>()

export const logEventsMutation = gql`
    mutation LogEvents($events: [Event!]) {
        logEvents(events: $events) {
            alwaysNil
        }
    }
`

function sendEvents(events: Event[]): Promise<void> {
    return lastValueFrom(
        requestGraphQL<LogEventsResult, LogEventsVariables>(logEventsMutation, {
            events,
        })
    )
        .then(dataOrThrowErrors)
        .then(() => {})
}

BATCHED_EVENTS.pipe(
    bufferTime(1000),
    concatMap(events => {
        if (events.length > 0) {
            return sendEvents(events)
        }
        return EMPTY
    }),
    // TODO: log errors to Sentry
    catchError(() => [])
)
    // eslint-disable-next-line rxjs/no-ignored-subscription
    .subscribe()

/**
 * Log a raw user action (used to allow site admins on a Sourcegraph instance
 * to see a count of unique users on a daily, weekly, and monthly basis).
 *
 * When invoked on a non-Sourcegraph.com instance, this data is stored in the
 * instance's database, and not sent to Sourcegraph.com.
 *
 * @deprecated Use a TelemetryRecorder or TelemetryRecorderProvider from
 * src/telemetry instead.
 */
export function logEvent(event: Event): void {
    BATCHED_EVENTS.next(event)
}
