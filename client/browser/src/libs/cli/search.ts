import { buildSearchURLQuery } from '../../../../../shared/src/util/url'
import { storage } from '../../browser/storage'
import { createSuggestionFetcher } from '../../shared/backend/search'
import { sourcegraphUrl } from '../../shared/util/context'

const isURL = /^https?:\/\//

class SearchCommand {
    public description = 'Enter a search query'

    private suggestionFetcher = createSuggestionFetcher(20)

    private prev: { query: string; suggestions: browser.omnibox.SuggestResult[] } = { query: '', suggestions: [] }

    public getSuggestions = (query: string): Promise<browser.omnibox.SuggestResult[]> =>
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

    public action = async (query: string, disposition?: string): Promise<void> => {
        const { sourcegraphURL: url } = await storage.sync.get()
        const props = {
            url: isURL.test(query) ? query : `${url}/search?${buildSearchURLQuery(query)}&utm_source=omnibox`,
        }

        switch (disposition) {
            case 'newForegroundTab':
                await browser.tabs.create(props)
                break
            case 'newBackgroundTab':
                await browser.tabs.create({ ...props, active: false })
                break
            case 'currentTab':
            default:
                await browser.tabs.update(props)
                break
        }
    }
}

export default new SearchCommand()
