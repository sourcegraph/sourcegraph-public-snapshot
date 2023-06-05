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
    | { command: 'removeToken' }
    | { command: 'removeHistory' }
    | { command: 'restoreHistory'; chatID: string }
    | { command: 'links'; value: string }
    | { command: 'openFile'; filePath: string }
    | { command: 'edit'; text: string }
    | { command: 'insert'; text: string }

/**
 * A message sent from the extension host to the webview.
 */
export type ExtensionMessage =
    | { type: 'showTab'; tab: string }
    | { type: 'config'; config: ConfigurationSubsetForWebview; authStatus: AuthStatus }
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

export const DOTCOM_URL = new URL('https://sourcegraph.com')
export const LOCAL_APP_URL = new URL('http://localhost:3080')

/**
 * The status of a users authentication, whether they're authenticated and have a
 * verified email.
 */
export interface AuthStatus {
    showInvalidAccessTokenError: boolean
    authenticated: boolean
    hasVerifiedEmail: boolean
    requiresVerifiedEmail: boolean
    siteHasCodyEnabled: boolean
    siteVersion: string
}

export const defaultAuthStatus = {
    showInvalidAccessTokenError: false,
    authenticated: false,
    hasVerifiedEmail: false,
    requiresVerifiedEmail: false,
    siteHasCodyEnabled: false,
    siteVersion: '',
}

export const unauthenticatedStatus = {
    showInvalidAccessTokenError: true,
    authenticated: false,
    hasVerifiedEmail: false,
    requiresVerifiedEmail: false,
    siteHasCodyEnabled: false,
    siteVersion: '',
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
