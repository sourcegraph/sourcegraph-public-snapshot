import storage from '../../browser/storage'
import * as tabs from '../../browser/tabs'

import { createSuggestionFetcher } from '../../shared/backend/search'
import { sourcegraphUrl } from '../../shared/util/context'
import { buildSearchURLQuery } from '../../shared/util/url'

const isURL = /^https?:\/\//

class SearchCommand {
    public description = 'Enter a search query'

    private suggestionFetcher = createSuggestionFetcher(20)

    private prev: { query: string; suggestions: chrome.omnibox.SuggestResult[] } = { query: '', suggestions: [] }

    public getSuggestions = (query: string): Promise<chrome.omnibox.SuggestResult[]> =>
        new Promise(resolve => {
            if (this.prev.query === query) {
                resolve(this.prev.suggestions)
                return
            }

            this.suggestionFetcher({
                query,
                handler: suggestions => {
                    const built = suggestions.map(({ title, url, urlLabel }) => ({
                        content: `${sourcegraphUrl}${url}`,
                        description: `${title} - ${urlLabel}`,
                    }))

                    this.prev = {
                        query,
                        suggestions: built,
                    }

                    resolve(built)
                },
            })
        })

    public action = (query: string, disposition?: string): void => {
        storage.getSync(({ sourcegraphURL: url }) => {
            const props = {
                url: isURL.test(query) ? query : `${url}/search?${buildSearchURLQuery(query)}&utm_source=omnibox`,
            }

            switch (disposition) {
                case 'newForegroundTab':
                    tabs.create(props)
                    break
                case 'newBackgroundTab':
                    tabs.create({ ...props, active: false })
                    break
                case 'currentTab':
                default:
                    tabs.update(props)
                    break
            }
        })
    }
}

export default new SearchCommand()
