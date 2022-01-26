import { SearchCommand } from './search'

export function initializeOmniboxInterface(): void {
    const searchCommand = new SearchCommand()
    // Note: This is needed because getting current active tab is asynchronous, and in a case when you switch tabs quickly, we mis-use wrong tab.
    let currentTabId: number | undefined
    let isOnEnterRunning = false

    browser.tabs.onActivated.addListener(activeInfo => {
        if (!isOnEnterRunning) {
            currentTabId = activeInfo.tabId
        }
    })

    browser.omnibox.onInputChanged.addListener((query, suggest) => {
        searchCommand
            .getSuggestions(query)
            .then(suggest)
            .catch(error => console.error('error getting suggestions', error))
    })

    browser.omnibox.onInputEntered.addListener((query, disposition) => {
        isOnEnterRunning = true
        searchCommand
            .action(query, disposition, currentTabId)
            .catch(error => console.error('error processing search query', error))
            .finally(() => (isOnEnterRunning = false))
    })
}
