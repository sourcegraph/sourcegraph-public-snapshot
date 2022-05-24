import { render } from 'react-dom'

import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

import { App } from './App'
import {
    getConfig,
    getTheme,
    indicateFinishedLoading,
    onOpen,
    onPreviewChange,
    onPreviewClear,
    PluginConfig,
    Request,
    Theme,
} from './jsToJavaBridgeUtil'

setLinkComponent(AnchorLink)

let isDarkTheme = false
let instanceURL = 'https://sourcegraph.com'
let isGlobbingEnabled = false
let accessToken: string | null = null

/* Add global functions to global window object */
declare global {
    interface Window {
        initializeSourcegraph: () => void
        callJava: (request: Request) => Promise<object>
    }
}

function renderReactApp(): void {
    const node = document.querySelector('#main') as HTMLDivElement
    render(
        <App
            isDarkTheme={isDarkTheme}
            instanceURL={instanceURL}
            isGlobbingEnabled={isGlobbingEnabled}
            accessToken={accessToken}
            onOpen={onOpen}
            onPreviewChange={onPreviewChange}
            onPreviewClear={onPreviewClear}
        />,
        node
    )
}

function applyConfig(config: PluginConfig): void {
    instanceURL = config.instanceURL
    isGlobbingEnabled = config.isGlobbingEnabled || false
    accessToken = config.accessToken || null
}

function applyTheme(theme: Theme): void {
    // Dark/light theme
    document.documentElement.classList.add('theme')
    document.documentElement.classList.remove(theme.isDarkTheme ? 'theme-light' : 'theme-dark')
    document.documentElement.classList.add(theme.isDarkTheme ? 'theme-dark' : 'theme-light')
    isDarkTheme = theme.isDarkTheme

    // Button color (test)
    const buttonColor = theme.buttonColor
    const root = document.querySelector(':root') as HTMLElement
    if (buttonColor) {
        root.style.setProperty('--button-color', buttonColor)
    }
    root.style.setProperty('--primary', buttonColor)
}

window.initializeSourcegraph = async () => {
    const [theme, config] = await Promise.all([getTheme(), getConfig()])
    applyConfig(config)
    applyTheme(theme)
    renderReactApp()
    await indicateFinishedLoading()
}
