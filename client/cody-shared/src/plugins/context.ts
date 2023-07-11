
import { ContextMessage } from '../codebase-context/messages'
import { Plugin } from './index'

export interface ContextSearchOptions {
    numResults: number
}

/**
 * Returns list of context messages for a given query, sorted in *reverse* order of importance (that is,
 * the most important context message appears *last*)
 */
export function getPluginContextMessages (query: string, options: ContextSearchOptions, plugins: Plugin[]): Promise<ContextMessage[]> {
    return Promise.all(plugins.map(plugin => plugin.search(query))
}
