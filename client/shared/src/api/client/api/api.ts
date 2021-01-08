import { ClientCodeEditorAPI } from './codeEditor'
import { ClientContentAPI } from './content'
import { ClientContextAPI } from './context'
import { ClientLanguageFeaturesAPI } from './languageFeatures'
import { ClientViewsAPI } from './views'
import { MainThreadAPI } from '../../contract'

/**
 * The API that is exposed from the client (main thread) to the extension host (worker)
 */
export interface ClientAPI extends MainThreadAPI {
    ping(): 'pong'

    context: ClientContextAPI
    languageFeatures: ClientLanguageFeaturesAPI
    codeEditor: ClientCodeEditorAPI
    views: ClientViewsAPI
    content: ClientContentAPI
}
