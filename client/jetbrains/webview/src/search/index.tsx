import { render } from 'react-dom'

import { App } from './App'

interface RequestToJava {
    action: string,
    arguments: object,
}

interface ResponseFromJava {
    success: boolean
    errorMessage: string | null
    data: object
}

/* Add global functions to global window object */
declare global {
    interface Window {
        initializeSourcegraph: (isDarkTheme: boolean) => void,
        javaBridge: (request: RequestToJava) => Promise<ResponseFromJava>
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
    window.initializeSourcegraph(true)
}
