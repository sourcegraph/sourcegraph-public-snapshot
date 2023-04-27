import { CHARS_PER_TOKEN, MAX_AVAILABLE_PROMPT_LENGTH } from '../../prompt/constants'
import { truncateText } from '../../prompt/truncation'
import { Interaction } from '../transcript/interaction'

import { Recipe, RecipeContext } from './recipe'

export class NextQuestion implements Recipe {
    public id = 'next-question'

    public async getInteraction(humanChatInput: string, context: RecipeContext): Promise<Interaction | null> {
        const selection = context.editor.getActiveTextEditorSelectionOrEntireFile()
        if (!selection) {
            return Promise.resolve(null)
        }

        const promptPrefix = 'Assume I have an answer to the following request:'
        const promptSuffix =
            'Generate a follow-up question acting as the human in one sentence to uphold the conversation.'

        const maxTokenCount =
            MAX_AVAILABLE_PROMPT_LENGTH - (promptPrefix.length + promptSuffix.length) / CHARS_PER_TOKEN
        const truncatedText = truncateText(humanChatInput, maxTokenCount)
        const promptMessage = `${promptPrefix}\n\n\`\`\`\n${truncatedText}\n\`\`\`\n\n${promptSuffix}`

        const assistantResponsePrefix = 'Sure, here is great follow-up question:\n\n'
        return new Interaction(
            { speaker: 'human', text: promptMessage },
            {
                speaker: 'assistant',
                prefix: assistantResponsePrefix,
                text: assistantResponsePrefix,
            },
            Promise.resolve([])
        )
    }
}
