import { render } from 'react-dom'

import { splitPath } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ContentMatch } from '@sourcegraph/shared/src/search/stream'
import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

import { App } from './App'
import { loadContent } from './lib/blob'
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

async function onOpen(match: ContentMatch, lineIndex: number): Promise<void> {
    const content = await loadContent(match)
    console.log('open', await loadContent(match), match.lineMatches[lineIndex])
    await window.callJava({
        action: 'open',
        arguments: {
            content,
            lineNumber: match.lineMatches[lineIndex].lineNumber,
        },
    })
}

async function onPreviewChange(match: ContentMatch, lineIndex: number): Promise<void> {
    const fileName = splitPath(match.path)[1]
    const content = await loadContent(match)
    console.log('preview', content, match.lineMatches[lineIndex])
    await window.callJava({
        action: 'preview',
        arguments: {
            fileName,
            path: match.path,
            content,
            lineNumber: match.lineMatches[lineIndex].lineNumber,
        },
    })
}

function onPreviewClear(): void {
    console.log('clear preview')
    window
        .callJava({ action: 'clearPreview', arguments: {} })
        .then(() => {})
        .catch(() => {})
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
