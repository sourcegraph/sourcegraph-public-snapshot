import { CodebaseContext } from '../../codebase-context'
import { ContextMessage } from '../../codebase-context/messages'
import { Editor } from '../../editor'
import { IntentDetector } from '../../intent-detector'
import { MAX_HUMAN_INPUT_TOKENS } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { getShortTimestamp } from '../../timestamp'
import { renderMarkdown } from '../markdown'
import { Interaction } from '../transcript/interaction'

import { Recipe } from './recipe'

export class ChatQuestion implements Recipe {
    public getID(): string {
        return 'chat-question'
    }

    public async getInteraction(
        humanChatInput: string,
        _editor: Editor,
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<Interaction | null> {
        const timestamp = getShortTimestamp()
        const truncatedText = truncateText(humanChatInput, MAX_HUMAN_INPUT_TOKENS)
        const displayText = renderMarkdown(humanChatInput)

        // TODO: Include current file as context.
        return new Interaction(
            { speaker: 'human', text: truncatedText, displayText, timestamp },
            { speaker: 'assistant', text: '', displayText: '', timestamp },
            this.getContextMessages(truncatedText, intentDetector, codebaseContext)
        )
    }

    private async getContextMessages(
        text: string,
        intentDetector: IntentDetector,
        codebaseContext: CodebaseContext
    ): Promise<ContextMessage[]> {
        const isCodebaseContextRequired = await intentDetector.isCodebaseContextRequired(text)
        if (!isCodebaseContextRequired) {
            return []
        }
        return codebaseContext.getContextMessages(text, { numCodeResults: 8, numTextResults: 2 })
    }
}
