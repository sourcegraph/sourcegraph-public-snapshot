import { take } from 'rxjs/operators'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { createSuggestionFetcher } from '../backend/search'
import { observeSourcegraphURL } from '../util/context'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'

const isURL = /^https?:\/\//

export class SearchCommand {
    public description = 'Enter a search query'

    private suggestionFetcher = createSuggestionFetcher(20, this.requestGraphQL)

    private prev: { query: string; suggestions: browser.omnibox.SuggestResult[] } = { query: '', suggestions: [] }

    constructor(private requestGraphQL: PlatformContext['requestGraphQL']) {}

    public getSuggestions = (query: string): Promise<browser.omnibox.SuggestResult[]> =>
        new Promise(resolve => {
            if (this.prev.query === query) {
                resolve(this.prev.suggestions)
                return
            }

            this.suggestionFetcher({
                query,
                handler: async suggestions => {
                    const sourcegraphURL = await observeSourcegraphURL(true) // isExtension=true, this feature is only supported in the browser extension
                        .pipe(take(1))
                        .toPromise()
                    const built = suggestions.map(({ title, url, urlLabel }) => ({
                        content: `${sourcegraphURL}${url}`,
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
        const sourcegraphURL = await observeSourcegraphURL(true) // isExtension=true, this feature is only supported in the browser extension
            .pipe(take(1))
            .toPromise()
        const props = {
            url: isURL.test(query)
                ? query
                : `${sourcegraphURL}/search?${buildSearchURLQuery(
                      query,
                      SearchPatternType.literal,
                      false
                  )}&utm_source=omnibox`,
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
