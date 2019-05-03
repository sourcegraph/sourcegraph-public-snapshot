import searchCommand from './search'

export function initializeOmniboxInterface(): void {
    browser.omnibox.onInputChanged.addListener(async (query, suggest) => {
        try {
            const suggestions = await searchCommand.getSuggestions(query)
            suggest(suggestions)
        } catch (err) {
            console.error('error getting suggestions', err)
        }
    })

    browser.omnibox.onInputEntered.addListener(async (query, disposition) => {
        await searchCommand.action(query, disposition)
    })
}
