import { render } from 'react-dom'

import { App } from './App'
import { callJava } from './mockJavaInterface'

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

function renderReactApp(): void {
    const node = document.querySelector('#main') as HTMLDivElement
    render(<App />, node)
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
