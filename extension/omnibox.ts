const chrome = global.chrome

export const setDefaultSuggestion = (suggestion: chrome.omnibox.Suggestion) => {
    if (chrome && chrome.omnibox) {
        chrome.omnibox.setDefaultSuggestion(suggestion)
    }
}

export const onInputChanged = (
    handler: (text: string, suggest: (suggestResults: chrome.omnibox.SuggestResult[]) => void) => void
) => {
    if (chrome && chrome.omnibox) {
        chrome.omnibox.onInputChanged.addListener(handler)
    }
}

export const onInputEntered = (handler: (url: string, disposition: string) => void) => {
    if (chrome && chrome.omnibox) {
        chrome.omnibox.onInputEntered.addListener(handler)
    }
}
