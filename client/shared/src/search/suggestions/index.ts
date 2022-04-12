import { noop } from 'lodash'
import { fromEvent, Observable, of, Subscriber, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'

import { SearchPatternType, SearchVersion } from '../../graphql-operations'
import {
    AggregateStreamingSearchResults,
    messageHandlers,
    MessageHandlers,
    observeMessages,
    search,
    SearchEvent,
    SearchMatch,
    StreamSearchOptions,
    switchAggregateSearchResults,
} from '../stream'

const noopHandler = <T extends SearchEvent>(
    type: T['type'],
    eventSource: EventSource,
    _observer: Subscriber<SearchEvent>
): Subscription => fromEvent(eventSource, type).subscribe(noop)

const firstMatchMessageHandlers: MessageHandlers = {
    ...messageHandlers,
    matches: (type, eventSource, observer) =>
        observeMessages(type, eventSource).subscribe(data => {
            observer.next(data)
            // Once we observer the first `matches` event, complete the stream and close the event source.
            observer.complete()
            eventSource.close()
        }),
    progress: noopHandler,
    filters: noopHandler,
    alert: noopHandler,
}

/** Initiates a streaming search, stop at the first `matches` event, and aggregate the results. */
function firstMatchStreamingSearch(
    queryObservable: Observable<string>,
    options: StreamSearchOptions
): Observable<AggregateStreamingSearchResults> {
    return search(queryObservable, options, firstMatchMessageHandlers).pipe(switchAggregateSearchResults)
}

export function fetchStreamSuggestions(query: string, sourcegraphURL?: string): Observable<SearchMatch[]> {
    return firstMatchStreamingSearch(of(query), {
        version: SearchVersion.V2,
        patternType: SearchPatternType.literal,
        caseSensitive: false,
        trace: undefined,
        sourcegraphURL,
    }).pipe(map(suggestions => suggestions.results))
}
