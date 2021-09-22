import { from } from 'rxjs'
import { take } from 'rxjs/operators'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined, isNot } from '@sourcegraph/shared/src/util/types'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { SearchPatternType } from '../../graphql-operations'
import { createSuggestionFetcher } from '../backend/search'
import { createPlatformContext } from '../platform/context'
import { observeSourcegraphURL, getAssetsURL, DEFAULT_SOURCEGRAPH_URL } from '../util/context'

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

    public action = async (
        query: string,
        disposition?: 'newForegroundTab' | 'newBackgroundTab' | 'currentTab'
    ): Promise<void> => {
        const sourcegraphURL = await observeSourcegraphURL(IS_EXTENSION).pipe(take(1)).toPromise()

        const [patternType, caseSensitive] = await this.getDefaultSearchSettings(sourcegraphURL)

        const props = {
            url: isURL.test(query)
                ? query
                : `${sourcegraphURL}/search?${buildSearchURLQuery(
                      query,
                      patternType,
                      caseSensitive
                  )}&utm_source=omnibox`,
        }

        if (disposition === 'newForegroundTab') {
            await browser.tabs.create(props)
            return
        }
        if (disposition === 'newBackgroundTab') {
            await browser.tabs.create({ ...props, active: false })
            return
        }

        const [currentTab] = await browser.tabs.query({ active: true, currentWindow: true })
        if (!currentTab.id) {
            await browser.tabs.update(props)
            return
        }

        // Note: this is done in order to blur browser omnibox and set focus on page
        await Promise.all([browser.tabs.create(props), browser.tabs.remove(currentTab.id)])
    }

    private lastSourcegraphUrl = ''
    private settingsTimeoutHandler = 0
    private defaultPatternType = SearchPatternType.literal
    private defaultCaseSensitive = false
    private readonly settingsTimeoutDuration = 60 * 60 * 1000 // one hour

    // Returns the pattern type to use, and whether the query should be treated case sensitively.
    private async getDefaultSearchSettings(sourcegraphURL: string): Promise<[SearchPatternType, boolean]> {
        try {
            // Refresh settings when either:
            // - First search
            // - Sourcegraph URL changes
            // - Over an hour has passed since last refresh

            if (this.lastSourcegraphUrl !== sourcegraphURL || this.settingsTimeoutHandler === 0) {
                clearTimeout(this.settingsTimeoutHandler)
                this.settingsTimeoutHandler = 0

                const platformContext = createPlatformContext(
                    { urlToFile: undefined },
                    { sourcegraphURL, assetsURL: getAssetsURL(DEFAULT_SOURCEGRAPH_URL) },
                    IS_EXTENSION
                )

                await platformContext.refreshSettings()
                const settings = (await from(platformContext.settings).pipe(take(1)).toPromise()).final

                if (isDefined(settings) && isNot<ErrorLike | Settings, ErrorLike>(isErrorLike)(settings)) {
                    this.defaultPatternType =
                        (settings['search.defaultPatternType'] as SearchPatternType) || this.defaultPatternType
                    this.defaultCaseSensitive =
                        (settings['search.defaultCaseSensitive'] as boolean) || this.defaultCaseSensitive
                }

                this.lastSourcegraphUrl = sourcegraphURL
                this.settingsTimeoutHandler = window.setTimeout(() => {
                    this.settingsTimeoutHandler = 0
                }, this.settingsTimeoutDuration)
            }
        } catch {
            // Ignore errors trying to get settings, fall to return default below
        }

        return [this.defaultPatternType, this.defaultCaseSensitive]
    }
}
