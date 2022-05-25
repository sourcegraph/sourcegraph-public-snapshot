import { render } from 'react-dom'

import polyfillEventSource from '@sourcegraph/shared/src/polyfills/vendor/eventSource'
import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

import { App } from './App'
import {
    getConfig,
    getTheme,
    indicateFinishedLoading,
    loadLastSearch,
    onOpen,
    onPreviewChange,
    onPreviewClear,
} from './jsToJavaBridgeUtil'
import type { Theme, PluginConfig, Search } from './types'

setLinkComponent(AnchorLink)

let isDarkTheme = false
let instanceURL = 'https://sourcegraph.com'
let isGlobbingEnabled = false
let accessToken: string | null = null
let initialSearch: Search | null = null

window.initializeSourcegraph = async () => {
    const [theme, config, lastSearch] = await Promise.all([getTheme(), getConfig(), loadLastSearch()])

    applyConfig(config)
    applyTheme(theme)
    applyLastSearch(lastSearch)

    polyfillEventSource(accessToken ? { Authorization: `token ${accessToken}` } : {})

    renderReactApp()

    await indicateFinishedLoading()
}

function renderReactApp(): void {
    const node = document.querySelector('#main') as HTMLDivElement
    render(
        <App
            isDarkTheme={isDarkTheme}
            instanceURL={instanceURL}
            isGlobbingEnabled={isGlobbingEnabled}
            accessToken={accessToken}
            initialSearch={initialSearch}
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

function applyLastSearch(lastSearch: Search | null): void {
    initialSearch = lastSearch
}
