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
    firstInteraction: boolean
}

export type RecipeID =
    | 'chat-question'
    | 'explain-code-detailed'
    | 'explain-code-high-level'
    | 'generate-unit-test'
    | 'generate-docstring'
    | 'improve-variable-names'
    | 'translate-to-language'
    | 'git-history'
    | 'find-code-smells'
    | 'fixup'
    | 'context-search'
    | 'release-notes'
    | 'inline-chat'
    | 'next-questions'
    | 'optimize-code'

export interface Recipe {
    id: RecipeID
    getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null>
}
