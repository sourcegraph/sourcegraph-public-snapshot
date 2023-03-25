import { CodebaseContext } from '../../codebase-context'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { Interaction } from '../transcript/interaction'

export interface Recipe {
    getID(): string
    getInteraction(
        humanChatInput: string,
        editor: Editor,
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<Interaction | null>
}
