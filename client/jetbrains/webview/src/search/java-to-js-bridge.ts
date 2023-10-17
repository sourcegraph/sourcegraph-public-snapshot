import {
    applyConfig,
    applyTheme,
    getAuthenticatedUser,
    renderReactApp,
    retrySearch,
    updateVersionAndAuthDataFromServer,
    wasServerAccessSuccessful,
} from './index'
import { indicateFinishedLoading } from './js-to-java-bridge'
import type { PluginConfig, Theme } from './types'

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
        const pluginConfig = argumentsAsObject as PluginSettingsChangedRequestArguments
        applyConfig(pluginConfig)
        await updateVersionAndAuthDataFromServer()
        await indicateFinishedLoading(wasServerAccessSuccessful() || false, !!getAuthenticatedUser())
        renderReactApp()
        return callback(JSON.stringify(null))
    }

    if (action === 'retrySearch') {
        if (!wasServerAccessSuccessful()) {
            await updateVersionAndAuthDataFromServer()
        }
        await indicateFinishedLoading(wasServerAccessSuccessful() || false, !!getAuthenticatedUser())
        retrySearch()
        renderReactApp()
        return callback(JSON.stringify(null))
    }

    return callback('Unknown action.')
}
