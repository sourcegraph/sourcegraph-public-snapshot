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
    document.documentElement.classList.add('theme')
    document.documentElement.classList.add(isDarkTheme ? 'theme-dark' : 'theme-light')
    renderReactApp()
}

/* Initialize app for standalone server */
if (window.location.search.includes('standalone=true')) {
    window.callJava = callJava
    window.initializeSourcegraph(true)
}
