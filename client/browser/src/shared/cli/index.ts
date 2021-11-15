import { SearchCommand } from './search'

export function initializeOmniboxInterface(): void {
    const searchCommand = new SearchCommand()
    browser.omnibox.onInputChanged.addListener(async (query, suggest) => {
        try {
            const suggestions = await searchCommand.getSuggestions(query)
            suggest(suggestions)
        } catch (error) {
            console.error('error getting suggestions', error)
        }
    })

    browser.omnibox.onInputEntered.addListener(async (query, disposition) => {
        await searchCommand.action(query, disposition)
    })
}
