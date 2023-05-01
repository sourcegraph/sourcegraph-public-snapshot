import { CodebaseContext } from '../../codebase-context'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { BotResponseMultiplexer } from '../bot-response-multiplexer'
import { Interaction } from '../transcript/interaction'

/** Tools and context recipes can use at the point they are invoked. */
export interface RecipeContext {
    editor: Editor
    intentDetector: IntentDetector
    codebaseContext: CodebaseContext
    responseMultiplexer: BotResponseMultiplexer
}

export interface Recipe {
    id: string
    getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null>
}
