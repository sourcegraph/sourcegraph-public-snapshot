import { render } from 'react-dom'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import polyfillEventSource from '@sourcegraph/shared/src/polyfills/vendor/eventSource'
import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

import { getAuthenticatedUser } from '../sourcegraph-api-access/api-gateway'
import { EventLogger } from '../telemetry/EventLogger'

import { App } from './App'
import { handleRequest } from './java-to-js-bridge'
import {
    getConfigAlwaysFulfill,
    getThemeAlwaysFulfill,
    indicateFinishedLoading,
    loadLastSearchAlwaysFulfill,
    onOpen,
    onPreviewChange,
    onPreviewClear,
} from './js-to-java-bridge'
import type { PluginConfig, Search, Theme } from './types'

setLinkComponent(AnchorLink)

let isDarkTheme = false
let instanceURL = 'https://sourcegraph.com'
let isGlobbingEnabled = false
let accessToken: string | null = null
let anonymousUserId: string
let pluginVersion: string
let initialSearch: Search | null = null
let initialAuthenticatedUser: AuthenticatedUser | null
let telemetryService: EventLogger

window.initializeSourcegraph = async () => {
    const [theme, config, lastSearch, authenticatedUser] = await Promise.allSettled([
        getThemeAlwaysFulfill(),
        getConfigAlwaysFulfill(),
        loadLastSearchAlwaysFulfill(),
        getAuthenticatedUser(instanceURL, accessToken),
    ])

    applyConfig((config as PromiseFulfilledResult<PluginConfig>).value)
    applyTheme((theme as PromiseFulfilledResult<Theme>).value)
    applyLastSearch((lastSearch as PromiseFulfilledResult<Search | null>).value)
    applyAuthenticatedUser(authenticatedUser.status === 'fulfilled' ? authenticatedUser.value : null)
    if (accessToken && authenticatedUser.status === 'rejected') {
        console.warn(`No initial authenticated user with access token “${accessToken}”`)
    }

    telemetryService = new EventLogger(anonymousUserId, { editor: 'jetbrains', version: pluginVersion })

    renderReactApp()

    await indicateFinishedLoading()
}

window.callJS = handleRequest

export function renderReactApp(): void {
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
            initialAuthenticatedUser={initialAuthenticatedUser}
            telemetryService={telemetryService}
        />,
        node
    )
}

export function applyConfig(config: PluginConfig): void {
    instanceURL = config.instanceURL
    isGlobbingEnabled = config.isGlobbingEnabled || false
    accessToken = config.accessToken || null
    anonymousUserId = config.anonymousUserId || 'no-user-id'
    pluginVersion = config.pluginVersion
    polyfillEventSource(accessToken ? { Authorization: `token ${accessToken}` } : {})
}

export function applyTheme(theme: Theme): void {
    // Dark/light theme
    document.documentElement.classList.add('theme')
    document.documentElement.classList.remove(theme.isDarkTheme ? 'theme-light' : 'theme-dark')
    document.documentElement.classList.add(theme.isDarkTheme ? 'theme-dark' : 'theme-light')
    isDarkTheme = theme.isDarkTheme

    // Find the name of properties here: https://plugins.jetbrains.com/docs/intellij/themes-metadata.html#key-naming-scheme
    const intelliJTheme = theme.intelliJTheme
    const root = document.querySelector(':root') as HTMLElement

    root.style.setProperty('--button-color', intelliJTheme['Button.default.startBackground'])
    root.style.setProperty('--primary', intelliJTheme['Button.default.startBackground'])
    root.style.setProperty('--subtle-bg', intelliJTheme['ScrollPane.background'])

    root.style.setProperty('--dropdown-link-active-bg', intelliJTheme['List.selectionBackground'])
    root.style.setProperty('--light-text', intelliJTheme['List.selectionForeground'])

    root.style.setProperty('--jb-border-color', intelliJTheme['Component.borderColor'])
    root.style.setProperty('--jb-icon-color', intelliJTheme['Component.iconColor'] || '#7f8b91')

    // There is no color for this in the serialized theme, so I have picked this option from the
    // Dracula theme
    root.style.setProperty('--code-bg', theme.isDarkTheme ? '#2b2b2b' : '#ffffff')
    root.style.setProperty('--body-bg', theme.isDarkTheme ? '#2b2b2b' : '#ffffff')
}

function applyLastSearch(lastSearch: Search | null): void {
    initialSearch = lastSearch
}

function applyAuthenticatedUser(authenticatedUser: AuthenticatedUser | null): void {
    initialAuthenticatedUser = authenticatedUser
}

export function getAccessToken(): string | null {
    return accessToken
}

export function getInstanceURL(): string {
    return instanceURL
}
