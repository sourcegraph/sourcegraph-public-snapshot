import { render } from 'react-dom'

import { ContentMatch } from '@sourcegraph/shared/src/search/stream'
import { setLinkComponent, AnchorLink } from '@sourcegraph/wildcard'

import { App } from './App'
import { callJava } from './mockJavaInterface'

setLinkComponent(AnchorLink)

export interface RequestToJava {
    action: string
    arguments: object
}

/* Add global functions to global window object */
declare global {
    interface Window {
        initializeSourcegraph: (isDarkTheme: boolean) => void
        callJava: (request: RequestToJava) => Promise<object>
    }
}

// @TODO: We need a mechanism for fetching the whole file for the follwoing two functions and make
// sure to cache the result (so that any given file is only ever downloaded once). A simple Map
// should do the trick.
//
// When we have this, these callbacks should just link through to our Java bridge.
function onOpen(match: ContentMatch, lineIndex: number): void {
    const line = match.lineMatches[lineIndex]
    console.log('open', match, line)
}
function onPreviewChange(match: ContentMatch, lineIndex: number): void {
    const line = match.lineMatches[lineIndex]
    console.log('preview', match, line)
}
function onPreviewClear(): void {
    // nothing
}

function renderReactApp(): void {
    const node = document.querySelector('#main') as HTMLDivElement
    render(<App onOpen={onOpen} onPreviewChange={onPreviewChange} onPreviewClear={onPreviewClear} />, node)
}

window.initializeSourcegraph = (isDarkTheme: boolean) => {
    window
        .callJava({ action: 'getTheme', arguments: {} })
        .then(response => {
            const root = document.querySelector(':root') as HTMLElement
            root.style.setProperty('--primary', (response as { buttonColor: string }).buttonColor)
            renderReactApp()
        })
        .catch((error: Error) => {
            console.error(`Failed to get theme: ${error.message}`)
            renderReactApp()
        })
    document.documentElement.classList.add('theme')
    document.documentElement.classList.add(isDarkTheme ? 'theme-dark' : 'theme-light')
}

/* Initialize app for standalone server */
if (window.location.search.includes('standalone=true')) {
    window.callJava = callJava
    window.initializeSourcegraph(true)
}
