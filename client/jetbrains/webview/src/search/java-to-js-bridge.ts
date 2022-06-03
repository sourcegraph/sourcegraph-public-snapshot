import { PluginConfig, Theme } from './types'

import { applyConfig, applyTheme, renderReactApp } from './index'

export type ActionName = 'themeChanged' | 'pluginSettingsChanged'

type ThemeChangedRequestArguments = Theme
type PluginSettingsChangedRequestArguments = PluginConfig

type JavaToJSRequestArguments = ThemeChangedRequestArguments | PluginSettingsChangedRequestArguments

export function handleRequest(
    action: ActionName,
    argumentsAsJsonString: string,
    callback: (result: string) => void
): void {
    const argumentsAsObject = JSON.parse(argumentsAsJsonString) as JavaToJSRequestArguments
    if (action === 'themeChanged') {
        applyTheme(argumentsAsObject as ThemeChangedRequestArguments)
        renderReactApp()
        return callback(JSON.stringify(null))
    }

    if (action === 'pluginSettingsChanged') {
        applyConfig(argumentsAsObject as PluginSettingsChangedRequestArguments)
        renderReactApp()
        return callback(JSON.stringify(null))
    }

    return callback('Unknown action.')
}
