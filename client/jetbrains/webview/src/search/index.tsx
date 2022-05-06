import { render } from 'react-dom'

import { ContentMatch } from '@sourcegraph/shared/src/search/stream'
import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

import { App } from './App'
import { createRequestForMatch, RequestToJava } from './jsToJavaBridgeUtil'
import { callJava } from './mockJavaInterface'

setLinkComponent(AnchorLink)

/* Add global functions to global window object */
declare global {
    interface Window {
        initializeSourcegraph: (isDarkTheme: boolean) => void
        callJava: (request: RequestToJava) => Promise<object>
    }
}

async function onPreviewChange(match: ContentMatch, lineMatchIndex: number): Promise<void> {
    await window.callJava(await createRequestForMatch(match, lineMatchIndex, 'preview'))
}

function onPreviewClear(): void {
    window
        .callJava({ action: 'clearPreview', arguments: {} })
        .then(() => {})
        .catch(() => {})
}

async function onOpen(match: ContentMatch, lineMatchIndex: number): Promise<void> {
    await window.callJava(await createRequestForMatch(match, lineMatchIndex, 'open'))
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
            const buttonColor = (response as { buttonColor: string }).buttonColor
            if (buttonColor) {
                root.style.setProperty('--button-color', buttonColor)
            }
            root.style.setProperty('--primary', buttonColor)
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
