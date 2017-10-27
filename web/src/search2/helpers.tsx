
import * as H from 'history'
import { events } from '../tracking/events'
import { buildSearchURLQuery, SearchOptions } from './index'

export function submitSearch(history: H.History, options: SearchOptions): void {
    // Go to search results page
    const path = '/search?' + buildSearchURLQuery(options)
    events.SearchSubmitted.log({
        code_search: {
            pattern: options.scopeQuery ? `${options.scopeQuery} ${options.query}` : options.query,
            query: options.query,
            scopeQuery: options.scopeQuery,
        },
    })
    history.push(path)
}
