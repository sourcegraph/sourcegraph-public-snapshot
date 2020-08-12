import { take, map, filter, catchError, timeout } from 'rxjs/operators'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { createSuggestionFetcher } from '../backend/search'
import { observeSourcegraphURL, getAssetsURL, DEFAULT_SOURCEGRAPH_URL } from '../util/context'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { createPlatformContext, BrowserPlatformContext } from '../platform/context'
import { from } from 'rxjs'
import { isDefined, isNot } from '../../../../shared/src/util/types'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { Settings } from '../../../../shared/src/settings/settings'

const isURL = /^https?:\/\//
const IS_EXTENSION = true // This feature is only supported in browser extension

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
                    const sourcegraphURL = await observeSourcegraphURL(IS_EXTENSION).pipe(take(1)).toPromise()
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
        const sourcegraphURL = await observeSourcegraphURL(IS_EXTENSION).pipe(take(1)).toPromise()

        const platformContext = createPlatformContext(
            { urlToFile: undefined, getContext: undefined },
            { sourcegraphURL, assetsURL: getAssetsURL(DEFAULT_SOURCEGRAPH_URL) },
            IS_EXTENSION
        )

        const patternType = await this.getDefaultSearchPatternType(platformContext)

        const props = {
            url: isURL.test(query)
                ? query
                : `${sourcegraphURL}/search?${buildSearchURLQuery(query, patternType, false)}&utm_source=omnibox`,
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

    private async getDefaultSearchPatternType(platformContext: BrowserPlatformContext): Promise<SearchPatternType> {
        try {
            await platformContext.refreshSettings()
            const settings = (await from(platformContext.settings).pipe(take(1)).toPromise()).final

            if (isDefined(settings) && isNot<ErrorLike | Settings, ErrorLike>(isErrorLike)(settings)) {
                return settings['search.defaultPatternType'] as SearchPatternType
            }
        } catch {
            // Ignore errors
        }

        return SearchPatternType.literal
    }
}
