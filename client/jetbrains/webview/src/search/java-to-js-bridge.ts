import { getAuthenticatedUser } from '../sourcegraph-api-access/api-gateway'

import { indicateFinishedLoading } from './js-to-java-bridge'
import { PluginConfig, Theme } from './types'

import { applyConfig, applyTheme, renderReactApp } from './index'

export type ActionName = 'themeChanged' | 'pluginSettingsChanged'

type ThemeChangedRequestArguments = Theme
type PluginSettingsChangedRequestArguments = PluginConfig

type JavaToJSRequestArguments = ThemeChangedRequestArguments | PluginSettingsChangedRequestArguments

export async function handleRequest(
    action: ActionName,
    argumentsAsJsonString: string,
    callback: (result: string) => void
): Promise<void> {
    const argumentsAsObject = JSON.parse(argumentsAsJsonString) as JavaToJSRequestArguments
    if (action === 'themeChanged') {
        applyTheme(argumentsAsObject as ThemeChangedRequestArguments)
        renderReactApp()
        return callback(JSON.stringify(null))
    }

    if (action === 'pluginSettingsChanged') {
        applyConfig(argumentsAsObject as PluginSettingsChangedRequestArguments)
        try {
            const authenticatedUser = await getAuthenticatedUser(
                (argumentsAsObject as PluginSettingsChangedRequestArguments).instanceURL,
                (argumentsAsObject as PluginSettingsChangedRequestArguments).accessToken
            )
            await indicateFinishedLoading(!!authenticatedUser)
        } catch {
            await indicateFinishedLoading(false)
        }
        renderReactApp()
        return callback(JSON.stringify(null))
    }

    return callback('Unknown action.')
}
