import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { RecipeID } from '@sourcegraph/cody-shared/src/chat/recipes/recipe'
import { ChatMessage, UserLocalHistory } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { Configuration } from '@sourcegraph/cody-shared/src/configuration'

import { View } from '../../webviews/NavBar'

/**
 * A message sent from the webview to the extension host.
 */
export type WebviewMessage =
    | { command: 'initialized' }
    | { command: 'event'; event: string; value: string }
    | { command: 'submit'; text: string; submitType: 'user' | 'suggestion' }
    | { command: 'executeRecipe'; recipe: RecipeID }
    | { command: 'settings'; serverEndpoint: string; accessToken: string }
    | { command: 'removeHistory' }
    | { command: 'restoreHistory'; chatID: string }
    | { command: 'deleteHistory'; chatID: string }
    | { command: 'links'; value: string }
    | { command: 'openFile'; filePath: string }
    | { command: 'edit'; text: string }
    | { command: 'insert'; text: string }
    | { command: 'auth'; type: 'signin' | 'signout' | 'support' | 'app' | 'app-poll' | 'callback'; endpoint?: string }
    | { command: 'abort' }

/**
 * A message sent from the extension host to the webview.
 */
export type ExtensionMessage =
    | { type: 'showTab'; tab: string }
    | { type: 'config'; config: ConfigurationSubsetForWebview & LocalEnv; authStatus: AuthStatus }
    | { type: 'login'; authStatus: AuthStatus }
    | { type: 'history'; messages: UserLocalHistory | null }
    | { type: 'transcript'; messages: ChatMessage[]; isMessageInProgress: boolean }
    | { type: 'debug'; message: string }
    | { type: 'contextStatus'; contextStatus: ChatContextStatus }
    | { type: 'view'; messages: View }
    | { type: 'errors'; errors: string }
    | { type: 'suggestions'; suggestions: string[] }
    | { type: 'app-state'; isInstalled: boolean }

/**
 * The subset of configuration that is visible to the webview.
 */
export interface ConfigurationSubsetForWebview extends Pick<Configuration, 'debugEnable' | 'serverEndpoint'> {}

/**
 * URLs for the Sourcegraph instance and app.
 */
export const DOTCOM_URL = new URL('https://sourcegraph.com')
export const DOTCOM_CALLBACK_URL = new URL('https://sourcegraph.com/user/settings/tokens/new/callback')
export const CODY_DOC_URL = new URL('https://docs.sourcegraph.com/cody')
// Community and support
export const DISCORD_URL = new URL('https://discord.gg/s2qDtYGnAE')
export const CODY_FEEDBACK_URL = new URL(
    'https://github.com/sourcegraph/sourcegraph/discussions/new?category=product-feedback&labels=cody,cody/vscode'
)
// APP
export const LOCAL_APP_URL = new URL('http://localhost:3080')
export const APP_LANDING_URL = new URL('https://about.sourcegraph.com/app')
export const APP_CALLBACK_URL = new URL('sourcegraph://user/settings/tokens/new/callback')
export const APP_OPEN_URL = new URL('sourcegraph://')
// TODO: Update URLs to always point to the latest app release: https://github.com/sourcegraph/sourcegraph/issues/53511
export const APP_DOWNLOAD_URLS: { [os: string]: { [arch: string]: string } } = {
    darwin: {
        arm64: 'https://github.com/sourcegraph/sourcegraph/releases/download/app-v2023.6.16%2B1314.6c2d49d47c/Cody_2023.6.16+1314.6c2d49d47c_aarch64.dmg',
        x64: 'https://github.com/sourcegraph/sourcegraph/releases/download/app-v2023.6.16%2B1314.6c2d49d47c/Cody_2023.6.16+1314.6c2d49d47c_x64.dmg',
    },
}

/**
 * The status of a users authentication, whether they're authenticated and have a
 * verified email.
 */
export interface AuthStatus {
    username?: string
    endpoint: string | null
    isLoggedIn: boolean
    showInvalidAccessTokenError: boolean
    authenticated: boolean
    hasVerifiedEmail: boolean
    requiresVerifiedEmail: boolean
    siteHasCodyEnabled: boolean
    siteVersion: string
}

export const defaultAuthStatus = {
    endpoint: '',
    isLoggedIn: false,
    showInvalidAccessTokenError: false,
    authenticated: false,
    hasVerifiedEmail: false,
    requiresVerifiedEmail: false,
    siteHasCodyEnabled: false,
    siteVersion: '',
}

export const unauthenticatedStatus = {
    isLoggedIn: false,
    showInvalidAccessTokenError: true,
    authenticated: false,
    hasVerifiedEmail: false,
    requiresVerifiedEmail: false,
    siteHasCodyEnabled: false,
    siteVersion: '',
}

/** The local environment of the editor. */
export interface LocalEnv {
    // The operating system kind
    os: string
    arch: string
    homeDir?: string | undefined

    // The URL scheme the editor is registered to in the operating system
    uriScheme: string
    // The application name of the editor
    appName: string
    extensionVersion: string

    // App Local State
    hasAppJson: boolean
    isAppInstalled: boolean
    isAppRunning: boolean
}

export function isLoggedIn(authStatus: AuthStatus): boolean {
    if (!authStatus.siteHasCodyEnabled) {
        return false
    }
    return authStatus.authenticated && (authStatus.requiresVerifiedEmail ? authStatus.hasVerifiedEmail : true)
}

export function isLocalApp(url: string): boolean {
    return new URL(url).origin === LOCAL_APP_URL.origin
}

export function isDotCom(url: string): boolean {
    return new URL(url).origin === DOTCOM_URL.origin
}
