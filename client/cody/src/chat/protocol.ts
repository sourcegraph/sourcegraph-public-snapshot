import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { ChatMessage, UserLocalHistory } from '@sourcegraph/cody-shared/src/chat/transcript/messages'

/**
 * A message sent from the webview to the extension host.
 */
export type WebviewMessage =
    | {
          command: 'initialized'
      }
    | { command: 'reset' }
    | { command: 'submit'; text: string }
    | { command: 'executeRecipe'; recipe: string }
    | { command: 'acceptTOS'; version: string }
    | { command: 'settings'; serverEndpoint: string; accessToken: string }
    | { command: 'removeToken' }
    | { command: 'removeHistory' }
    | { command: 'links'; value: string }
    | { command: 'openFile'; filePath: string }

/**
 * A message sent from the extension host to the webview.
 */
export type ExtensionMessage =
    | { type: 'showTab'; tab: string }
    | { type: 'login'; isValid: boolean }
    | { type: 'token'; value: string; mode: 'development' | 'production' }
    | { type: 'history'; messages: UserLocalHistory | null }
    | { type: 'transcript'; messages: ChatMessage[]; isMessageInProgress: boolean }
    | { type: 'debug'; message: string }
    | { type: 'contextStatus'; contextStatus: ChatContextStatus }
