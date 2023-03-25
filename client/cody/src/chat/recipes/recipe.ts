import { Interaction } from '@sourcegraph/cody-shared/src/chat/transcript/interaction'
import { IntentDetector } from '@sourcegraph/cody-shared/src/intent-detector'

import { CodebaseContext } from '../../codebase-context'
import { Editor } from '../../editor'

export interface Recipe {
    getID(): string
    getInteraction(
        humanChatInput: string,
        editor: Editor,
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<Interaction | null>
}
