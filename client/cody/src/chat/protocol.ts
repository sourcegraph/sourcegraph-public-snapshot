import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { ChatMessage, UserLocalHistory } from '@sourcegraph/cody-shared/src/chat/transcript/messages'
import { Configuration } from '@sourcegraph/cody-shared/src/configuration'

import { View } from '../../webviews/NavBar'

/**
 * A message sent from the webview to the extension host.
 */
export type WebviewMessage =
    | {
          command: 'initialized'
      }
    | { command: 'event'; event: string; value: string }
    | { command: 'submit'; text: string }
    | { command: 'executeRecipe'; recipe: string }
    | { command: 'settings'; serverEndpoint: string; accessToken: string }
    | { command: 'removeToken' }
    | { command: 'removeHistory' }
    | { command: 'links'; value: string }
    | { command: 'openFile'; filePath: string }
    | { command: 'edit'; text: string }

/**
 * A message sent from the extension host to the webview.
 */
export type ExtensionMessage =
    | { type: 'showTab'; tab: string }
    | { type: 'config'; config: ConfigurationSubsetForWebview }
    | { type: 'login'; isValid: boolean }
    | { type: 'history'; messages: UserLocalHistory | null }
    | { type: 'transcript'; messages: ChatMessage[]; isMessageInProgress: boolean }
    | { type: 'debug'; message: string }
    | { type: 'contextStatus'; contextStatus: ChatContextStatus }
    | { type: 'view'; messages: View }

/**
 * The subset of configuration that is visible to the webview.
 */
export interface ConfigurationSubsetForWebview extends Pick<Configuration, 'debug' | 'serverEndpoint'> {
    hasAccessToken: boolean
}

export const DOTCOM_URL = new URL('https://sourcegraph.com')
