import * as H from 'history'
import { eventLogger } from '../tracking/eventLogger'
import { buildSearchURLQuery, SearchOptions } from './index'

export function submitSearch(history: H.History, options: SearchOptions): void {
    // Go to search results page
    const path = '/search?' + buildSearchURLQuery(options)
    eventLogger.log('SearchSubmitted', {
        code_search: {
            pattern: options.scopeQuery ? `${options.scopeQuery} ${options.query}` : options.query,
            query: options.query,
            scopeQuery: options.scopeQuery,
        },
    })
    history.push(path, { ...history.location.state, ...options })
}
