import { first } from 'lodash'
import * as OmniCLI from 'omnicli'

import storage from '../../extension/storage'
import * as tabs from '../../extension/tabs'

import { createSuggestionFetcher } from '../backend/search'
import { sourcegraphUrl } from '../util/context'
import { buildSearchURLQuery } from '../util/url'

const isURL = /^https?:\/\//

class SearchCommand implements OmniCLI.Command {
    public name = OmniCLI.DEFAULT_NAME
    public description = 'Enter a search query'

    private suggestionFetcher = createSuggestionFetcher(20)

    private prev: { query: string; suggestions: OmniCLI.Suggestion[] } = { query: '', suggestions: [] }

    public getSuggestions = (args: string[]): Promise<OmniCLI.Suggestion[]> => {
        const query = args.join(' ')

        return new Promise(resolve => {
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
    }

    public action = (args: string[], disposition?: string): void => {
        const query = args.join(' ')

        storage.getSync(({ serverUrls, sourcegraphURL }) => {
            const url = sourcegraphURL || first(serverUrls)
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
